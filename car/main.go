package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
)

var car types.Car = types.Car{}

func main() {
	// Conectar ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		// Se houver um erro ao conectar, o programa será encerrado
		log.Fatal("Erro ao conectar ao servidor:", err)
	}
	// Garantir que a conexão será fechada ao final da execução
	defer conn.Close()
	for {
		var username, password string
		fmt.Print("Digite o usuário: ")
		fmt.Scanln(&username)
		fmt.Print("Digite a senha: ")
		fmt.Scanln(&password)

		loginMessage, err := HandleLogin(username, password)
		if err != nil {
			log.Fatal("Erro ao preparar a mensagem de login:", err)
		}

		buf, err := json.Marshal(loginMessage)
		if err != nil {
			log.Fatal("Erro ao serializar a mensagem de login:", err)
		}
		_, err = conn.Write(buf)
		if err != nil {
			log.Fatal("Erro ao enviar a mensagem de login:", err)
		}

		var response types.ResponseMessage
		decoder := json.NewDecoder(conn)
		err = decoder.Decode(&response)
		if err != nil {
			log.Fatal("Erro ao decodificar a resposta do servidor:", err)
		}

		if response.Status != "Login bem-sucedido" {
			fmt.Println("Erro no login:", response.Status)			
		}else {
			fmt.Println(response.Status)
			break
		}

	}
	for {
		// Menu para o cliente escolher o que fazer
		var choice string
		fmt.Println(`Escolha uma opção:
1 - Registrar um Carro
2 - Pedir Recomendação de Estação
3 - Reservar Estação
4 - Recarregar Carro
5 - Pagar Recarga
6 - Sair`)
		// Ler a escolha do usuário
		fmt.Scanln(&choice)

		// Processar a escolha
		var message types.Message
		var err error
		switch choice {
		case "1":
			message, car, err = HandleRegisterCar()
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "2":
			message, err = HandleGetRecommendedStation(car)
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "3":
			message, err = HandleReserveStation(car)
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "4":
			message, err = HandleRechargeCar(car)
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "5":
			message, err = HandlePayRecharge(car)
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "6":
			fmt.Println("Saindo...")
		default:
			fmt.Println("Opção inválida.")
			continue
		}

		// Enviar a mensagem ao servidor
		buf, err := json.Marshal(message)
		if err != nil {
			// Se houver um erro ao serializar a mensagem, o programa será encerrado
			log.Fatal("Erro ao serializar a mensagem:", err)
		}

		_, err = conn.Write(buf)
		if err != nil {
			// Se houver um erro ao enviar os dados, o programa será encerrado
			log.Fatal("Erro ao enviar dados:", err)
		}

		// Receber a resposta do servidor
		var response types.ResponseMessage
		decoder := json.NewDecoder(conn)
		err = decoder.Decode(&response)
		if err != nil {
			// Se houver um erro ao decodificar a resposta, o programa será encerrado
			log.Fatal("Erro ao decodificar a resposta:", err)
		}

		// Mostrar a resposta
		fmt.Println("Status:", response.Status)
		if response.Data != nil {
			// Mostrar os dados recebidos, se houver
			fmt.Printf("Dados recebidos: %+v\n", response.Data)
		}
	}
}
