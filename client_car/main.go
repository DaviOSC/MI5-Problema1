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
type CarClient struct {
	// Conexão com o servidor
	conn net.Conn
	// Carro do usuário
	car types.Car
	// batteryTicker *time.Ticker
	// syncTicker    *time.Ticker
	// batteryMutex sync.Mutex
}

// Cria e conecta um novo cliente ao servidor
func NewCarClient() *CarClient {
	// Conecta ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Erro ao se conectar com o servidor:", err)
	}

	// Novo cliente
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

// Envia uma mensagem ao servidor,
func (c *CarClient) SendMessage(message types.Message) error {
	// Marca a mensagem com o tipo de cliente que a enviou
	message.ClientType = types.CarClientType
	// Serializa o objeto mensagem estruturado em JSON
	buf, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(buf)

	return err
}

// Lê a resposta do servidor
func (c *CarClient) ReadResponse() (types.Message, error) {
	var responseMessage types.Message
	// Decodificador da conexão do cliente, aguarda até que receba algo
	decoder := json.NewDecoder(c.conn)
	// Decodifica o JSON serializado e armazena os dados no objeto responseMessage
	err := decoder.Decode(&responseMessage)

	return responseMessage, err
}

func main() {

	client := NewCarClient()
	// Canal para notificar quando o programa for interrompido inesperadamente
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// Goroutine para lidar com a interrompção
	go func() {
		<-sigs
		client.SendMessage(types.Message{
			Req: types.ExitCar,
			Car: client.car,
		})
		client.conn.Close()
		os.Exit(0)
	}()

	// Loop para Login
	for {
		car, err := HandleLogin(client.conn)
		if err != nil {
			fmt.Println("Erro ao fazer login:", err)
		} else {
			client.car = car
			break
		}
	}

	// Loop da aplicação
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
9 - Mover carro para (x, y)
10 - Sair`)
		// Ler a escolha do usuário
		fmt.Scanln(&choice)

		// Processar a escolha
		var message types.Message
		var err error
		/*
			Cada handle forma uma mensagem com as informações necessárias
			para o servidor realizar a requisição
		*/
		switch choice {
		case "1":
			message, err = HandleRegisterCar()
			if err != nil {
				fmt.Println("Erro ao registrar o carro:", err)
				// Volta ao inicio do loop em caso de erro
				continue
			}
		case "2":
			message, err = HandleGetRecommendedStation(client.car)
			if err != nil {
				fmt.Println("Erro ao retornar Estação:", err)
				continue
			}
		case "3":
			message, err = HandleReserveStation(client.car)
			if err != nil {
				fmt.Println("Erro ao registrar o carro:", err)
				continue
			}
		case "4":
			message, err = HandleGetReservedStation(client.car)
			if err != nil {
				fmt.Println("Erro ao retornar a estação:", err)
				continue
			}
		case "5":
			message, err = HandleStartRecharge(client.car)
			if err != nil {
				fmt.Println("Erro ao iniciar a recarga:", err)
				continue
			}
		case "6":
			message, err = HandlePayRecharge(client.car)
			if err != nil {
				fmt.Println("Erro ao registrar o carro:", err)
				continue
			}
		case "7":
			message, err = HandleListActiveStations()
			if err != nil {
				fmt.Println("Erro ao listar as estações:", err)
				continue
			}
		case "8":
			/*
				Lista o histórico de pagamentos, não utiliza requisições
				para o servidor
			*/
			for _, payment := range client.car.PaymentHistory {
				fmt.Printf(`
ID do Pagamento: %d
ID do Posto: %d 
ID do Carro: %d
Valor: %d
Data: %s
`, payment.PaymentID, payment.From, payment.To, payment.Value, payment.TimeStamp)
			}
			continue
		case "9":
			x := 0
			y := 0
			fmt.Println("Coordenada X:")
			fmt.Scanln(&x)
			fmt.Println("Coordenada Y:")
			fmt.Scanln(&y)
			MoveCarTo(client, x, y)
		case "10":
			/*
				Antes de sair, envia uma mensagem para que o servidor
				apague qualquer informação em requisições não finalizadas
			*/
			message = types.Message{
				Req: types.ExitCar,
				Car: client.car,
			}
		default:
			fmt.Println("Opção inválida.")
			continue
		}

		// Envia a requisição ao servidor
		err = client.SendMessage(message)
		if err != nil {
			fmt.Println("Erro ao enviar mensagem ao servidor:", err)
			fmt.Println("Tente novamente")
			continue
		}

		// Recebe a resposta do servidor
		responseMessage := types.Message{}
		responseMessage, err = client.ReadResponse()
		if err != nil {
			fmt.Println("Erro receber a resposta do servidor:", err)
			fmt.Println("Tente novamente")
			continue
		}

		// Verifica a requisição da resposta para trata-la de acordo
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
			_, err := HandleListActiveStationsResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			}
		case types.PaymentHistory:
			_, err := HandlePaymentHistoryResponse(responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			}
		case types.GetReservedStation:
			car, err := HandleStartCarMovement(client, responseMessage)
			if err != nil {
				fmt.Println("Erro:", err)
			} else {
				client.car = car
			}
		case types.ExitCar:
			fmt.Println("Saindo")
			return
		default:
			fmt.Println("Requisição da resposta inválida.")
			continue
		}
	}
}
