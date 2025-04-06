package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// Estrutura para respresentar os dados do cliente
type StationClient struct {
	// Conexão com o servidor
	conn net.Conn
	// Posto de Recarga
	station types.Station
}

// Cria e conecta um novo cliente ao servidor
func NewStationClient() *StationClient {
	// Conecta ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Erro ao conectar ao servidor:", err)
	}

	// retorna um novo posto
	return &StationClient{
		conn: conn,
	}
}

// IDs sequenciais para cada pagamento
var paymentIDs int = 0

// Envia uma mensagem ao servidor
func (c *StationClient) SendMessage(message types.Message) error {
	// Marca a mensagem com o tipo de cliente que a enviou
	message.ClientType = types.StationClientType
	// Serializa o objeto mensagem estruturado em JSON
	buf, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(buf)

	return err
}

// Lê a resposta do servidor
func (c *StationClient) ReadResponse() (types.Message, error) {
	var responseMessage types.Message
	// Decodificador da conexão do cliente, aguarda até que receba algo
	decoder := json.NewDecoder(c.conn)
	// Decodifica o JSON serializado e armazena os dados no objeto responseMessage
	err := decoder.Decode(&responseMessage)

	return responseMessage, err
}

func main() {
	client := NewStationClient()

	// Canal para notificar quando o programa for interrompido inesperadamente
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// Goroutine para lidar com a interrompção
	go func() {
		<-sigs
		client.SendMessage(types.Message{
			Req:     types.ExitStation,
			Station: client.station,
		})
		client.conn.Close()
		os.Exit(0)
	}()

	stationChosen := false
	for {
		var choice string
		var message types.Message
		var err error

		// Menu para o cliente escolher o que fazer
		if !stationChosen {
			fmt.Println(`Escolha uma opção:
1 - Registrar uma Estação
2 - Escolher uma Estação
3 - Sair`)
			fmt.Scanln(&choice)
			switch choice {
			// Registrar Estação
			case "1":
				err = client.HandleRegisterStation()
				if err != nil {
					fmt.Println("Erro:", err)
				}
			// Escolher Estação
			case "2":
				station, err := client.HandleListAndChooseStation()
				if err != nil {
					fmt.Println("Erro:", err)
				} else {
					client.station = station
					stationChosen = true
				}
			// Sair
			case "3":
				fmt.Println("Saindo...")
				message = client.HandleExit()
				err = client.SendMessage(message)
				if err != nil {
					fmt.Println("Erro:", err)
				}
				return
			default:
				fmt.Println("Opção inválida.")
				continue
			}
			// stationChonsen = true
		} else {
			fmt.Printf("=============================================================\n")
			fmt.Println("Estação iniciada, para sair digite '1'.")
			// Rotina para receber requisições do servidor
			go client.readLoop(client.conn)
			for {
				fmt.Scanln(&choice)
				// Sair
				if choice == "1" {
					fmt.Println("Saindo...")
					message = client.HandleExit()
					err = client.SendMessage(message)
					if err != nil {
						fmt.Println("Erro:", err)
						continue
					}
					return
				}
			}
		}
	}
}

// Cada Handle trata uma requisição recebida do servidor
func (c *StationClient) HandleRequest(message types.Message) types.Message {
	switch message.Req {
	case types.StationUpdate:
		return c.HandleGetStationInfo(message)
	case types.ReserveStation:
		return c.HandleReserveStation(message)
	case types.PayRecharge:
		return c.HandlePayRecharge(message)
	case types.RechargeComplete:
		return c.HandleRechargeComplete(message)
	case types.StartRecharge:
		return c.HandleStartRecharge(message)
	}
	return message
}

// Loop para receber as requisições
func (c *StationClient) readLoop(conn net.Conn) {
	var responseMessage types.Message

	for {
		// Decodificador da conexão do cliente, aguarda até que receba algo
		decoder := json.NewDecoder(conn)
		message := types.Message{}
		// Decodifica o JSON serializado e armazena os dados no objeto responseMessage
		var err = decoder.Decode(&message)
		if err != nil {
			fmt.Println("Erro ao decodificar JSON:", err)
			break
		}
		responseMessage = c.HandleRequest(message)
		// Enviar a resposta ao servidor
		err = c.SendMessage(responseMessage)
		if err != nil {
			fmt.Println("Erro ao enviar resposta:", err)
			break
		}
	}

}
