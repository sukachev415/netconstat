# NetConStat — Architecture Document

> NetFlow v9 collector & visualization web application

---

## 1. High-Level Overview

```
┌──────────────┐       UDP/2055        ┌──────────────────────────────────────────┐
│  Router/Switch│ ────────────────────► │  netconstat (Go)                         │
│  (NetFlow v9) │   NetFlow v9 packets  │                                          │
└──────────────┘                        │  ┌──────────┐  ┌──────────┐  ┌────────┐ │
                                        │  │ collector │→ │  parser  │→│enricher│ │
                                        │  └──────────┘  └──────────┘  └───┬────┘ │
                                        │                                  │       │
                                        │  ┌──────────┐              ┌────▼─────┐ │
                                        │  │cleanup job│◄─────────────│    db    │ │
                                        │  └──────────┘              └────▲─────┘ │
                                        │                                 │       │
                                        │                           ┌─────┴────┐  │
                                        │                           │   api    │  │
                                        │                           └─────┬────┘  │
                                        └──────────────────────────────────┼───────┘
                                                                           │ HTTP
                                        ┌──────────────────────────────────┴───────┐
                                        │  React + Vite + TailwindCSS (frontend)    │
                                        │  served by Go or nginx in production       │
                                        └───────────────────────────────────────────┘
```

---

## 2. Folder Structure

```
netconstat/
├── ARCHITECTURE.md               ← this file
├── .env.example                  ← all configurable env vars
├── docker-compose.yml
├── Dockerfile.go                 ← multi-stage Go build
├── Dockerfile.frontend           ← multi-stage React build
├── Makefile                      ← dev shortcuts
│
├── cmd/
│   └── netconstat/
│       └── main.go               ← entrypoint: wires everything together
│
├── internal/
│   ├── collector/
│   │   ├── udp.go                ← UDP listener on :2055
│   │   └── udp_test.go
│   ├── parser/
│   │   ├── netflow9.go           ← NFv9 header/template/flowset parsing
│   │   ├── fields.go             ← field type registry (IP, ports, protocol, bytes, mark…)
│   │   └── netflow9_test.go
│   ├── enricher/
│   │   ├── enricher.go           ← orchestrator: parse → enrich → store
│   │   ├── asn/
│   │   │   ├── loader.go         ← load GeoLite2-ASN mmdb (or CSV fallback)
│   │   │   ├── resolver.go       ← IP → ASN number + org name
│   │   │   └── cache.go          ← sync.Map / LRU in-memory cache with TTL
│   │   └── enricher_test.go
│   ├── db/
│   │   ├── duckdb.go             ← open/init/migrate DuckDB
│   │   ├── flows.go              ← insert/query flows
│   │   ├── stats.go              ← aggregation queries for API
│   │   ├── cleanup.go            ← retention goroutine (configurable days)
│   │   └── duckdb_test.go
│   ├── api/
│   │   ├── router.go             ← chi/http.ServeMux routes
│   │   ├── handlers_flows.go     ← flow query endpoints
│   │   ├── handlers_stats.go     ← aggregation/chart endpoints
│   │   ├── handlers_settings.go  ← health, config, ASN status
│   │   ├── middleware.go         ← CORS, logging, request-id
│   │   └── types.go              ← request/response DTOs
│   └── config/
│       └── config.go             ← loads .env via envparse, validates
│
├── frontend/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── postcss.config.js
│   │
│   ├── public/
│   │   └── favicon.svg
│   │
│   └── src/
│       ├── main.tsx               ← React entry
│       ├── App.tsx                ← router + theme provider
│       ├── api/
│       │   └── client.ts          ← fetch wrapper for REST API
│       ├── hooks/
│       │   ├── useFlows.ts
│       │   ├── useStats.ts
│       │   └── useTheme.ts
│       ├── components/
│       │   ├── layout/
│       │   │   ├── Header.tsx
│       │   │   ├── Sidebar.tsx
│       │   │   └── Footer.tsx
│       │   ├── charts/
│       │   │   ├── TrafficChart.tsx    ← stacked bar / line chart
│       │   │   ├── TopServicesTable.tsx
│       │   │   └── UnitToggle.tsx      ← KB / MB / GB
│       │   ├── filters/
│       │   │   ├── TimeIntervalSwitch.tsx  ← 30m, 1h, 12h, 1d, 7d, 30d
│       │   │   ├── DeviceFilter.tsx
│       │   │   └── ServiceFilter.tsx
│       │   └── ui/
│       │       ├── ThemeToggle.tsx
│       │       ├── LoadingSpinner.tsx
│       │       └── ErrorBoundary.tsx
│       ├── pages/
│       │   ├── Dashboard.tsx      ← main overview
│       │   └── Settings.tsx       ← retention, ASN DB status, env info
│       ├── contexts/
│       │   └── ThemeContext.tsx    ← light/dark
│       ├── types/
│       │   └── index.ts           ← TS interfaces matching API DTOs
│       └── styles/
│           └── globals.css
│
├── data/                          ← mounted volume in Docker
│   ├── netconstat.duckdb          ← DuckDB file
│   └── asn/
│       └── GeoLite2-ASN.mmdb     ← MaxMind ASN database
│
└── scripts/
    ├── download-asn-db.sh         ← fetch & update GeoLite2-ASN
    └── dev.sh                     ← hot-reload dev server helper
```

