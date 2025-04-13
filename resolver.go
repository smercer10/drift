package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	defaultTimeout    = 2 * time.Second
	wildcardTestCount = 3
)

// DNSResolver provides DNS resolution capabilities.
type DNSResolver interface {
	LookupHost(fqdn string) ([]string, error)
	DetectWildcard(apexDomain string) (bool, []string)
}

// Resolver implements the DNSResolver interface.
type Resolver struct {
	client  *net.Resolver
	rng     *rand.Rand
	rngMu   sync.Mutex
	timeout time.Duration
}

// ResolverOption defines functional options for the Resolver.
type ResolverOption func(*Resolver)

// WithTimeout sets a custom timeout for DNS queries.
func WithTimeout(timeout time.Duration) ResolverOption {
	return func(r *Resolver) {
		r.timeout = timeout
	}
}

// NewResolver creates a new Resolver with the specified nameserver.
func NewResolver(nameserver string, opts ...ResolverOption) *Resolver {
	seed := time.Now().UnixNano() // Deterministic but "random enough".

	r := &Resolver{
		rng:     rand.New(rand.NewSource(seed)),
		timeout: defaultTimeout,
	}

	for _, opt := range opts {
		opt(r)
	}

	dialer := &net.Dialer{Timeout: r.timeout}
	r.client = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, fmt.Sprintf("%s:53", nameserver)) // 53 is the default DNS port.
		},
	}

	return r
}

// LookupHost resolves a FQDN to IP addresses.
func (r *Resolver) LookupHost(fqdn string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.client.LookupHost(ctx, fqdn)
}

// generateRandomFQDN generates a random FQDN for wildcard testing.
func (r *Resolver) generateRandomFQDN(apexDomain string, length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)

	r.rngMu.Lock()
	for i := range result {
		result[i] = charset[r.rng.Intn(len(charset))]
	}
	r.rngMu.Unlock()

	return fmt.Sprintf("%s.%s", string(result), apexDomain)
}

// DetectWildcard checks if an apex domain uses wildcard DNS records by testing multiple random FQDNs.
func (r *Resolver) DetectWildcard(apexDomain string) (bool, []string) {
	// Get apex domain IPs for comparison.
	apexIPs, err := r.LookupHost(apexDomain)
	if err != nil || len(apexIPs) == 0 {
		return false, nil
	}

	wildcardIPs := make(map[string]int)
	for range wildcardTestCount {
		testFQDN := r.generateRandomFQDN(apexDomain, 10)
		ips, err := r.LookupHost(testFQDN)

		if err == nil && len(ips) > 0 {
			// Count occurrence of each IP.
			for _, ip := range ips {
				wildcardIPs[ip]++
			}
		}
	}

	// If any IP appeared in all tests, it's likely a wildcard.
	confirmedWildcardIPs := make([]string, 0, len(wildcardIPs))
	for ip, count := range wildcardIPs {
		if count == wildcardTestCount {
			confirmedWildcardIPs = append(confirmedWildcardIPs, ip)
		}
	}

	return len(confirmedWildcardIPs) > 0, confirmedWildcardIPs
}
