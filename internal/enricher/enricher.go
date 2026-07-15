package enricher

import (
	"log"

	"github.com/sukachev415/netconstat/internal/db"
	"github.com/sukachev415/netconstat/internal/enricher/asn"
	"github.com/sukachev415/netconstat/internal/parser"
)

// Enricher enriches parsed flow records with ASN data.
type Enricher struct {
	resolver *asn.CachedResolver
}

func New(resolver *asn.CachedResolver) *Enricher {
	return &Enricher{resolver: resolver}
}

// Enrich takes a batch of FlowRecords and returns EnrichedFlows with ASN info.
func (e *Enricher) Enrich(records []parser.FlowRecord) []db.EnrichedFlow {
	result := make([]db.EnrichedFlow, len(records))
	for i, r := range records {
		ef := db.EnrichedFlow{
			Timestamp:   r.Timestamp,
			SrcAddr:     r.SrcAddr,
			DstAddr:     r.DstAddr,
			SrcPort:     r.SrcPort,
			DstPort:     r.DstPort,
			Protocol:    r.Protocol,
			TrafficMark: r.TrafficMark,
			Bytes:       r.Bytes,
			Packets:     r.Packets,
			InputIf:     r.InputIf,
			OutputIf:    r.OutputIf,
		}

		if r.SrcAddr != "" {
			ef.SrcASN, ef.SrcASNOrg = e.resolver.Lookup(r.SrcAddr)
		}
		if r.DstAddr != "" {
			ef.DstASN, ef.DstASNOrg = e.resolver.Lookup(r.DstAddr)
		}

		result[i] = ef
	}
	return result
}

func init() {
	log.Println("[enricher] initialized")
}
