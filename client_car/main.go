package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
	"sync"
)

type CarClient struct {
	conn net.Conn
	car  types.Car
	// batteryTicker *time.Ticker
	// syncTicker    *time.Ticker
	batteryMutex sync.Mutex
}

func NewCarClient() *CarClient {
	// Conectar ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Erro ao conectar ao servidor:", err)
	}

	// Criar client com os novos campos
	c := &CarClient{
		conn: conn,
		car:  types.Car{},
		// batteryTicker: time.NewTicker(5 * time.Second),  // Decremento a cada 5s
		// syncTicker:    time.NewTicker(60 * time.Second), // Sincronização a cada 60s
		// batteryMutex:  sync.Mutex{},
	}

	// Iniciar monitoramento de bateria em goroutine separada
	// go c.batteryMonitor()

	return c
}

func (c *CarClient) SendMessage(message types.Message) error {
	// Enviar a mensagem ao servidor
	buf, err := json.Marshal(message)
	if err != nil {
		// Se houver um erro ao serializar a mensagem, o programa será encerrado
		log.Fatal("Erro ao serializar a mensagem:", err)
	}

	_, err = c.conn.Write(buf)
	if err != nil {
		// Se houver um erro ao enviar os dados, o programa será encerrado
		log.Fatal("Erro ao enviar dados:", err)
	}

	return err
}

func (c *CarClient) ReadResponse() (types.Message, error) {
	// Receber a resposta do servidor
	var responseMessage types.Message
	decoder := json.NewDecoder(c.conn)
	err := decoder.Decode(&responseMessage)
	if err != nil {
		// Se houver um erro ao decodificar a resposta, o programa será encerrado
		log.Fatal("Erro ao decodificar a resposta:", err)
	}

	return responseMessage, err
}

func main() {

	client := NewCarClient()
	defer client.conn.Close()

	for {
		car, err := HandleLogin(client.conn)
		if err != nil {
			fmt.Println("Erro ao fazer login:", err)
		} else {
			client.car = car
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
4 - Ir para Estação
5 - Iniciar a recarga
6 - Pagar Recarga
7 - Listar Estações
8 - Listar Histórico de Recargas
9 - Sair`)
		// Ler a escolha do usuário
		fmt.Scanln(&choice)

		// Processar a escolha
		var message types.Message
		var err error
		switch choice {
		case "1":
			message, err = HandleRegisterCar()
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "2":
			fmt.Print(client.car.CarID)
			message, err = HandleGetRecommendedStation(client.car)
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "3":
			message, err = HandleReserveStation(client.car)
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "4":
			message, err = HandleGetReservedStation(client.car)
			if err != nil {
				log.Fatal("Erro ao retornar a estação:", err)
			}
		case "5":
			message, err = HandleStartRecharge(client.car)
			if err != nil {
				log.Fatal("Erro ao iniciar a recarga:", err)
			}
		case "6":
			message, err = HandlePayRecharge(client.car)
			if err != nil {
				log.Fatal("Erro ao registrar o carro:", err)
			}
		case "7":
			message, err = HandleListActiveStations()
			if err != nil {
				log.Fatal("Erro ao listar as estações:", err)
			}
		case "8":
			message, err = HandlePaymentHistory(client.car)
			if err != nil {
				log.Fatal("Erro ao listar o histórico de pagamentos:", err)
			}
		case "9":
			log.Fatal("Saindo...")
		default:
			fmt.Println("Opção inválida.")
			continue
		}

		err = client.SendMessage(message)
		if err != nil {
			log.Fatal("Erro ao enviar mensagem:", err)
		}

		responseMessage := types.Message{}
		responseMessage, err = client.ReadResponse()
		if err != nil {
			log.Fatal("Erro ao receber resposta:", err)
		}

		switch responseMessage.Req {
		case types.RegisterCar:
			_, err = HandleRegisterCarResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			}
		case types.GetRecommendedStation:
			car, err := HandleGetRecommendedStationResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.ReserveStation:
			car, err := HandleReserveStationResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.RechargeComplete:
			car, err := HandleRechargeCompleteResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.StartRecharge:
			car, err := HandleStartRechargeResponse(client, responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.PayRecharge:
			car, err := HandlePayRechargeResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.ListActiveStations:
			car, err := HandleListActiveStationsResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.PaymentHistory:
			car, err := HandlePaymentHistoryResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.GetReservedStation:
			if responseMessage.Status == types.Success {
				client.car = responseMessage.Car
			}
			startCarMovement(client, responseMessage.Station)
		default:
			fmt.Println("Requisição da resposta inválida.")
			continue
		}
	}
}