---

## 3. Database Schema (DuckDB)

DuckDB is column-oriented and excels at analytical aggregations over time-series data.

### 3.1 Flows Table

```sql
CREATE SEQUENCE IF NOT EXISTS seq_flows_id START 1;

CREATE TABLE IF NOT EXISTS flows (
    id             BIGINT       DEFAULT nextval('seq_flows_id') PRIMARY KEY,
    timestamp      TIMESTAMP    NOT NULL,            -- flow start time
    source_ip      VARCHAR(45)  NOT NULL,            -- supports IPv4 & IPv6
    dest_ip        VARCHAR(45)  NOT NULL,
    source_port    INTEGER,
    dest_port      INTEGER,
    protocol       SMALLINT     NOT NULL,            -- 6=TCP, 17=UDP, etc.
    traffic_mark   INTEGER      NOT NULL DEFAULT 0,  -- DSCP / ToS / custom mark
    bytes          BIGINT       NOT NULL DEFAULT 0,  -- total bytes in flow
    packets        BIGINT       NOT NULL DEFAULT 0,
    src_asn        INTEGER,                           -- resolved ASN number
    src_asn_org    VARCHAR(255),                      -- ASN org name (service label)
    dst_asn        INTEGER,
    dst_asn_org    VARCHAR(255),
    input_iface    INTEGER,                           -- SNMP interface index
    output_iface   INTEGER
);

-- Time-partitioned index for fast range scans
CREATE INDEX IF NOT EXISTS idx_flows_timestamp ON flows (timestamp);
CREATE INDEX IF NOT EXISTS idx_flows_src_asn   ON flows (src_asn);
CREATE INDEX IF NOT EXISTS idx_flows_dst_asn   ON flows (dst_asn);
```

### 3.2 Template Cache Table (optional, for debugging)

```sql
CREATE TABLE IF NOT EXISTS nf9_templates (
    source_ip      VARCHAR(45) NOT NULL,
    template_id    INTEGER     NOT NULL,
    field_count    INTEGER     NOT NULL,
    fields_json    VARCHAR     NOT NULL,   -- JSON array of {type, length}
    received_at    TIMESTAMP   DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (source_ip, template_id)
);
```

### 3.3 Aggregation Materialized View (optional perf optimization)

```sql
CREATE TABLE IF NOT EXISTS flows_hourly (
    hour           TIMESTAMP   NOT NULL,
    source_ip      VARCHAR(45) NOT NULL,
    dest_ip        VARCHAR(45) NOT NULL,
    service        VARCHAR(255),          -- src_asn_org or dst_asn_org
    protocol       SMALLINT,
    traffic_mark   INTEGER,
    total_bytes    BIGINT,
    total_packets  BIGINT,
    flow_count     INTEGER,
    PRIMARY KEY (hour, source_ip, dest_ip, service, protocol, traffic_mark)
);
```

### 3.4 Retention

The cleanup job runs periodically and executes:

```sql
DELETE FROM flows
WHERE timestamp < now() - INTERVAL '1 day' * :retention_days;
```

Default `:retention_days = 30`. Configurable via `RETENTION_DAYS` env var.

---

## 4. Module Design — IP-to-Service Mapping via ASN

### 4.1 Architecture

