package main

import (
	"fmt"
	"sync"
	"time"
)

var mu sync.Mutex

func teste() {
	fmt.Println("teste")
	mu.Lock()
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Acessei 1")
	}
	mu.Unlock()
}

func teste2() {
	fmt.Println("teste2")
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Assecei 2")
	}
}

func main() {
	go teste()
	go teste2()
	time.Sleep(500000000000)
}
