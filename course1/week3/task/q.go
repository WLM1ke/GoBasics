package main

import "fmt"

var m = map[string]bool{
	"Bell Labs": true,
	"Google":    true,
}

func main() {
	m["q"] = true
	fmt.Println(m["q"])
}