```
          ┌──────────────┐
          │  GeoLite2-ASN │  (MaxMind mmdb, loaded at startup)
          │    .mmdb      │
          └──────┬───────┘
                 │ open once at boot
          ┌──────▼───────┐
          │  asn/loader   │  file path from config, graceful error if missing
          └──────┬───────┘
                 │
          ┌──────▼───────┐
          │asn/resolver   │  Lookup(ip) → (asn, org)
          │  (stateless)  │  wraps mmdb reader
          └──────┬───────┘
                 │
          ┌──────▼───────┐
          │  asn/cache    │  sync.Map or golang-lru
          │  (in-memory)  │  TTL=1h, max 500k entries
          └──────────────┘
```

### 4.2 Data Flow

1. **Startup**: `loader.go` opens the `.mmdb` file. If the file is missing or corrupted, it logs a warning and the resolver returns `("", "Unknown")` for every lookup.
2. **Per-flow enrichment**: When a parsed flow arrives, the enricher calls `cache.Get(ip)`:
   - **Cache hit**: returns cached `(asn, org)`.
   - **Cache miss**: calls `resolver.Lookup(ip)`, stores result in cache with TTL.
3. **Fallback**: If resolver has no DB loaded → returns `(0, "Unknown")`.
4. **Periodic reload**: A ticker every 24h checks if the `.mmdb` file's mtime changed and reloads it (supports live updates without restart).

### 4.3 Interface Contract

```go
// internal/enricher/asn/resolver.go
type Resolver interface {
    Lookup(ip net.IP) (asn int, org string, err error)
}

// internal/enricher/asn/cache.go
type CachedResolver struct {
    inner Resolver
    cache *lru.Cache  // or sync.Map
    ttl   time.Duration
}
```

### 4.4 Database Fallback Strategy

If the GeoLite2-ASN `.mmdb` is not available, an alternative approach uses a pre-built CSV mapping `prefix → asn → org`. The loader tries `.mmdb` first, then falls back to a CSV at `data/asn/asn.csv`. If neither exists, all flows are tagged as `"Unknown"`.

---

## 5. API Endpoints (REST)

Base URL: `/api/v1`

| Method | Path                          | Description                                           |
|--------|-------------------------------|-------------------------------------------------------|
| `GET`  | `/health`                     | Health check: DB status, collector uptime, ASN status |
|        |                               |                                                       |
| **Flows**                                                     |
| `GET`  | `/flows`                      | List flows with pagination & filters                  |
|        | `?from=&to=`                  | ISO-8601 timestamps                                   |
|        | `&source_ip=&dest_ip=`        | Exact or CIDR match                                   |
|        | `&service=`                   | Filter by ASN org name (case-insensitive LIKE)        |
|        | `&protocol=`                  | 6, 17, etc.                                           |
|        | `&limit=&offset=`             | Pagination (default limit=100)                        |
|        |                               |                                                       |
| **Statistics / Charts**                                       |
| `GET`  | `/stats/traffic`              | Aggregated bytes over time, bucketed by interval      |
|        | `?from=&to=`                  | Time range                                            |
|        | `&interval=`                  | `1m`, `5m`, `15m`, `1h`, `1d`                         |
|        | `&service=&protocol=`         | Optional filters                                      |
|        | `&group_by=`                  | `service`, `protocol`, `traffic_mark` (default=`service`) |
|        |                               |                                                       |
| `GET`  | `/stats/top-services`         | Top N services by bytes in time range                 |
|        | `?from=&to=&top=10`           | Default top=10                                        |
|        |                               |                                                       |
| `GET`  | `/stats/top-destinations`     | Top N destination IPs/subnets by bytes                 |
|        | `?from=&to=&top=10`           |                                                       |
|        |                               |                                                       |
| `GET`  | `/stats/protocols`            | Protocol breakdown (bytes per protocol)               |
|        |                               |                                                       |
| **Settings / Meta**                                           |
| `GET`  | `/settings`                   | Current config snapshot (retention, ASN status, etc.) |
| `GET`  | `/services`                   | Distinct ASN org names (for filter dropdown)          |
| `GET`  | `/devices`                    | Distinct source IPs sending NetFlow (exporters)       |

### Response Format

```json
{
  "data": { ... },
  "meta": {
    "total": 12345,
    "limit": 100,
    "offset": 0
  }
}
```

---

## 6. Frontend Component Structure

### 6.1 Page Layout

