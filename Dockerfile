FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/epaperbackend ./server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/epaperbackend /app/epaperbackend
COPY services /app/services
ENV PORT=5678
ENV DEBUG_PORT=4242
ENV DATA_DIR=/data
EXPOSE 5678 4242
VOLUME ["/data"]
ENTRYPOINT ["/app/epaperbackend"]
