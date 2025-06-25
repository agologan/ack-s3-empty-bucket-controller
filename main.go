package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	apiGroup         = "s3.services.k8s.aws"
	apiVersion       = "v1alpha1"
	apiResource      = "buckets"
	targetAnnotation = "s3.services.k8s.aws/empty-on-delete"
)

func getKubeConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig, found := os.LookupEnv("KUBECONFIG")
		if !found {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		log.Println("Using kubeconfig file: ", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return config, err
}

func emptyBucket(ctx context.Context, s3client *s3.Client, bucket string) error {
	log.Printf("Emptying bucket: %s", bucket)
	paginator := s3.NewListObjectsV2Paginator(s3client, &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}
		if len(page.Contents) == 0 {
			continue
		}
		var objects []s3types.ObjectIdentifier
		for _, obj := range page.Contents {
			objects = append(objects, s3types.ObjectIdentifier{Key: obj.Key})
		}
		_, err = s3client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &s3types.Delete{
				Objects: objects,
			},
		})
		if err != nil {
			return err
		}
	}
	log.Printf("Bucket %s emptied.", bucket)
	return nil
}

type BucketUpdater struct {
	s3client *s3.Client
}

func (bu *BucketUpdater) onBucketUpdate(_, newObj interface{}) {
	obj := newObj.(*unstructured.Unstructured)
	meta := obj.Object["metadata"].(map[string]interface{})
	annotations, _ := meta["annotations"].(map[string]interface{})
	deletionTimestamp, _ := meta["deletionTimestamp"].(string)
	if deletionTimestamp != "" && annotations != nil {
		if val, ok := annotations[targetAnnotation]; ok && val == "true" {
			bucketName := obj.Object["spec"].(map[string]interface{})["name"].(string)
			if bucketName == "" {
				bucketName = obj.GetName()
			}
			if err := emptyBucket(context.TODO(), bu.s3client, bucketName); err != nil {
				log.Printf("Error emptying bucket %s: %v", bucketName, err)
				return
			}
		}
	}
}

func main() {
	log.Println("Starting S3 ACK empty bucket controller...")
	cfg, err := getKubeConfig()
	if err != nil {
		log.Fatalf("Failed to get kubeconfig: %v", err)
	}
	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create dynamic client: %v", err)
	}
	gvr := schema.GroupVersionResource{Group: apiGroup, Version: apiVersion, Resource: apiResource}

	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	s3client := s3.NewFromConfig(awsCfg)

	updater := &BucketUpdater{s3client: s3client}

	factory := dynamicinformer.NewDynamicSharedInformerFactory(dynClient, 30*time.Second)
	informer := factory.ForResource(gvr).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: updater.onBucketUpdate,
	})
	stop := make(chan struct{})
	informer.Run(stop)
}
