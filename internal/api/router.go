package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sukachev415/netconstat/internal/collector"
	"github.com/sukachev415/netconstat/internal/db"
)

type Server struct {
	db        *db.DuckDB
	collector *collector.Collector
	startTime time.Time
}

func NewServer(database *db.DuckDB, coll *collector.Collector) *Server {
	return &Server{
		db:        database,
		collector: coll,
		startTime: time.Now(),
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/flows", s.handleFlows)
	mux.HandleFunc("GET /api/v1/stats/traffic", s.handleStatsTraffic)
	mux.HandleFunc("GET /api/v1/stats/top-services", s.handleTopServices)
	mux.HandleFunc("GET /api/v1/stats/protocols", s.handleProtocols)
	mux.HandleFunc("GET /api/v1/services", s.handleServices)
	mux.HandleFunc("GET /api/v1/devices", s.handleDevices)

	return s.corsMiddleware(s.loggingMiddleware(mux))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	pktsRecv, pktsErr, flowsRecv := s.collector.Stats()
	totalFlows, _ := s.db.TotalFlows()

	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ok",
		"uptime":        time.Since(s.startTime).String(),
		"packets_recv":  pktsRecv,
		"packets_err":   pktsErr,
		"flows_recv":    flowsRecv,
		"flows_in_db":   totalFlows,
	})
}

