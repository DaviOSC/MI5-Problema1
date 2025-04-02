package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/types"
	"net"
	"sync"
)

type Server struct {
	address           string
	listener          net.Listener
	quitch            chan struct{}
	loggedCars        map[int]net.Conn // Novo: mapa de carros logados
	connectedStations map[int]net.Conn // Mapa de estações conectadas (StationID -> Conexão)

	sessionsMu sync.Mutex
}

func NewServer(address string) *Server {
	return &Server{
		address:           address,
		quitch:            make(chan struct{}),
		loggedCars:        make(map[int]net.Conn),
		connectedStations: make(map[int]net.Conn),
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

		// Iniciar o loop de leitura para processar mensagens
		go s.readLoop(conn)
	}
}

func (s *Server) HandleRequest(message types.Message, conn net.Conn) types.Message {
	switch message.Req {
	case types.RegisterCar:
		return s.HandleRegisterCar(message)
	case types.UserLogin:
		return s.HandleUserLogin(message, conn)
	case types.CarUpdate:
		return s.HandleCarUpdate(message)
	case types.GetRecommendedStation:
		return s.HandleGetRecommendedStation(message)
	case types.GetReservedStation:
		return s.HandleGetReservedStation(message)
	case types.ReserveStation:
		return s.HandleReserveStation(message)
	case types.StartRecharge:
		return s.HandleStartRecharge(message)
	case types.RechargeComplete:
		return s.HandleRechargeComplete(message)
	case types.PayRecharge:
		return s.HandlePayRecharge(message)
	case types.ListActiveStations:
		return s.HandleListActiveStations(message)
	case types.RegisterStation:
		return s.HandleRegisterStation(message)
	case types.ListStations:
		return s.HandleListStations(message)
	case types.SelectStation:
		return s.HandleSelectStation(message, conn)
	default:
		return types.Message{Status: types.Error, Err: "requisição inválida."}
	}
}

func (s *Server) readLoop(conn net.Conn) {
	responseMessage := types.Message{}

	for {
		decoder := json.NewDecoder(conn)
		message := types.Message{}
		var err = decoder.Decode(&message)
		if err != nil {
			// Verificar se o erro é EOF (cliente desconectou)
			if err == io.EOF {
				fmt.Println("Cliente desconectado:", conn.RemoteAddr())
				break
			}
			// Outros erros são exibidos
			fmt.Println("Erro ao decodificar JSON:", err)
			break
		}
		responseMessage = s.HandleRequest(message, conn)

		// Enviar a resposta ao cliente
		err = s.SendResponse(conn, responseMessage)
		if err != nil {
			fmt.Println("Erro ao enviar resposta:", err)
			break
		}
	}
	defer s.CloseConnection(responseMessage)
}
func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
