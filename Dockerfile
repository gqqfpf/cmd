# Build the manager binary
FROM golang:1.19-bullseye as builder

USER root
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY pkg/ pkg/
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn"
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod tidy

# Copy the go source
# Build
RUN  go build -a -o backup main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM 192.168.251.78/edoc2v5/static:nonroot
WORKDIR /
COPY --from=builder /workspace/backup .
USER 65532:65532

ENTRYPOINT ["/backup"]
