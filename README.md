wizbal
======

Magically autoconfigured HTTP load balancer written in Go.

wizbal takes the Host header, strips the suffix given by `-host` and
appends the suffix given by `-domain`. It asks DNS for SRV records
with that resulting name and uses the returns host/port pairs as
backend pool for load balancing.

This is just a proof of concept with a simple health check and
on-demand (but cached) dns resolution, put together in a few hours
as a [Hacker Time](http://backstage.soundcloud.com/2011/12/stop-hacker-time/)
project. Thanks to [Grégoire](http://soundcloud.com/greguar) for
suggesting the name.


# Usage

1. Expose your backends as SRV records under some domain: `srv.example.com`
2. Setup a wildcard DNS entry for serving client requests, pointing to wizbal: `app.example.com`
3. Fire up wizbal: `./wizbal -domain=srv.example.com -host=app.example.com`

Let say you have some backends under my-service.srv.example.com. Now you can access http://my-service.app.example.com:8080


# TODO
- Cache SRV names and refresh regularly to avoid doing on-demand resolving
