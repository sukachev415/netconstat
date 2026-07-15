package collector

import (
	"log"
	"net"
	"sync"

	"github.com/sukachev415/netconstat/internal/parser"
)

// FlowBatch is a batch of parsed records from one UDP packet.
type FlowBatch struct {
	Records []parser.FlowRecord
	Err     error
}

// Collector listens for NetFlow v9 UDP packets and emits parsed records.
type Collector struct {
	addr    string
	bufSize int
	parser  *parser.Parser

	// Stats
	mu          sync.RWMutex
	PacketsRecv uint64
	PacketsErr  uint64
	FlowsRecv   uint64
}

func New(addr string, bufSize int) *Collector {
	return &Collector{
		addr:    addr,
		bufSize: bufSize,
		parser:  parser.NewParser(),
	}
}

// Start begins listening for UDP packets and sends parsed flow batches to the returned channel.
func (c *Collector) Start() (<-chan FlowBatch, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", c.addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	ch := make(chan FlowBatch, 10000)

	go c.readLoop(conn, ch)

	log.Printf("[collector] listening on %s (buffer=%d)", c.addr, c.bufSize)
	return ch, nil
}

func (c *Collector) readLoop(conn *net.UDPConn, ch chan<- FlowBatch) {
	defer conn.Close()
	buf := make([]byte, c.bufSize)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			ch <- FlowBatch{Err: err}
			continue
		}

		c.mu.Lock()
		c.PacketsRecv++
		c.mu.Unlock()

		payload := make([]byte, n)
		copy(payload, buf[:n])

		records, parseErr := c.parser.Parse(payload)
		if parseErr != nil {
			c.mu.Lock()
			c.PacketsErr++
			c.mu.Unlock()
			log.Printf("[collector] parse error (pkt %d bytes): %v", n, parseErr)
			if len(records) == 0 {
				continue
			}
		}

		if len(records) > 0 {
			log.Printf("[collector] parsed %d flows from %d byte packet", len(records), n)
		}

		c.mu.Lock()
		c.FlowsRecv += uint64(len(records))
		c.mu.Unlock()

		ch <- FlowBatch{Records: records}
	}
}

// Stats returns current collector statistics.
func (c *Collector) Stats() (packetsRecv, packetsErr, flowsRecv uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.PacketsRecv, c.PacketsErr, c.FlowsRecv
}
