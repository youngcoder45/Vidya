# ============================================================================
# SchoolOS API — production image
# Stage 1: build; Stage 2: minimal Alpine runtime (non-root).
# (Alpine chosen over distroless so compose healthchecks can use curl; swap
# for distroless in production and healthcheck from the orchestrator.)
# ============================================================================

FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/schoolos-api ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/schoolos-api /app/schoolos-api
COPY migrations /app/migrations
USER app
EXPOSE 8080
ENTRYPOINT ["/app/schoolos-api"]
