# syntax=docker/dockerfile:1

# Samakan versi Go dengan baris "go" pada go.mod (ubah lewat --build-arg bila perlu).
ARG GO_VERSION=1.23

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sipasdos .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 app \
 && mkdir -p /app/logs \
 && chown -R app:app /app

WORKDIR /app
COPY --from=build /out/sipasdos /app/sipasdos

USER app
EXPOSE 3000
ENTRYPOINT ["/app/sipasdos"]