```
┌─────────────────────────────────────────────────────┐
│  Header          [Dark/Light toggle]                 │
├──────────┬──────────────────────────────────────────┤
│          │  TimeIntervalSwitch: [30m] [1h] [12h] [1d]│
│ Sidebar  │  Filters: Device ▼  Service ▼             │
│          │  Unit: [KB] [MB] [GB]                     │
│ - Dash   │                                            │
│ - Sett.  │  ┌─────────────────────────────────────┐  │
│          │  │  TrafficChart (stacked bar / line)   │  │
│          │  └─────────────────────────────────────┘  │
│          │                                            │
│          │  ┌─────────────────────────────────────┐  │
│          │  │  TopServicesTable                   │  │
│          │  └─────────────────────────────────────┘  │
│          │                                            │
│          │  ┌─────────────────────────────────────┐  │
│          │  │  ProtocolBreakdown (donut/pie)      │  │
│          │  └─────────────────────────────────────┘  │
├──────────┴──────────────────────────────────────────┤
│  Footer: collector uptime, last flow received        │
└─────────────────────────────────────────────────────┘
```

### 6.2 Key Components

| Component               | Responsibility                                                 |
|--------------------------|----------------------------------------------------------------|
| `Dashboard.tsx`         | Composes all chart/filter components, manages query state      |
| `Settings.tsx`          | Shows retention days, ASN DB status, version, config           |
| `TimeIntervalSwitch.tsx`| Toggle buttons: 30m, 1h, 12h, 1d, 7d, 30d → sets `from`/`to` |
| `DeviceFilter.tsx`      | Dropdown populated from `/devices` endpoint                    |
| `ServiceFilter.tsx`     | Dropdown populated from `/services` endpoint                   |
| `UnitToggle.tsx`        | Switch between KB/MB/GB display; multiplies/divides chart data |
| `TrafficChart.tsx`      | Uses Recharts or Chart.js; stacked bar for services over time  |
| `TopServicesTable.tsx`  | Sorted table of top N services by traffic volume               |
| `ThemeToggle.tsx`       | Sun/Moon icon, toggles ThemeContext                             |
| `ThemeContext.tsx`       | Stores `light`/`dark`, applies class to `<html>`               |

### 6.3 State Management

- **React Query (TanStack Query)** for server state: `useFlows`, `useStats`, `useDevices`, `useServices`.
- **React Context** for UI state: theme, selected filters, time range, unit.
- No Redux needed — the app is read-heavy with simple local UI state.

### 6.4 Theming

TailwindCSS `dark:` variant. `<html class="dark">` toggled by ThemeContext. All components use `bg-white dark:bg-gray-900`, `text-gray-900 dark:text-gray-100`, etc. Persisted to `localStorage`.

---

## 7. Docker Setup

### 7.1 `Dockerfile.go` (multi-stage)

```dockerfile
# ── Stage 1: Build ──
FROM golang:1.22-alpine AS builder
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
```

### 7.2 `Dockerfile.frontend` (multi-stage)

```dockerfile
# ── Stage 1: Build ──
FROM node:22-alpine AS builder
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ── Stage 2: Serve ──
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY scripts/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

### 7.3 `docker-compose.yml`

```yaml
version: "3.9"

services:
  netconstat:
    build:
      context: .
      dockerfile: Dockerfile.go
    container_name: netconstat-api
    restart: unless-stopped
    env_file: .env
    ports:
      - "${NETFLOW_PORT:-2055}:2055/udp"
      - "${API_PORT:-8080}:8080/tcp"
    volumes:
      - ./data:/data
    environment:
      - DUCKDB_PATH=/data/netconstat.duckdb
      - ASN_DB_PATH=/data/asn/GeoLite2-ASN.mmdb
      - RETENTION_DAYS=30
      - API_ADDR=0.0.0.0:8080
      - NETFLOW_ADDR=0.0.0.0:2055
      - LOG_LEVEL=info

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    container_name: netconstat-ui
    restart: unless-stopped
    ports:
      - "${WEB_PORT:-3000}:80"
    depends_on:
      - netconstat
    volumes:
      - ./scripts/nginx.conf:/etc/nginx/conf.d/default.conf:ro
```

### 7.4 `scripts/nginx.conf`

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API calls to Go backend
    location /api/ {
        proxy_pass http://netconstat:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 8. `.env` Configuration

```env
# ── NetFlow Collector ──
NETFLOW_ADDR=0.0.0.0:2055
NETFLOW_BUFFER_SIZE=65535

