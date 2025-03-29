package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
	"sync"
)

type Server struct {
	address    string
	listener   net.Listener
	quitch     chan struct{}
	loggedCars map[int]types.Car // Novo: mapa de carros logados
	sessionsMu sync.Mutex
}

func NewServer(address string) *Server {
	return &Server{
		address:    address,
		quitch:     make(chan struct{}),
		loggedCars: make(map[int]types.Car),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return err
	}
	defer ln.Close()
	s.listener = ln
	fmt.Println("Servidor iniciado na porta", s.address)

	go s.acceptLoop()
	<-s.quitch
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		fmt.Println("Conexão aceita de", conn.RemoteAddr())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	car := types.Car{}
	for {
		//@DaviOSC usei pra escrever o JSON no terminal
		// decoder := json.NewDecoder(conn)
		// message := types.Message{}
		// err := decoder.Decode(&message)
		// if err != nil {
		//     fmt.Println("Erro ao decodificar JSON:", err)

		// }

		// formattedJSON, err := json.MarshalIndent(message, "", "  ")
		// if err != nil {
		//     fmt.Println("Erro ao formatar JSON:", err)

		// }
		// fmt.Println("JSON recebido formatado:")
		// fmt.Println(string(formattedJSON))
		// responseMessage := types.Message{}

		decoder := json.NewDecoder(conn)
		message := types.Message{}
		err := decoder.Decode(&message)
		if err != nil {
			fmt.Println("Erro ao decodificar JSON:", err)
			break
		}
		car = message.Car
		responseMessage := types.Message{}

		switch message.Req {
		case types.RegisterCar:
			responseMessage = s.HandleRegisterCar(message)

		case types.RegisterStation:
			responseMessage = s.HandleRegisterStation(message)

		case types.UserLogin:
			responseMessage = s.HandleUserLogin(message)

		case types.ListStations:
			responseMessage = s.HandleListStations(message)

		case types.ReserveStation:
			responseMessage = s.HandleReserveStation(message)

		case types.RechargeComplete:
			responseMessage = s.HandleRechargeComplete(message)

		case types.StartRecharge:
			responseMessage = s.HandleStartRecharge(message)

		case types.PayRecharge:
			responseMessage = s.HandlePayRecharge(message)

		case types.GetRecommendedStation:
			responseMessage = s.HandleGetRecommendedStation(message)

		case types.BatterySync:
			responseMessage = s.HandleBatterySync(message)
		// case types.UserLogout:  // Novo caso
		// 	responseMessage = s.HandleUserLogout(message)
		// case types.PaymentHistory:
		// 	responseMessage = HandlePaymentHistory(message)
		default:
			responseMessage.Status = types.Error
			fmt.Println("Requisição inválida")
		}

		err = s.SendResponse(conn, responseMessage)
		if err != nil {
			log.Fatal("Erro ao enviar resposta:", err)
		}
		car = responseMessage.Car
	}
	defer s.CloseConnection(car)
}

func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
