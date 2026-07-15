package parser

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

// NetFlow v9 field type constants
const (
	NF9_IN_BYTES                    = 1
	NF9_IN_PACKETS                  = 2
	NF9_PROTOCOL                    = 4
	NF9_TOS                         = 5
	NF9_TCP_FLAGS                   = 6
	NF9_L4_SRC_PORT                 = 7
	NF9_IPV4_SRC_ADDR               = 8
	NF9_INPUT_SNMP                  = 10
	NF9_L4_DST_PORT                 = 11
	NF9_IPV4_DST_ADDR               = 12
	NF9_OUTPUT_SNMP                 = 14
	NF9_FLOW_START_SEC              = 150
	NF9_FLOW_END_SEC                = 151
	NF9_SRC_AS                      = 16
	NF9_DST_AS                      = 17
	NF9_FIRST_SWITCHED              = 22
	NF9_LAST_SWITCHED               = 21
	NF9_IPV6_SRC_ADDR               = 27
	NF9_IPV6_DST_ADDR               = 28
	NF9_BGP_NEXT_HOP_IPV4          = 18
	NF9_MPLS_TOP_LABEL              = 70
	NF9_FORWARDING_STATUS           = 89

	// FlowSet types
	FlowSetTemplate    = 0
	FlowSetOptionsTmpl = 1
	FlowSetMinData     = 256
)

// FlowRecord represents a single parsed NetFlow v9 flow data record.
type FlowRecord struct {
	SrcAddr     string
	DstAddr     string
	SrcPort     uint16
	DstPort     uint16
	Protocol    uint8
	Bytes       uint64
	Packets     uint64
	Timestamp   int64
	InputIf     uint16
	OutputIf    uint16
	TrafficMark uint32
}

// templateField describes one field in a NetFlow v9 template.
type templateField struct {
	Type   uint16
	Length uint16
}

// template holds a parsed template with its fields and total record length.
type template struct {
	Fields    []templateField
	RecordLen int
}

// NF9Header is the 20-byte NetFlow v9 header.
type NF9Header struct {
	Version   uint16
	Count     uint16
	SysUptime uint32
	UnixSecs  uint32
	UnixNSecs uint32
	FlowSeq   uint32
	SourceID  uint32
}

// Parser manages NetFlow v9 template state and decoding.
type Parser struct {
	mu        sync.RWMutex
	templates map[string]*template
	now       func() time.Time
}

func NewParser() *Parser {
	return &Parser{
		templates: make(map[string]*template),
		now:       time.Now,
	}
}

// Parse decodes a raw NetFlow v9 UDP payload into FlowRecords.
func (p *Parser) Parse(payload []byte) ([]FlowRecord, error) {
	if len(payload) < 24 {
		return nil, fmt.Errorf("payload too short: %d bytes", len(payload))
	}

	hdr := p.parseHeader(payload)
	if hdr.Version != 9 {
		return nil, fmt.Errorf("not NetFlow v9: version=%d", hdr.Version)
	}

	offset := 20
	var records []FlowRecord
	var errs []error

	for i := uint16(0); i < hdr.Count; i++ {
		if offset+4 > len(payload) {
			break
		}

		flowSetID := binary.BigEndian.Uint16(payload[offset : offset+2])
		flowSetLen := int(binary.BigEndian.Uint16(payload[offset+2 : offset+4]))

		if flowSetLen < 4 || offset+flowSetLen > len(payload) {
			break
		}

		fsData := payload[offset+4 : offset+flowSetLen]

		switch {
		case flowSetID == FlowSetTemplate:
			p.parseTemplates(fsData, hdr.SourceID)
		case flowSetID == FlowSetOptionsTmpl:
			// skip options templates
		case flowSetID >= FlowSetMinData:
			recs, err := p.parseDataFlowSet(fsData, flowSetID, hdr)
			if err != nil {
				errs = append(errs, err)
			} else {
				records = append(records, recs...)
			}
		}

		offset += flowSetLen
	}

	if len(errs) > 0 {
		return records, fmt.Errorf("parse errors: %v", errs[0])
	}
	return records, nil
}

func (p *Parser) parseHeader(data []byte) NF9Header {
	return NF9Header{
		Version:   binary.BigEndian.Uint16(data[0:2]),
		Count:     binary.BigEndian.Uint16(data[2:4]),
		SysUptime: binary.BigEndian.Uint32(data[4:8]),
		UnixSecs:  binary.BigEndian.Uint32(data[8:12]),
		UnixNSecs: binary.BigEndian.Uint32(data[12:16]),
		FlowSeq:   binary.BigEndian.Uint32(data[16:20]),
		SourceID:  uint32(binary.BigEndian.Uint16(data[20:22])),
	}
}

func (p *Parser) templateKey(sourceID uint32, templateID uint16) string {
	return fmt.Sprintf("%d:%d", sourceID, templateID)
}

