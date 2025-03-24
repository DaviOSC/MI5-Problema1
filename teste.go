package main

import (
	"encoding/json"
	"fmt"
)

type estrutura struct {
	Teste  int             `json:"teste"`
	Teste2 outra_estrutura `json:"teste2"`
}

type outra_estrutura struct {
	Texto string `json:"texto"`
}

func teste(arg estrutura) {
	fmt.Println(arg.Teste2.Texto)
}

func main() {
	a := estrutura{Teste: 1, Teste2: outra_estrutura{Texto: "teste"}}
	b, _ := json.Marshal(a)
	var c estrutura = estrutura{}
	json.Unmarshal(b, &c)
	fmt.Println(c.Teste2 == outra_estrutura{} || c.Teste2.Texto == "")
}
