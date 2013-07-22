package lb

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
)

type proxy struct {
	listen    string
	hostStrip string
	domainAdd string
	registry  *registry
	proxy     *httputil.ReverseProxy
}

func New(listen, hostStrip, domainAdd string) *proxy {
	p := &proxy{
		listen:    listen,
		hostStrip: hostStrip,
		domainAdd: domainAdd,
		registry:  NewRegistry(),
	}

	p.proxy = &httputil.ReverseProxy{
		Director: p.director,
	}
	return p
}

func (p *proxy) ListenAndServe() error {
	http.Handle("/", p.proxy)
	if err := http.ListenAndServe(p.listen, nil); err != nil {
		return err
	}
	return nil
}

func (p *proxy) director(req *http.Request) {
	req.URL.Scheme = "http"
	fields := strings.Split(req.Host, ":")
	name := strings.TrimSuffix(fields[0], p.hostStrip)

	serviceName := fmt.Sprintf("%s.%s", name, p.domainAdd)

	backend := p.registry.getBackend(service(serviceName))
	req.URL.Host = fmt.Sprintf("%s:%d", backend.host, backend.port)
}
