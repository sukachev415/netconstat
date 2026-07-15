# ── Stage 1: Build ──
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/netconstat ./cmd/netconstat

# ── Stage 2: Runtime ──
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/netconstat /usr/local/bin/netconstat
EXPOSE 2055/udp 8080/tcp
ENTRYPOINT ["netconstat"]
