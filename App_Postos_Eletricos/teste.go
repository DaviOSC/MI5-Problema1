package main

import (
	"fmt"
)

func main() {
	s := ""
	_, err := fmt.Scanln(&s)
	if err != nil {
		fmt.Println("Erro ao ler dados:", err)
	}
	fmt.Println(s)
}
