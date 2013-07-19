package lb

import (
	"github.com/miekg/dns"
	"github.com/soundcloud/go-dns-resolver/resolv"
	"log"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	urlf = "http://%s:%d"
	serverErrorStart = 500
	serverErrorEnd   = 599
)

var (
	cacheTime = 10 * time.Second
)

type service string

type backend struct {
	host string
	port uint16
}

func (b *backend) url() string {
	return fmt.Sprintf(urlf, b.host, b.port)
}

func (b *backend) alive() bool {
	url := b.url()

	res, err := http.Head(url)
	if err != nil {
		log.Printf("%s dead: %s", url, err)
		return false
	}
	defer res.Body.Close()
	if res.StatusCode >= serverErrorStart && res.StatusCode <= serverErrorEnd {
		log.Printf("%s dead: %s", url, res.Status)
		return false
	}
	return true
}
// pool
type pool struct {
	time     time.Time
	backends []*backend
}

func (p *pool) fresh() bool {
	return time.Now().Sub(p.time) > cacheTime
}

func (p *pool) randomBackend() *backend {
	return p.backends[rand.Intn(len(p.backends))]
}

// registry
func NewRegistry() *registry {
	return &registry{pools: make(map[service]*pool)}
}

type registry struct {
	pools map[service]*pool
}

func (r *registry) getPool(name service) *pool {
	pool, ok := r.pools[name]
	if ok {
		log.Printf("%s: cache hit", name)
		if pool.fresh() {
			return pool
		}
		log.Printf("%s: cache too old", name)
	}
	return r.resolvPool(name)
}

func (r *registry) resolvPool(name service) *pool {
	msg, err := resolv.Lookup(dns.TypeSRV, string(name))
	if err != nil {
		log.Fatalf("Couldn't lookup %s: %s", name, err)
	}

	backends := []*backend{}
	for _, r := range msg.Answer {
		record := r.(*dns.SRV)
		b := &backend{host: record.Target, port: record.Port}
		if b.alive() {
			backends = append(backends, b)
		}
	}

	p := &pool{
		time:     time.Now(),
		backends: backends,
	}
	r.pools[name] = p
	return p
}

func (r *registry) getBackend(name service) *backend {
	return r.getPool(name).randomBackend()
}
