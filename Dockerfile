# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY app/go.mod ./
RUN go mod download

COPY app/ .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

# Set at build time from the repo's VERSION file (see Makefile's APP_VERSION /
# --build-arg APP_VERSION) — not read from the file directly, since a
# Dockerfile has no shell in the runtime stage to read it at container
# startup. Defaults to "dev" for a plain `docker build .` with no build-arg.
ARG APP_VERSION=dev
LABEL org.opencontainers.image.version="$APP_VERSION" \
      org.opencontainers.image.source="https://github.com/aeciopires/my-world-cup-app"

WORKDIR /app
COPY --from=builder /out/server /app/server

ENV PORT=8080
EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/app/server"]
