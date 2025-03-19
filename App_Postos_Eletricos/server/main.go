package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"net"
)

type Server struct {
	address  string // Porta do tcp
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
	}
	defer ln.Close()
	s.listener = ln

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
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	for {
		decoder := json.NewDecoder(conn)
		car := types.Car{}
		decoder.Decode(&car)

		fmt.Println("Recebido:", car.CarID, car.CoordX, car.CoordY, car.BatteryLevel, car.RecomendedStation, car.ReservedStation)
	}
}

func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
