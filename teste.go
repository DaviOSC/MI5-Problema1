package main

import (
	"fmt"
	"sync"
)

var mu sync.Mutex

func teste() {
	mu.Lock()
	fmt.Println("1")
	teste2()
	mu.Unlock()
}

func teste2() {
	mu.Lock()
	fmt.Println("2")
	mu.Unlock()
}

func main() {
	fmt.Printf("%d%%", 5)
}
