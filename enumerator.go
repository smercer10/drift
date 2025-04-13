package main

import (
	"fmt"
	"sync"
)

type SubdomainEnumerator struct {
	domain   string
	words    []string
	resolver *Resolver
	limiter  chan struct{}
}

func NewSubdomainEnumerator(domain string, words []string, resolver *Resolver, concurrencyLimit int) *SubdomainEnumerator {
	return &SubdomainEnumerator{
		domain:   domain,
		words:    words,
		resolver: resolver,
		limiter:  make(chan struct{}, concurrencyLimit),
	}
}

func (e *SubdomainEnumerator) Start() {
	var wg sync.WaitGroup

	for _, word := range e.words {
		e.limiter <- struct{}{}
		wg.Add(1)

		go func(subdomain string) {
			defer wg.Done()
			defer func() { <-e.limiter }()

			fullDomain := fmt.Sprintf("%s.%s", subdomain, e.domain)
			ips, err := e.resolver.LookupHost(fullDomain)
			if err == nil && len(ips) > 0 {
				fmt.Printf("Found subdomain %s -> %s\n", fullDomain, ips[0])
			}
			if err != nil {
				fmt.Printf("Error resolving %s: %v\n", fullDomain, err)
			}
		}(word)
	}

	wg.Wait()
}
