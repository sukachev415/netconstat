# NetConStat

NetFlow v9 collector & visualization web application.

Collects NetFlow v9 traffic from MikroTik routers, enriches it with ASN data (IP → service name mapping), and visualizes it in a clean web dashboard.

## Features

- **NetFlow v9 Collector** — UDP server on port 2055, parses templates and flow data from MikroTik
- **ASN Enrichment** — Maps destination IPs to service names (Google, Facebook, GitHub, etc.) via local MaxMind-compatible ASN database
- **In-Memory Cache** — LRU cache with configurable TTL for fast repeated lookups
- **DuckDB Storage** — Embedded columnar database, optimized for analytical queries
- **Retention Policy** — Auto-cleanup of old flows (configurable, default 30 days)
- **REST API** — 10 endpoints for flows, traffic stats, top services, protocols
- **Web Dashboard** — React + TailwindCSS with dark/light theme, filters, charts
- **Docker Compose** — One-command deployment

## Quick Start

```bash
# Clone
git clone https://github.com/sukachev415/netconstat.git
cd netconstat

# Copy env file
cp .env.example .env

# Download ASN database (optional, enables IP-to-service mapping)
# Option 1: O-X-L (no registration, daily updates)
curl -L "https://geoip.oxl.app/file/asn_ipv4_small.mmdb.zip" -o /tmp/asn.zip
unzip /tmp/asn.zip -d data/asn/

# Option 2: MaxMind GeoLite2 (requires free account)
# Register at https://www.maxmind.com/en/geolite2/signup
# Download GeoLite2-ASN.mmdb and place in data/asn/

# Run with Docker
docker compose up -d

# Or build locally
go build -o netconstat ./cmd/netconstat/
./netconstat
```

Open http://localhost:3000 for the dashboard.

## MikroTik Configuration

Enable NetFlow v9 export on your MikroTik router:

```
# Add a traffic flow
/ip traffic-flow set enabled=yes

# Add a target (your server IP)
/ip traffic-flow target add dst-address=YOUR_SERVER_IP port=2055 version=9

# Enable on interfaces (adjust as needed)
/ip traffic-flow interface add interface=ether1
```

For more granular traffic analysis, enable traffic marking:

```
# Mark traffic with DSCP or custom marks
/ip firewall mangle add chain=forward dst-address-list=YouTube action=mark-packet new-packet-mark=YouTube
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NETFLOW_ADDR` | `0.0.0.0:2055` | NetFlow v9 UDP listen address |
| `NETFLOW_BUFFER_SIZE` | `65535` | UDP read buffer size |
| `API_ADDR` | `0.0.0.0:8080` | REST API listen address |
| `DUCKDB_PATH` | `./data/netconstat.duckdb` | DuckDB file path |
| `RETENTION_DAYS` | `30` | Days to keep flow records |
| `CLEANUP_INTERVAL_HOURS` | `6` | How often to run cleanup |
| `ASN_DB_PATH` | `./data/asn/GeoLite2-ASN.mmdb` | Path to MaxMind .mmdb file |
| `ASN_CACHE_TTL_HOURS` | `1` | Cache TTL for ASN lookups |
| `ASN_CACHE_MAX_ENTRIES` | `500000` | Max entries in ASN cache |
| `LOG_LEVEL` | `info` | Log level |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check + collector stats |
| GET | `/api/v1/flows` | List flows with pagination & filters |
| GET | `/api/v1/stats/traffic` | Traffic over time, bucketed by interval |
| GET | `/api/v1/stats/top-services` | Top N services by traffic volume |
| GET | `/api/v1/stats/protocols` | Protocol breakdown |
| GET | `/api/v1/services` | Distinct service names (for filter dropdown) |
| GET | `/api/v1/devices` | Distinct source IPs (for filter dropdown) |

## Architecture

```
Router (NetFlow v9) → UDP:2055 → Collector → Parser → Enricher (ASN) → DuckDB
                                                                         ↓
                                                              REST API (:8080)
                                                                         ↓
                                                              React Dashboard (:3000)
```

## License

MIT
