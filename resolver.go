package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type Resolver struct {
	client *net.Resolver
	rng    *rand.Rand
}

const defaultTimeout = 2 * time.Second

// NewResolver creates a new Resolver instance with the specified nameserver.
func NewResolver(nameserver string) *Resolver {
	randSource := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(randSource)

	dialer := &net.Dialer{Timeout: defaultTimeout}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, nameserver)
		},
	}

	return &Resolver{client: resolver, rng: rng}
}

// LookupHost resolves a subdomain using the configured resolver.
func (r *Resolver) LookupHost(subdomain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return r.client.LookupHost(ctx, subdomain)
}

// DetectWildcard checks if a domain uses wildcard DNS records.
func (r *Resolver) DetectWildcard(domain string) (bool, string) {
	testSubdomain := fmt.Sprintf("test-%d.%s", r.rng.Intn(1000), domain)

	ips, err := r.LookupHost(testSubdomain)
	if err != nil || len(ips) == 0 {
		return false, ""
	}

	apexIps, err := r.LookupHost(domain)
	if err != nil || len(apexIps) == 0 {
		return false, ""
	}

	// If the subdomain resolves to the same IP as the apex domain, it's a wildcard.
	if ips[0] == apexIps[0] {
		return true, ips[0]
	}

	return false, ""
}
