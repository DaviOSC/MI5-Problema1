package main

import (
	"fmt"
	"main/types"
	"net"
)

/*
Registra um novo posto e salva no arquivo JSON
*/
func (s *Server) HandleRegisterStationFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	station := message.Station
	responseMessage := types.Message{Req: types.RegisterStation}

	err := s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintln("Erro ao salvar a estação no arquivo JSON")
	} else {
		responseMessage.Status = types.Success
		responseMessage.Station = station
	}
	return responseMessage, conn
}

/*
Requisição acionada quando um cliente posto desconecta do servidor
Qualquer requisição anterior não concluida é finalizada
*/
func (s *Server) HandleExitStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.ExitStation}

	s.sessionsMu.Lock()
	delete(s.connectedStations, message.Station.StationID)
	s.sessionsMu.Unlock()

	err := s.saveStationToFile(message.Station)
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintln("Erro em ExitStation:", err)
	}
	responseMessage.Status = types.Success
	return responseMessage, conn
}

/*
Retorna a lista de todas as estações salvas no servidor
*/
func (s *Server) HandleListStationsFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	stations, err := s.listStationsFromFile()
	responseMessage := types.Message{Req: types.ListStations}

	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintln("Erro ao listar estações:", err)
	} else {
		responseMessage.Status = types.Success
		responseMessage.StationList = stations
	}

	return responseMessage, conn
}

/*
Verifica o estado da requisição, salva o carro e a estação em JSON
em caso de sucesso, e repassa para o caro que fez a requisição original
*/
func (s *Server) HandleReserveStationFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.ReserveStation, Status: types.Success}
	if message.Status == types.Error {
		responseMessage.Status = types.Error
		responseMessage.Err = message.Err
		return responseMessage, conn
	}
	s.saveCarToFile(message.Car)
	s.saveStationToFile(message.Station)

	return message, s.loggedCars[message.Car.CarID]
}

/*
Adiciona a estação, selecionada por um cliente posto, à lista de estações conectadas
*/
func (s *Server) HandleSelectStationFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.SelectStation}
	// Bloquear o acesso ao mapa de estações conectadas
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	fmt.Println(message.Station.StationID)
	stationID := message.Station.StationID

	// Verificar se a estação já está conectada
	if _, exists := s.connectedStations[stationID]; exists {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintf("estação com ID %d já está conectada", stationID)
		return responseMessage, conn
	}
	// Adicionar a estação ao mapa de estações conectadas
	//message.Station.Conn = conn.RemoteAddr().String() // Salvar o IP da conexão
	s.connectedStations[stationID] = conn

	responseMessage.Status = types.Success
	responseMessage.Station = message.Station

	return responseMessage, conn
}

/*
Retorna a lista das estações conectadas atualmente no servidor
*/
func (s *Server) HandleListActiveStationsFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.ListActiveStations, Status: types.Success}

	// Bloquear o acesso ao mapa de estações conectadas para evitar condições de corrida
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	// Criar uma lista de estações conectadas
	var activeStations []types.Station
	for ID, _ := range s.connectedStations {
		station, err := s.getStationFromFile(ID)
		if err != nil {
			// não é necessário retornar um erro ao usuário caso não seja possível buscar
			// alguma das estações
			continue
		}
		activeStations = append(activeStations, station)
	}

	// Adicionar a lista de estações conectadas à resposta
	responseMessage.StationList = activeStations
	return responseMessage, conn
}

/*
Verifica o estado da requisição, salva o carro e a estação em JSON
em caso de sucesso, e repassa para o caro que fez a requisição original
*/
func (s *Server) HandlePayRechargeFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	if message.Status == types.Success {
		s.saveCarToFile(message.Car)
		s.saveStationToFile(message.Station)
	}
	return message, s.loggedCars[message.Car.CarID]
}

/*
Verifica o estado da requisição, salva o carro e a estação em JSON
em caso de sucesso, e repassa para o caro que fez a requisição original
*/
func (s *Server) HandleRechargeCompleteFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	if message.Status == types.Success {
		s.saveCarToFile(message.Car)
		s.saveStationToFile(message.Station)
	}
	return message, s.loggedCars[message.Car.CarID]
}

/*
Verifica o estado da requisição, salva o carro e a estação em JSON
em caso de sucesso, e repassa para o caro que fez a requisição original
*/
func (s *Server) HandleStartRechargeFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	if message.Status == types.Success {
		s.saveCarToFile(message.Car)
		s.saveStationToFile(message.Station)
	}
	return message, s.loggedCars[message.Car.CarID]
}

/*
Recebe dados de uma estação para serem salvos em JSON,
não retorna respostas para o cliente que fez a requisição
*/
func (s *Server) HandleStationUpdate(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.StationUpdate, Status: types.Success}
	station := message.Station
	err := s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintln("Em HandleStationUpdate:", err)
	}
	// nil para não enviar a mensagem para o cliente
	return responseMessage, nil
}
