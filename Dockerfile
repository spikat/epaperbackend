FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/epaperbackend ./server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 1000 app
WORKDIR /app
COPY --from=builder /out/epaperbackend /app/epaperbackend
COPY services /app/services
RUN mkdir -p /data && chown -R app:app /app /data

ENV PORT=5678 \
	DEBUG=false \
	DEBUG_PORT=4242 \
	DATA_DIR=/data \
	MAIN_API_BASE_URL=http://127.0.0.1:5678 \
	TZ=Europe/Paris

EXPOSE 5678 4242
VOLUME ["/data"]
USER app
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget -qO- http://127.0.0.1:5678/health >/dev/null || exit 1
ENTRYPOINT ["/app/epaperbackend"]
