.PHONY: build run clean docker-up docker-down download-asn

# Build Go binary
build:
	CGO_ENABLED=1 go build -o bin/netconstat ./cmd/netconstat/

# Run locally
run: build
	./bin/netconstat

# Build frontend
frontend:
	cd frontend && npm install && npm run build

# Docker
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# Download ASN database (O-X-L, no registration required)
download-asn:
	@echo "Downloading ASN database..."
	curl -L "https://geoip.oxl.app/file/asn_ipv4_small.mmdb.zip" -o /tmp/asn.zip
	unzip -o /tmp/asn.zip -d data/asn/
	rm -f /tmp/asn.zip
	@echo "ASN database saved to data/asn/"

# Clean
clean:
	rm -rf bin/ data/netconstat.duckdb frontend/dist/
