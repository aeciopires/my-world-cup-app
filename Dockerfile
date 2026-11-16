# syntax=docker/dockerfile:1

# Keep the Go minor version in sync with mise.toml and app/go.mod.
# The builder always runs on the build host's native platform and
# cross-compiles for the target (GOOS/GOARCH below), which is much faster
# than emulating arm64 under QEMU for multi-arch builds.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /src

# Never let `go` download a different toolchain inside the build.
ENV GOTOOLCHAIN=local CGO_ENABLED=0

# Copy only the module files first so this layer (the dependency download)
# is cached until go.mod/go.sum actually change.
COPY app/go.mod app/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY app/ .

# TARGETOS/TARGETARCH are set by BuildKit/Buildx for each requested
# --platform, so the same Dockerfile cross-compiles amd64 and arm64 images.
ARG TARGETOS=linux
ARG TARGETARCH
ARG APP_VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X main.version=${APP_VERSION}" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian13:nonroot AS runtime

# Set at build time from the repo's VERSION file (see Makefile's APP_VERSION /
# --build-arg APP_VERSION) — not read from the file directly, since a
# Dockerfile has no shell in the runtime stage to read it at container
# startup. Defaults to "dev" for a plain `docker build .` with no build-arg.
ARG APP_VERSION=dev
LABEL org.opencontainers.image.title="my-world-cup-app" \
      org.opencontainers.image.version="$APP_VERSION" \
      org.opencontainers.image.source="https://github.com/aeciopires/my-world-cup-app" \
      org.opencontainers.image.licenses="GPL-3.0"

WORKDIR /app
COPY --from=builder /out/server /app/server

ENV PORT=8080
EXPOSE 8080

USER 65532:65532

# The runtime image has no shell/curl, so the binary checks itself.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "-healthcheck"]

ENTRYPOINT ["/app/server"]
