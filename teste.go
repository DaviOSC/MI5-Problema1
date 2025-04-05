package main

import (
	"fmt"
	"slices"
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
	list := make([]int, 1)
	for i := range 100 {
		list = slices.Insert(list, len(list)-1, i)
	}

	for i := range 100 {
		fmt.Print(list[i], " ")
	}
}
