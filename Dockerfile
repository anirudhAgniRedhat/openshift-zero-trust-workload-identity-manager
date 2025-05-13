# Build the Zero Trust Workload Identity Manager binary
FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-1.23-openshift-4.19 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Add replace directive to go.mod (if needed)
# RUN echo 'replace github.com/openshift/zero-trust-workload-identity-manager => ./' >> go.mod

# Cache deps
RUN go mod download

# Copy the entire project
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -mod=mod -a -o zero-trust-workload-identity-manager ./cmd/zero-trust-workload-identity-manager/main.go

# Use distroless as minimal base image to package the manager binary
FROM registry.access.redhat.com/ubi9-minimal:9.4
WORKDIR /
COPY --from=builder /workspace/zero-trust-workload-identity-manager .
USER 65532:65532

ENTRYPOINT ["/zero-trust-workload-identity-manager"]
