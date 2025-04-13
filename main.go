package main

import "fmt"

func main() {
	wordlistPath := "./wordlist.txt"
	words, err := LoadWordlist(wordlistPath)
	if err != nil {
		panic(err)
	}

	fmt.Println(words)
}
