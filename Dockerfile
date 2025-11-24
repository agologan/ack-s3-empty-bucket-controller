FROM golang:1.25 AS builder
WORKDIR /workspace

COPY . .

ARG GOARCH
ARG LDFLAGS_FLAG
ARG TAGS_FLAG

RUN CGO_ENABLED=0 GOOS=linux go build -o empty-bucket-controller-$GOARCH $LDFLAGS_FLAG $TAGS_FLAG

FROM gcr.io/distroless/static:nonroot
ARG GOARCH
COPY --from=builder /workspace/empty-bucket-controller-$GOARCH /empty-bucket-controller

WORKDIR /
CMD ["/empty-bucket-controller"]
