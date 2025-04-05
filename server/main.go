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

// Estrutura para armazenar os dados do servidor
type Server struct {
	address           string
	listener          net.Listener
	quitch            chan struct{}    // Canal para fechar o servidor
	loggedCars        map[int]net.Conn // Mapa de carros conectados (CarID -> Conexão)
	connectedStations map[int]net.Conn // Mapa de estações conectadas (StationID -> Conexão)

	sessionsMu sync.Mutex // Mutex para evitar race condition e assegurar Mutual Exclusion
}

// Cria um novo servidor
func NewServer(address string) *Server {
	return &Server{
		address:           address,
		quitch:            make(chan struct{}),
		loggedCars:        make(map[int]net.Conn),
		connectedStations: make(map[int]net.Conn),
	}
}

// Inicia o servidor usando TCP
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

// Loop para aceitar conexão de clientes
func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		fmt.Println("Conexão aceita de", conn.RemoteAddr())

		// Iniciar o loop de leitura para processar as requisições
		go s.readLoop(conn)
	}
}

// Possíveis requisições de um cliente tipo Carro
func (s *Server) HandleRequestCar(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	fmt.Println("Req:", message.Req.String(), "do carro:", message.Car.CarID)
	switch message.Req {
	case types.RegisterCar:
		return s.HandleRegisterCar(message, conn)
	case types.UserLogin:
		return s.HandleUserLogin(message, conn)
	case types.ListStations:
		return s.HandleListStations(message, conn)
	case types.ReserveStation:
		return s.HandleReserveStation(message, conn)
	case types.ListActiveStations:
		return s.HandleListActiveStations(message, conn)
	case types.PayRecharge:
		return s.HandlePayRecharge(message, conn)
	case types.GetRecommendedStation:
		return s.HandleGetRecommendedStation(message, conn)
	case types.GetReservedStation:
		return s.HandleGetReservedStation(message, conn)
	case types.RechargeComplete:
		return s.HandleRechargeComplete(message, conn)
	case types.StartRecharge:
		return s.HandleStartRecharge(message, conn)
	case types.CarUpdate:
		return s.HandleCarUpdate(message, conn)
	case types.ExitCar:
		return s.HandleExitCar(message, conn)
	default:
		return types.Message{Status: types.Error, Err: "requisição inválida."}, conn
	}
}

// Possíveis requisições de um cliente tipo Posto
func (s *Server) HandleRequestStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	fmt.Println("Req:", message.Req.String(), "da estação:", message.Station.StationID)
	switch message.Req {
	case types.RegisterStation:
		return s.HandleRegisterStationFromStation(message, conn)
	case types.ListStations:
		return s.HandleListStationsFromStation(message, conn)
	case types.ReserveStation:
		return s.HandleReserveStationFromStation(message, conn)
	case types.SelectStation:
		return s.HandleSelectStationFromStation(message, conn)
	case types.ListActiveStations:
		return s.HandleListActiveStationsFromStation(message, conn)
	case types.PayRecharge:
		return s.HandlePayRechargeFromStation(message, conn)
	case types.RechargeComplete:
		return s.HandleRechargeCompleteFromStation(message, conn)
	case types.StartRecharge:
		return s.HandleStartRechargeFromStation(message, conn)
	case types.StationUpdate:
		return s.HandleStationUpdate(message, conn)
	case types.ExitStation:
		return s.HandleExitStation(message, conn)
	default:
		return types.Message{Status: types.Error, Err: "requisição inválida."}, conn
	}
}

// Loop para ler e decodificar requisições dos clientes
func (s *Server) readLoop(conn net.Conn) {
	responseMessage := types.Message{}
	// Conexão auxiliar para o servidor trocar mensagens entre carro e estação
	var auxiliaryConn net.Conn

	for {
		// Decodificador da conexão do cliente, aguarda até que receba algo
		decoder := json.NewDecoder(conn)
		message := types.Message{}
		// Decodifica o JSON serializado e armazena os dados no objeto responseMessage
		var err = decoder.Decode(&message)

		if err != nil {
			// Verificar se o erro é EOF (cliente desconectou)
			if err == io.EOF {
				fmt.Println("Cliente desconectado:", conn.RemoteAddr())
				break
			}
			fmt.Println("Erro ao decodificar JSON:", err)
			continue
		}

		if message.ClientType == types.CarClientType {
			// auxiliaryConn pode ser uma outra conexão ou a conexão atual(conn)
			responseMessage, auxiliaryConn = s.HandleRequestCar(message, conn)
		} else {
			responseMessage, auxiliaryConn = s.HandleRequestStation(message, conn)
		}

		fmt.Println("Status:", responseMessage.Status.String())
		if responseMessage.Status == types.Error {
			fmt.Println(responseMessage.Err)
		}

		// Enviar a resposta ao cliente
		if auxiliaryConn != nil {
			err = s.SendResponse(auxiliaryConn, responseMessage)
			if err != nil {
				fmt.Println("Erro ao enviar resposta:", err)
				continue
			}
		}
	}
}
func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
