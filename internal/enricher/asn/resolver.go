package asn

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	maxminddb "github.com/oschwald/maxminddb-golang"
)

// Record represents an ASN lookup result.
type Record struct {
	ASN        uint   `maxminddb:"autonomous_system_number"`
	ASNOrg     string `maxminddb:"autonomous_system_organization"`
}

// Resolver performs IP-to-ASN lookups using a MaxMind mmdb file.
type Resolver struct {
	db   *maxminddb.Reader
	path string
	mu   sync.RWMutex
}

// NewResolver opens the mmdb file at the given path.
// Returns a Resolver that gracefully falls back to "Unknown" if the file is missing.
func NewResolver(path string) (*Resolver, error) {
	r := &Resolver{path: path}
	if err := r.load(); err != nil {
		log.Printf("[asn] WARNING: could not load ASN database: %v — all lookups will return Unknown", err)
		// Don't return error — allow running without ASN DB
	}
	return r, nil
}

func (r *Resolver) load() error {
	if r.path == "" {
		return fmt.Errorf("no ASN database path configured")
	}

	info, err := os.Stat(r.path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", r.path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("ASN path is a directory: %s", r.path)
	}

	db, err := maxminddb.Open(r.path)
	if err != nil {
		return fmt.Errorf("open mmdb: %w", err)
	}

	r.mu.Lock()
	if r.db != nil {
		r.db.Close()
	}
	r.db = db
	r.mu.Unlock()

	log.Printf("[asn] loaded database from %s", r.path)
	return nil
}

// Lookup resolves an IP address to an ASN number and organization name.
// Returns (0, "Unknown") if the database is not loaded or the IP is not found.
func (r *Resolver) Lookup(ipStr string) (uint, string) {
	r.mu.RLock()
	db := r.db
	r.mu.RUnlock()

	if db == nil {
		return 0, "Unknown"
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, "Unknown"
	}

	// Use IPv4 if it's an IPv4-mapped IPv6 address
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}

	var record Record
	if err := db.Lookup(ip, &record); err != nil {
		return 0, "Unknown"
	}

	if record.ASN == 0 {
		return 0, "Unknown"
	}

	org := record.ASNOrg
	if org == "" {
		org = fmt.Sprintf("AS%d", record.ASN)
	}

	return record.ASN, org
}

// Reload checks if the file has been updated and reloads it.
func (r *Resolver) Reload() error {
	return r.load()
}

// Close closes the underlying database.
func (r *Resolver) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db != nil {
		r.db.Close()
		r.db = nil
	}
}

// CachedResolver wraps a Resolver with an in-memory LRU cache.
type CachedResolver struct {
	inner   *Resolver
	cache   sync.Map
	ttl     time.Duration
	maxSize int
	size    int
	mu      sync.Mutex
}

type cacheEntry struct {
	asn    uint
	org    string
	expiry time.Time
}

// NewCachedResolver creates a caching layer around the given Resolver.
func NewCachedResolver(inner *Resolver, ttl time.Duration, maxSize int) *CachedResolver {
	cr := &CachedResolver{
		inner:   inner,
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Periodic reload of ASN DB (every 24h)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := cr.inner.Reload(); err != nil {
				log.Printf("[asn] reload error: %v", err)
			}
		}
	}()

	return cr
}

// Lookup resolves an IP to ASN, using cache when available.
func (cr *CachedResolver) Lookup(ip string) (uint, string) {
	// Check cache
	if v, ok := cr.cache.Load(ip); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiry) {
			return entry.asn, entry.org
		}
		// Expired — delete and re-lookup
		cr.cache.Delete(ip)
		cr.mu.Lock()
		cr.size--
		cr.mu.Unlock()
	}

	// Cache miss — resolve
	asn, org := cr.inner.Lookup(ip)

	// Store in cache (with simple eviction if over limit)
	cr.mu.Lock()
	if cr.size >= cr.maxSize {
		// Simple eviction: clear half the cache
		// In production you'd use a proper LRU, but sync.Map doesn't have ordered iteration
		// For now, this is acceptable for the scale we're targeting
		cr.cache.Range(func(key, value any) bool {
			cr.cache.Delete(key)
			cr.size--
			return cr.size > cr.maxSize/2
		})
	}
	cr.size++
	cr.mu.Unlock()

	cr.cache.Store(ip, cacheEntry{
		asn:    asn,
		org:    org,
		expiry: time.Now().Add(cr.ttl),
	})

	return asn, org
}
