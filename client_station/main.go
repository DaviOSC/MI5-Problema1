package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
)

type StationClient struct {
	conn    net.Conn
	station types.Station
}

func (c *StationClient) SendMessage(message types.Message) error {
	// Enviar a mensagem ao servidor
	buf, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Erro ao serializar a mensagem:", err)
	}

	_, err = c.conn.Write(buf)
	if err != nil {
		log.Fatal("Erro ao enviar dados:", err)
	}

	return err
}

func (c *StationClient) ReadResponse() (types.Message, error) {
	// Receber a resposta do servidor
	var responseMessage types.Message
	decoder := json.NewDecoder(c.conn)
	err := decoder.Decode(&responseMessage)
	if err != nil {
		log.Fatal("Erro ao decodificar a resposta:", err)
	}

	return responseMessage, err
}

func main() {
	client := NewStationClient()
	defer client.conn.Close()
	stationChosen := false

	for {
		// Menu para o cliente escolher o que fazer
		var choice string
		if !stationChosen {
			fmt.Println(`Escolha uma opção:
1 - Registrar uma Estação
2 - Escolher uma Estação
3 - Sair`)
		} else {
			fmt.Println("Estação iniciada, para sair digite '1'.")
		}
		fmt.Scanln(&choice)

		// Processar a escolha
		var message types.Message
		var err error
		if !stationChosen {
			switch choice {
			case "1":
				message, err = RegisterStation()
				if err != nil {
					fmt.Println("Erro:", err)
					continue
				}
				err = client.SendMessage(message)
				if err != nil {
					fmt.Println("Erro ao enviar mensagem:", err)
					continue
				}

				// Receber a resposta do servidor
				responseMessage, err := client.ReadResponse()
				if err != nil {
					fmt.Println("Erro ao receber resposta:", err)
					continue
				}
				_, err = RegisterStationResponse(responseMessage)
				if err != nil {
					fmt.Println("Erro:", err)
				}
			case "2":
				// Criar a mensagem para listar as estações disponíveis
				message, err = ChooseStation(client.station)
				if err != nil {
					fmt.Println("Erro ao criar mensagem:", err)
					continue
				}
			
				// Enviar a mensagem para listar estações
				err = client.SendMessage(message)
				if err != nil {
					fmt.Println("Erro ao enviar mensagem:", err)
					continue
				}
			
				// Receber a resposta do servidor
				responseMessage, err := client.ReadResponse()
				if err != nil {
					fmt.Println("Erro ao receber resposta:", err)
					continue
				}
			
				// Processar a resposta e permitir que o usuário escolha uma estação
				station, err := ChooseStationResponse(responseMessage)
				if err != nil {
					fmt.Println("Erro:", err)
					continue
				}
			
				// Criar a mensagem para enviar a estação escolhida ao servidor
				message, err = SelectStation(station)
				if err != nil {
					fmt.Println("Erro ao criar mensagem de seleção:", err)
					continue
				}

				err = client.SendMessage(message)
				if err != nil {
					fmt.Println("Erro ao enviar mensagem:", err)
					continue
				}
				responseMessage, err = client.ReadResponse()
				if err != nil {
					fmt.Println("Erro ao receber resposta:", err)
					continue
				}
			
				// Receber a resposta do servidor
				station, err = SelectStationResponse(responseMessage)
				if err != nil {
					fmt.Println("Erro:", err)
				} else {
					client.station = station
					stationChosen = true
				}
			case "3":
				fmt.Println("Saindo...")
				return
			default:
				fmt.Println("Opção inválida.")
				continue
			}
		} else {
			if choice == "1" {
				fmt.Println("Saindo...")
				return
			} else {
				fmt.Println("Opção inválida.")
				continue
			}
		}
	}
}
