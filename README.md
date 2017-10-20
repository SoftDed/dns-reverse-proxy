# DNS reverse proxy #

[![Build Status](https://api.travis-ci.org/SoftDed/dns-reverse-proxy.png?branch=master)](https://travis-ci.org/SoftDed/dns-reverse-proxy) [![Godoc](https://godoc.org/github.com/SoftDed/dns-reverse-proxy?status.png)](https://godoc.org/github.com/SoftDed/dns-reverse-proxy)

A DNS reverse proxy to route queries to different DNS servers.
To illustrate, imagine an HTTP reverse proxy but for DNS.

It listens on both TCP/UDP IPv4/IPv6 on specified port.
Since the upstream servers will not see the real client IPs but the proxy,
you can specify a list of IPs allowed to transfer (AXFR/IXFR).

Send SIGHUP for reload config file

Example:

    $ go run dns_reverse_proxy.go -bind 127.0.0.1:53 -config /etc/dns-proxy.yaml

# Config #

    defaultserver: 8.8.8.8:53
    transfers:
      example.com.:
        - 111.111.111.111 # how can make AXFR requests
        - 222.222.222.222
    routes:
        example.com.: 192.168.1.2:53 # Request for domain `example.com` send to 192.168.1.2
        example.net.: 192.168.1.3:53

# License #

[Apache License, version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

# Thanks #

- the powerful Go [dns](https://github.com/miekg/dns) library by [Miek Gieben](https://github.com/miekg)
- the powerful Go [dns-reverse-proxy](https://github.com/StalkR/dns-reverse-proxy) library by [StalkR](https://github.com/StalkR)

# Bugs, feature requests, questions #

Create a [new issue](https://github.com/StalkR/dns-reverse-proxy/issues/new).
