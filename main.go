package main

import "fmt"

func main() {
	words, err := LoadWordlist("wordlist.txt")
	if err != nil {
		fmt.Println("Error loading wordlist:", err)
		return
	}

	resolver := NewResolver("8.8.8.8")

	subdomainEnum := NewSubdomainEnumerator("example.com", words, resolver, 50)

	subdomainEnum.Start()
}
