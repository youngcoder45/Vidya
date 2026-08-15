# ============================================================================
# SchoolOS API — production image
# Stage 1: build; Stage 2: distroless runtime (non-root).
# ============================================================================

FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/schoolos-api ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/schoolos-api /app/schoolos-api
# Migrations are applied by a separate CI job; the image ships them read-only
# for convenience (production uses golang-migrate applied before rollout).
COPY migrations /app/migrations
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/schoolos-api"]