func (s *Server) handleFlows(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from := q.Get("from")
	to := q.Get("to")
	sourceIP := q.Get("source_ip")
	destIP := q.Get("dest_ip")
	service := q.Get("service")
	protocol := q.Get("protocol")

	limit := queryInt(q, "limit", 100)
	offset := queryInt(q, "offset", 0)

	query := `SELECT timestamp, source_ip, dest_ip, source_port, dest_port, protocol, 
		traffic_mark, bytes, packets, COALESCE(dst_asn_org, 'Unknown') as service
		FROM flows WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND timestamp >= $1"
		args = append(args, from)
	}
	if to != "" {
		query += " AND timestamp <= $2"
		args = append(args, to)
	}
	if sourceIP != "" {
		query += " AND source_ip = $" + itoa(len(args)+1)
		args = append(args, sourceIP)
	}
	if destIP != "" {
		query += " AND dest_ip = $" + itoa(len(args)+1)
		args = append(args, destIP)
	}
	if service != "" {
		query += " AND dst_asn_org ILIKE $" + itoa(len(args)+1)
		args = append(args, "%"+service+"%")
	}
	if protocol != "" {
		query += " AND protocol = $" + itoa(len(args)+1)
		args = append(args, protocol)
	}

	query += " ORDER BY timestamp DESC LIMIT $" + itoa(len(args)+1) + " OFFSET $" + itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var flows []map[string]any
	for rows.Next() {
		var ts time.Time
		var srcIP, dstIP string
		var srcPort, dstPort, proto, mark int
		var bytes, packets int64
		var service string

		err := rows.Scan(&ts, &srcIP, &dstIP, &srcPort, &dstPort, &proto, &mark, &bytes, &packets, &service)
		if err != nil {
			continue
		}
		flows = append(flows, map[string]any{
			"timestamp":   ts,
			"source_ip":   srcIP,
			"dest_ip":     dstIP,
			"source_port": srcPort,
			"dest_port":   dstPort,
			"protocol":    proto,
			"traffic_mark": mark,
			"bytes":       bytes,
			"packets":     packets,
			"service":     service,
		})
	}

	if flows == nil {
		flows = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": flows,
		"meta": map[string]any{
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (s *Server) handleStatsTraffic(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	interval := q.Get("interval")
	groupBy := q.Get("group_by")

	if interval == "" {
		interval = "1h"
	}
	if groupBy == "" {
		groupBy = "service"
	}

	timeBucket := map[string]string{
		"1m":  "minute",
		"5m":  "minute",
		"15m": "minute",
		"1h":  "hour",
		"1d":  "day",
	}[interval]
	if timeBucket == "" {
		timeBucket = "hour"
	}

	groupCol := "COALESCE(dst_asn_org, 'Unknown')"
	if groupBy == "protocol" {
		groupCol = "protocol::VARCHAR"
	}

	query := `SELECT date_trunc('` + timeBucket + `', timestamp) as bucket, ` + groupCol + ` as label, 
		SUM(bytes) as total_bytes, SUM(packets) as total_packets, COUNT(*) as flow_count
		FROM flows WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND timestamp >= $" + itoa(len(args)+1)
		args = append(args, from)
	}
	if to != "" {
		query += " AND timestamp <= $" + itoa(len(args)+1)
		args = append(args, to)
	}

	query += " GROUP BY bucket, label ORDER BY bucket, total_bytes DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var data []map[string]any
	for rows.Next() {
		var bucket time.Time
		var label string
		var totalBytes, totalPackets, flowCount int64
		if err := rows.Scan(&bucket, &label, &totalBytes, &totalPackets, &flowCount); err != nil {
			continue
		}
		data = append(data, map[string]any{
			"bucket":       bucket,
			"label":        label,
			"total_bytes":  totalBytes,
			"total_packets": totalPackets,
			"flow_count":   flowCount,
		})
	}

	if data == nil {
		data = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (s *Server) handleTopServices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	top := queryInt(q, "top", 10)

	query := `SELECT COALESCE(dst_asn_org, 'Unknown') as service, 
		SUM(bytes) as total_bytes, SUM(packets) as total_packets, COUNT(*) as flow_count
		FROM flows GROUP BY service ORDER BY total_bytes DESC LIMIT $1`

	rows, err := s.db.Query(query, top)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var data []map[string]any
	for rows.Next() {
		var service string
		var totalBytes, totalPackets, flowCount int64
		if err := rows.Scan(&service, &totalBytes, &totalPackets, &flowCount); err != nil {
			continue
		}
		data = append(data, map[string]any{
			"service":      service,
			"total_bytes":  totalBytes,
			"total_packets": totalPackets,
			"flow_count":   flowCount,
		})
	}

	if data == nil {
		data = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (s *Server) handleProtocols(w http.ResponseWriter, r *http.Request) {
	query := `SELECT protocol, SUM(bytes) as total_bytes, COUNT(*) as flow_count
		FROM flows GROUP BY protocol ORDER BY total_bytes DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var data []map[string]any
	for rows.Next() {
		var proto int
		var totalBytes, flowCount int64
		if err := rows.Scan(&proto, &totalBytes, &flowCount); err != nil {
			continue
		}
		data = append(data, map[string]any{
			"protocol":    proto,
			"name":        protocolName(proto),
			"total_bytes": totalBytes,
			"flow_count":  flowCount,
		})
	}

	if data == nil {
		data = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`SELECT DISTINCT dst_asn_org FROM flows 
		WHERE dst_asn_org IS NOT NULL AND dst_asn_org != '' AND dst_asn_org != 'Unknown'
		ORDER BY 1 LIMIT 500`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var svc string
		if err := rows.Scan(&svc); err == nil {
			services = append(services, svc)
		}
	}
	if services == nil {
		services = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": services})
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`SELECT DISTINCT source_ip FROM flows 
		WHERE source_ip LIKE '192.168.%' OR source_ip LIKE '10.%' OR source_ip LIKE '172.1%.%' OR source_ip LIKE '172.2%.%' OR source_ip LIKE '172.3%.%'
		ORDER BY 1 LIMIT 500`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var devices []string
	for rows.Next() {
		var dev string
		if err := rows.Scan(&dev); err == nil {
			devices = append(devices, dev)
		}
	}
	if devices == nil {
		devices = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": devices})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func queryInt(q map[string][]string, key string, def int) int {
	if vals, ok := q[key]; ok && len(vals) > 0 {
		var n int
		if _, err := fmt.Sscanf(vals[0], "%d", &n); err == nil {
			return n
		}
	}
	return def
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func protocolName(proto int) string {
	switch proto {
	case 1:
		return "ICMP"
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 47:
		return "GRE"
	case 50:
		return "ESP"
	case 58:
		return "ICMPv6"
	case 132:
		return "SCTP"
	default:
		return fmt.Sprintf("Proto-%d", proto)
	}
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[api] %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
