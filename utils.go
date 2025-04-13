package main

import (
	"bufio"
	"os"
	"strings"
)

// LoadWordlist reads a file containing a line-separated list of potential subdomains.
func LoadWordlist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var subdomains []string
	scanner := bufio.NewScanner(file)

	// Preallocate slice to reduce reallocations.
	subdomains = make([]string, 0, 1000)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			subdomains = append(subdomains, line)
		}
	}

	return subdomains, scanner.Err()
}
