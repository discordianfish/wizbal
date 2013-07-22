package lb

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/soundcloud/go-dns-resolver/resolv"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	urlf             = "http://%s:%d"
	serverErrorStart = 500
	serverErrorEnd   = 599
)

var (
	cacheTime        = 10 * time.Second
	responseRedirect = errors.New("http-redirect")
)

type service string

type backend struct {
	host string
	port uint16
}

func (b *backend) url() string {
	return fmt.Sprintf(urlf, b.host, b.port)
}

func (b *backend) alive(client *http.Client) bool {
	backendUrl := b.url()

	res, err := client.Head(backendUrl)
	if err != nil {
		e, ok := err.(*url.Error)
		if ok && e.Err == responseRedirect {
			log.Printf("just an redirect, ignoring")
		} else {
			log.Printf("%s dead: %s", backendUrl, err)
			return false
		}
	}
	defer res.Body.Close()
	if res.StatusCode >= serverErrorStart && res.StatusCode <= serverErrorEnd {
		log.Printf("%s dead: %s", backendUrl, res.Status)
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
	return time.Now().Sub(p.time) < cacheTime
}

func (p *pool) randomBackend() *backend {
	n := len(p.backends)
	switch n {
	case 0:
		log.Printf("No backends found")
		return nil

	case 1:
		log.Printf("Only one backend found")
		return p.backends[0]

	default:
		return p.backends[rand.Intn(n)]
	}
}

// registry
func NewRegistry() *registry {
	return &registry{
		client: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return responseRedirect
			},
		},
		pools: make(map[service]*pool),
	}
}

type registry struct {
	client *http.Client
	pools  map[service]*pool
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
	log.Printf("Resp: %v", msg)
	backends := []*backend{}
	for _, rr := range msg.Answer {
		record := rr.(*dns.SRV)
		b := &backend{host: record.Target, port: record.Port}
		if b.alive(r.client) {
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
