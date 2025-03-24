package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
)

type Server struct {
	address  string // Porta do TCP
	listener net.Listener
	quitch   chan struct{}
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		quitch:  make(chan struct{}),
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
	for {
		decoder := json.NewDecoder(conn)
		message := types.Message{}
		var err = decoder.Decode(&message)
		if err != nil {
			fmt.Println("Erro ao decodificar JSON:", err)
			break
		}

		responseMessage := types.Message{}

		switch message.Req {

		case types.RegisterCar:
			responseMessage = HandleRegisterCar(message)

		case types.RegisterStation:
			responseMessage = HandleRegisterStation(message)

		case types.UserLogin:
			responseMessage = HandleUserLogin(message)

		case types.ListStations:
			responseMessage = HandleListStations(message)

		case types.ReserveStation:
			responseMessage = HandleReserveStation(message)

		case types.RechargeComplete:
			responseMessage = HandleRechargeComplete(message)

		case types.PayRecharge:
			responseMessage = HandlePayRecharge(message)

		case types.GetRecommendedStation:
			responseMessage = HandleGetRecommendedStation(message)

		case types.PaymentHistory:
			responseMessage = HandlePaymentHistory(message)
		default:
			responseMessage.Status = types.Error
			fmt.Println("Requisição inválida")
		}

		err = SendResponse(conn, responseMessage)
		if err != nil {
			log.Fatal("Erro ao enviar resposta:", err)
		}
	}
}

func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
