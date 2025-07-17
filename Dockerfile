FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gdelt-downloader main.go

# Create wrapper script in builder stage
RUN echo '#!/bin/sh' > /app/gdelt-downloader-wrapper && \
    echo 'exec /usr/local/bin/gdelt-downloader "$@"' >> /app/gdelt-downloader-wrapper && \
    chmod +x /app/gdelt-downloader-wrapper

FROM chainguard/wolfi-base:latest

WORKDIR /gdelt_data

COPY --from=builder /app/gdelt-downloader /usr/local/bin/gdelt-downloader
COPY --from=builder /app/gdelt-downloader-wrapper /usr/local/bin/gdelt-downloader-wrapper

LABEL org.opencontainers.image.title="GDELT Downloader"
LABEL org.opencontainers.image.description="A Go-based tool for downloading data from GDELT"
LABEL org.opencontainers.image.source="https://github.com/Sudo-Ivan/gdelt-downloader"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.authors="Sudo-Ivan"

ENTRYPOINT ["/usr/local/bin/gdelt-downloader-wrapper"]
