# syntax=docker/dockerfile:1

# The frontend does not depend on the target architecture: build it once, on the build machine.
FROM --platform=$BUILDPLATFORM node:24-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY web/ ./
RUN npm run build

# Go cross-compiles (pure Go SQLite, no cgo): no emulation needed for the target architecture.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
ENV CGO_ENABLED=0
ARG TARGETOS TARGETARCH
# Dependencies compiled in their own layer (the pure Go SQLite is the slow part): reused until go.mod changes.
# Keep -trimpath identical to the final build, or Go's build cache misses and recompiles everything.
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath modernc.org/sqlite github.com/BurntSushi/toml
COPY cmd/ cmd/
COPY internal/ internal/
COPY web/*.go web/
COPY --from=web /src/web/dist web/dist
ARG VERSION=dev
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o /out/open-niscat ./cmd/open-niscat

FROM alpine:3.24
LABEL org.opencontainers.image.title="Open Niscat" \
      org.opencontainers.image.description="Viewer for the NISCAT parts catalog (data not included)" \
      org.opencontainers.image.source="https://github.com/Bakeneko/open-niscat" \
      org.opencontainers.image.licenses="MIT"
RUN addgroup -S -g 10001 app && adduser -S -D -H -u 10001 -G app app
COPY --from=build /out/open-niscat /usr/local/bin/open-niscat
USER app
EXPOSE 8080
# Settings as environment variables, so that compose or docker run -e can override any of them.
ENV OPEN_NISCAT_DATA=/data \
    OPEN_NISCAT_ADDR=0.0.0.0:8080 \
    OPEN_NISCAT_OPEN_BROWSER=false
# Probes port 8080: keep it inside the container, remap with ports instead.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s CMD wget -q --spider http://127.0.0.1:8080/health
# The NISCAT data is not in the image: mount it read-only at /data.
ENTRYPOINT ["open-niscat"]
