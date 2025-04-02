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
	loggedCars        map[int]types.Car     // Novo: mapa de carros logados
	connectedStations map[int]types.Station // Mapa de estações conectadas (StationID -> Conexão)

	sessionsMu sync.Mutex
}

func NewServer(address string) *Server {
	return &Server{
		address:           address,
		quitch:            make(chan struct{}),
		loggedCars:        make(map[int]types.Car),
		connectedStations: make(map[int]types.Station),
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

		// Identificar o tipo de cliente com base na requisição
		switch message.Req {
		case types.RegisterCar, types.UserLogin, types.CarUpdate, types.GetRecommendedStation,
			types.GetReservedStation, types.ReserveStation, types.StartRecharge,
			types.RechargeComplete, types.PayRecharge, types.ListActiveStations:
			fmt.Println("Requisição recebida de um cliente_car.")
			responseMessage = s.handleCarRequest(message)

		case types.RegisterStation, types.ListStations:
			fmt.Println("Requisição recebida de um cliente_station.")
			responseMessage = s.handleStationRequest(message, conn)

		default:
			responseMessage.Status = types.Error
			fmt.Println("Requisição inválida.")
		}

		// Enviar a resposta ao cliente
		err = s.SendResponse(conn, responseMessage)
		if err != nil {
			fmt.Println("Erro ao enviar resposta:", err)
			break
		}
	}
	fmt.Print(responseMessage)
	defer s.CloseConnection(responseMessage)
}
func (s *Server) handleCarRequest(message types.Message) types.Message {
	switch message.Req {
	case types.RegisterCar:
		return s.HandleRegisterCar(message)
	case types.UserLogin:
		return s.HandleUserLogin(message)
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
	default:
		return types.Message{Status: types.Error, Err: fmt.Errorf("requisição inválida para cliente_car")}
	}
}
func (s *Server) handleStationRequest(message types.Message, conn net.Conn) types.Message {
	switch message.Req {
	case types.RegisterStation:
		return s.HandleRegisterStation(message)
	case types.ListStations:
		return s.HandleListStations(message)
	case types.SelectStation: // Nova requisição
		return s.HandleSelectStation(message, conn)
	default:
		return types.Message{Status: types.Error, Err: fmt.Errorf("requisição inválida para cliente_station")}
	}
}
func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
