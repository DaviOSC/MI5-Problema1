package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for {
	}
}
