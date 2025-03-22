package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
)

type ServerResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

func main() {
	// Conectar ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Erro ao conectar ao servidor:", err)
	}
	defer conn.Close()

	for {
		// Menu para o cliente escolher o que fazer
		var choice string
		fmt.Println("Escolha uma opção:")
		fmt.Println("1 - Registrar uma Estação")
		fmt.Println("2 - Consultar informações de uma Estação")
		fmt.Println("3 - Listar todas as Estações")
		fmt.Print("Escolha: ")
		fmt.Scanln(&choice)

		var message struct {
			Choice string          `json:"choice"`
			ID     int             `json:"id"`
			Data   json.RawMessage `json:"data"`
		}

		// Processar a escolha
		switch choice {
		case "1":
			message.Choice = "register_station"
			var station types.Station
			fmt.Print("Digite o ID da estação: ")
			fmt.Scanln(&station.StationID)
			fmt.Print("Digite a coordenada X da estação: ")
			fmt.Scanln(&station.CoordX)
			fmt.Print("Digite a coordenada Y da estação: ")
			fmt.Scanln(&station.CoordY)
			message.ID = station.StationID
			message.Data, _ = json.Marshal(station)

		case "2":
			message.Choice = "get_station"
			fmt.Print("Digite o ID da estação para consulta: ")
			fmt.Scanln(&message.ID)

		case "3":
			message.Choice = "list_stations"

		default:
			fmt.Println("Opção inválida.")
			return
		}

		// Enviar a mensagem ao servidor
		buf, err := json.Marshal(message)
		if err != nil {
			log.Fatal("Erro ao serializar a mensagem:", err)
		}

		_, err = conn.Write(buf)
		if err != nil {
			log.Fatal("Erro ao enviar dados:", err)
		}

		// Receber a resposta do servidor
		var response ServerResponse
		decoder := json.NewDecoder(conn)
		err = decoder.Decode(&response)
		if err != nil {
			log.Fatal("Erro ao decodificar a resposta:", err)
		}

		// Mostrar a resposta
		fmt.Println("Status:", response.Status)
		if response.Data != nil {
			fmt.Printf("Dados recebidos: %+v\n", response.Data)
		}
	}
}