func (p *Parser) parseTemplates(data []byte, sourceID uint32) {
	offset := 0
	for offset+4 <= len(data) {
		tmplID := binary.BigEndian.Uint16(data[offset : offset+2])
		fieldCount := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))
		offset += 4

		if offset+fieldCount*4 > len(data) {
			break
		}

		tmpl := &template{
			Fields: make([]templateField, fieldCount),
		}

		recordLen := 0
		for j := 0; j < fieldCount; j++ {
			ft := binary.BigEndian.Uint16(data[offset : offset+2])
			fl := binary.BigEndian.Uint16(data[offset+2 : offset+4])
			offset += 4
			tmpl.Fields[j] = templateField{Type: ft, Length: fl}
			recordLen += int(fl)
		}
		tmpl.RecordLen = recordLen

		key := p.templateKey(sourceID, tmplID)
		p.mu.Lock()
		p.templates[key] = tmpl
		p.mu.Unlock()
	}
}

func (p *Parser) parseDataFlowSet(data []byte, templateID uint16, hdr NF9Header) ([]FlowRecord, error) {
	key := p.templateKey(hdr.SourceID, templateID)

	p.mu.RLock()
	tmpl, ok := p.templates[key]
	p.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	if tmpl.RecordLen == 0 {
		return nil, nil
	}

	var records []FlowRecord
	offset := 0

	for offset+tmpl.RecordLen <= len(data) {
		recData := data[offset : offset+tmpl.RecordLen]
		rec := p.decodeRecord(recData, tmpl, hdr)
		records = append(records, rec)
		offset += tmpl.RecordLen
	}

	return records, nil
}

func (p *Parser) decodeRecord(data []byte, tmpl *template, hdr NF9Header) FlowRecord {
	rec := FlowRecord{}
	fieldOffset := 0

	for _, f := range tmpl.Fields {
		if fieldOffset+int(f.Length) > len(data) {
			break
		}
		val := data[fieldOffset : fieldOffset+int(f.Length)]

		switch f.Type {
		case NF9_IPV4_SRC_ADDR:
			if f.Length == 4 {
				rec.SrcAddr = net.IP(val).String()
			}
		case NF9_IPV4_DST_ADDR:
			if f.Length == 4 {
				rec.DstAddr = net.IP(val).String()
			}
		case NF9_IPV6_SRC_ADDR:
			if f.Length == 16 {
				rec.SrcAddr = net.IP(val).String()
			}
		case NF9_IPV6_DST_ADDR:
			if f.Length == 16 {
				rec.DstAddr = net.IP(val).String()
			}
		case NF9_L4_SRC_PORT:
			if f.Length >= 2 {
				rec.SrcPort = binary.BigEndian.Uint16(val[:2])
			}
		case NF9_L4_DST_PORT:
			if f.Length >= 2 {
				rec.DstPort = binary.BigEndian.Uint16(val[:2])
			}
		case NF9_PROTOCOL:
			if f.Length >= 1 {
				rec.Protocol = val[len(val)-1]
			}
		case NF9_IN_BYTES:
			rec.Bytes = readUint64(val, f.Length)
		case NF9_IN_PACKETS:
			rec.Packets = readUint64(val, f.Length)
		case NF9_INPUT_SNMP:
			if f.Length >= 2 {
				rec.InputIf = binary.BigEndian.Uint16(val[:2])
			}
		case NF9_OUTPUT_SNMP:
			if f.Length >= 2 {
				rec.OutputIf = binary.BigEndian.Uint16(val[:2])
			}
		case NF9_TOS:
			if f.Length >= 1 {
				rec.TrafficMark = uint32(val[len(val)-1])
			}
		case NF9_FIRST_SWITCHED:
			rel := readUint32(val, f.Length)
			rec.Timestamp = int64(hdr.UnixSecs) - int64((hdr.SysUptime-rel)/1000)
		case NF9_FLOW_START_SEC:
			rec.Timestamp = int64(readUint32(val, f.Length))
		case NF9_FLOW_END_SEC:
			if rec.Timestamp == 0 {
				rec.Timestamp = int64(readUint32(val, f.Length))
			}
		}

		fieldOffset += int(f.Length)
	}

	if rec.Timestamp == 0 {
		rec.Timestamp = int64(hdr.UnixSecs)
	}

	return rec
}

func readUint32(data []byte, length uint16) uint32 {
	switch length {
	case 1:
		return uint32(data[0])
	case 2:
		return uint32(binary.BigEndian.Uint16(data[:2]))
	case 4:
		return binary.BigEndian.Uint32(data[:4])
	default:
		return 0
	}
}

func readUint64(data []byte, length uint16) uint64 {
	switch length {
	case 1:
		return uint64(data[0])
	case 2:
		return uint64(binary.BigEndian.Uint16(data[:2]))
	case 4:
		return uint64(binary.BigEndian.Uint32(data[:4]))
	case 8:
		return binary.BigEndian.Uint64(data[:8])
	default:
		return 0
	}
}