# ── API ──
API_ADDR=0.0.0.0:8080
API_CORS_ORIGINS=*

# ── Database ──
DUCKDB_PATH=./data/netconstat.duckdb
DUCKDB_MAX_OPEN_CONNS=4

# ── Retention ──
RETENTION_DAYS=30
CLEANUP_INTERVAL_HOURS=6

# ── ASN Database ──
ASN_DB_PATH=./data/asn/GeoLite2-ASN.mmdb
ASN_CACHE_TTL_HOURS=1
ASN_CACHE_MAX_ENTRIES=500000

# ── Logging ──
LOG_LEVEL=info

# ── Frontend (build-time) ──
VITE_API_BASE_URL=/api/v1
```

---

## 9. Data Flow — End to End

```
1. UDP Listener (collector/udp.go)
   - Binds UDP socket on :2055
   - Reads raw packets into buffer pool
   - Sends []byte to parse channel

2. Parser (parser/netflow9.go)
   - Validates NFv9 header (version=9, source ID)
   - Caches templates (Template FlowSets)
   - Parses Data FlowSets against cached templates
   - Returns []FlowRecord (raw fields)

3. Enricher (enricher/enricher.go)
   - For each FlowRecord:
     a. Resolve src IP → (src_asn, src_asn_org) via cached resolver
     b. Resolve dst IP → (dst_asn, dst_asn_org)
     c. Map well-known ports to protocol names (optional)
   - Returns []EnrichedFlow

4. DB Writer (db/flows.go)
   - Batch INSERT using DuckDB Appender API (high-throughput)
   - Flushes every 1 second or 1000 flows (whichever first)

5. Cleanup (db/cleanup.go)
   - Goroutine with configurable ticker
   - DELETE WHERE timestamp < now() - retention_days

6. API (api/router.go)
   - Serves REST endpoints
   - Reads from DuckDB with parameterized queries

7. Frontend
   - Polls /stats/* endpoints every 30s (or on filter change)
   - Renders charts with Recharts
```

---

## 10. Edge Cases & Error Handling

| Edge Case                    | Handling                                                     |
|------------------------------|--------------------------------------------------------------|
| Corrupted UDP packet         | Parser returns error → logged, packet dropped, counter incr  |
| Unknown template ID          | Cache flowset bytes, request retransmit (NFv9 option), skip  |
| Missing ASN database         | Resolver returns `(0, "Unknown")`, warn once at startup      |
| DB file corruption           | On startup, run `PRAGMA integrity_check`; rename & recreate  |
| DB overflow / disk full      | Backpressure: collector pauses writes, logs critical alert   |
| High packet rate             | Buffer pool with backpressure; drop + metric if channel full |
| IPv6 flows                   | VARCHAR(45) supports both; parser handles both AFIs          |
| Clock skew in flow timestamps| Store device-supplied timestamp as-is; API filters use local |
| Empty result sets            | Return `{"data": [], "meta": {"total": 0}}`                  |

---

## 11. Technology Choices Summary

| Layer        | Technology                      | Rationale                                    |
|--------------|---------------------------------|----------------------------------------------|
| Language     | Go 1.22+                        | Concurrency, low memory, single binary       |
| Database     | DuckDB                          | Embedded, columnar, fast analytics            |
| Frontend     | React 18 + Vite + TailwindCSS  | Fast builds, utility CSS, huge ecosystem      |
| Charts       | Recharts                        | React-native chart lib, responsive, light/dark|
| HTTP Router  | chi (or stdlib `net/http`)      | Lightweight, idiomatic Go                     |
| State mgmt   | TanStack Query                  | Server state caching, auto-refetch            |
| ASN Lookup   | oschwald/maxminddb-golang       | Fast mmdb reader, maintained                  |
| Container    | Docker + Compose                | Standard deployment, multi-stage builds       |
| CI/CD        | GitHub Actions (optional)       | Free for public repos                         |

---

## 12. Implementation Phases

| Phase | Scope                                                           |
|-------|-----------------------------------------------------------------|
| 1     | UDP collector + NFv9 parser + DuckDB storage                   |
| 2     | ASN enrichment + in-memory cache + cleanup job                 |
| 3     | REST API (flows, stats, top-services)                          |
| 4     | React frontend: dashboard, charts, filters, theme toggle       |
| 5     | Docker multi-stage builds + docker-compose                     |
| 6     | Polish: error boundaries, loading states, README, CI           |
