package main

import (
	"encoding/json"
	"fmt"
)

type estrutura struct {
	Teste  int `json:"teste"`
	Teste2 any `json:"teste2"`
}

type outra_estrutura struct {
	Texto string `json:"texto"`
}

func teste(arg estrutura) {
	fmt.Println(arg.Teste2.(map[string]interface{})["texto"])
}

func main() {
	oe := outra_estrutura{Texto: "Tomar no cu"}
	a := estrutura{Teste: 1, Teste2: oe}
	b, _ := json.Marshal(a)
	var c estrutura = estrutura{}
	json.Unmarshal(b, &c)
	teste(c)
}
