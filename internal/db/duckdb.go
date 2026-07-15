package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// EnrichedFlow is a flow record with ASN enrichment data.
type EnrichedFlow struct {
	Timestamp   int64
	SrcAddr     string
	DstAddr     string
	SrcPort     uint16
	DstPort     uint16
	Protocol    uint8
	TrafficMark uint32
	Bytes       uint64
	Packets     uint64
	InputIf     uint16
	OutputIf    uint16
	SrcASN      uint
	SrcASNOrg   string
	DstASN      uint
	DstASNOrg   string
}

type DuckDB struct {
	conn *sql.DB
}

func Open(path string, maxConns int) (*DuckDB, error) {
	if dir := parentDir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	conn, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}

	conn.SetMaxOpenConns(maxConns)

	database := &DuckDB{conn: conn}
	if err := database.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	log.Printf("[db] opened %s", path)
	return database, nil
}

func (db *DuckDB) Close() error {
	return db.conn.Close()
}

func (db *DuckDB) migrate() error {
	queries := []string{
		`CREATE SEQUENCE IF NOT EXISTS seq_flows_id START 1`,
		`CREATE TABLE IF NOT EXISTS flows (
			id             BIGINT       DEFAULT nextval('seq_flows_id') PRIMARY KEY,
			timestamp      TIMESTAMP    NOT NULL,
			source_ip      VARCHAR(45)  NOT NULL,
			dest_ip        VARCHAR(45)  NOT NULL,
			source_port    INTEGER,
			dest_port      INTEGER,
			protocol       SMALLINT     NOT NULL,
			traffic_mark   INTEGER      NOT NULL DEFAULT 0,
			bytes          BIGINT       NOT NULL DEFAULT 0,
			packets        BIGINT       NOT NULL DEFAULT 0,
			src_asn        INTEGER,
			src_asn_org    VARCHAR(255),
			dst_asn        INTEGER,
			dst_asn_org    VARCHAR(255),
			input_iface    INTEGER,
			output_iface   INTEGER
		)`,
		`CREATE INDEX IF NOT EXISTS idx_flows_timestamp ON flows (timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_flows_src_ip    ON flows (source_ip)`,
		`CREATE INDEX IF NOT EXISTS idx_flows_dst_ip    ON flows (dest_ip)`,
		`CREATE INDEX IF NOT EXISTS idx_flows_dst_asn   ON flows (dst_asn)`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", truncate(q, 60), err)
		}
	}

	return nil
}

// InsertFlows batch-inserts enriched flow records.
func (db *DuckDB) InsertFlows(records []EnrichedFlow) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO flows
		(timestamp, source_ip, dest_ip, source_port, dest_port, protocol, traffic_mark,
		 bytes, packets, src_asn, src_asn_org, dst_asn, dst_asn_org, input_iface, output_iface)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		ts := time.Unix(r.Timestamp, 0).UTC()
		_, err := stmt.Exec(
			ts,
			r.SrcAddr,
			r.DstAddr,
			r.SrcPort,
			r.DstPort,
			r.Protocol,
			r.TrafficMark,
			r.Bytes,
			r.Packets,
			r.SrcASN,
			nullStr(r.SrcASNOrg),
			r.DstASN,
			nullStr(r.DstASNOrg),
			r.InputIf,
			r.OutputIf,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("insert flow: %w", err)
		}
	}

	return tx.Commit()
}

func nullStr(s string) interface{} {
	if s == "" || s == "Unknown" {
		return nil
	}
	return s
}

// Cleanup deletes flows older than retentionDays.
func (db *DuckDB) Cleanup(retentionDays int) (int64, error) {
	result, err := db.conn.Exec(
		`DELETE FROM flows WHERE timestamp < now() - INTERVAL '1 day' * $1`,
		retentionDays,
	)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return n, nil
}

// TotalFlows returns total flow count.
func (db *DuckDB) TotalFlows() (int64, error) {
	var count int64
	err := db.conn.QueryRow("SELECT COUNT(*) FROM flows").Scan(&count)
	return count, err
}

// Query wraps conn.Query for the API layer.
func (db *DuckDB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
