package main

import (
	"fmt"
	"main/types"
	"net"
)

func (s *Server) HandleRegisterStationFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	station := message.Station
	responseMessage := types.Message{Req: types.RegisterStation}

	err := s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Println("Erro ao salvar a estação")
	} else {
		responseMessage.Status = types.Success
		responseMessage.Station = station
		fmt.Printf("Estação com id %d registrada com sucesso", station.StationID)
	}
	return responseMessage, conn
}

func (s *Server) HandleListStationsFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	stations, err := s.listStationsFromFile()
	responseMessage := types.Message{Req: types.ListStations}

	if err != nil {
		responseMessage.Status = types.Error
		fmt.Println("Erro ao listar estações:", err)
	} else {
		responseMessage.Status = types.Success
		responseMessage.StationList = stations
		fmt.Println("Estações listadas com sucesso.")
	}

	return responseMessage, conn
}

func (s *Server) HandleReserveStationFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.ReserveStation, Status: types.Success}
	if message.Status == types.Error {
		responseMessage.Status = types.Error
		responseMessage.Err = message.Err
		fmt.Printf("Erro ao reservar a estação: %s\n", message.Err)
		return responseMessage, conn
	}
	s.saveCarToFile(message.Car)
	s.saveStationToFile(message.Station)

	return message, s.loggedCars[message.Car.CarID]
}

func (s *Server) HandleSelectStationFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.SelectStation}
	fmt.Printf("Requisição SelectStation para a estação de id %d\n", message.Station.StationID)
	// Bloquear o acesso ao mapa de estações conectadas
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	fmt.Println(message.Station.StationID)
	stationID := message.Station.StationID

	// Verificar se a estação já está conectada
	if _, exists := s.connectedStations[stationID]; exists {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintf("estação com ID %d já está conectada", stationID)
		fmt.Printf("Erro: Estação com ID %d já está conectada.\n", stationID)
		return responseMessage, conn
	}
	// Adicionar a estação ao mapa de estações conectadas
	//message.Station.Conn = conn.RemoteAddr().String() // Salvar o IP da conexão
	s.connectedStations[stationID] = conn

	responseMessage.Status = types.Success
	responseMessage.Station = message.Station
	fmt.Printf("Estação com ID %d conectada com sucesso. IP: %s\n", responseMessage.Station.StationID, conn.RemoteAddr().String())
	fmt.Printf("Estação %d adicionada com sucesso. IP: %s\n", stationID, conn.RemoteAddr().String())
	return responseMessage, conn
}

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
			fmt.Println("Em HandleListActiveStations:", err)
		}
		activeStations = append(activeStations, station)
	}

	// Adicionar a lista de estações conectadas à resposta
	responseMessage.StationList = activeStations
	fmt.Printf("Estações ativas listadas com sucesso: %d estações conectadas.\n", len(activeStations))
	return responseMessage, conn
}

func (s *Server) HandlePayRechargeFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.PayRecharge, Status: types.Success}
	stationID := message.Car.ReservedStation
	station, err := s.getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada, na requisição PayRecharge, para a estação de id %d\n", stationID)
		return responseMessage, conn
	}
	if message.Car.ReservedStation == 0 {
		responseMessage.Status = types.Error
		responseMessage.Err = "reserve a estação antes de pagar"
		return responseMessage, conn
	}

	if station.InUseBy == 0 && station.CarList[0] == message.Car.CarID {
		station.InUseBy = message.Car.CarID
		station.CarsWaiting -= 1
		station.CarList = station.CarList[1:]
		err := s.saveStationToFile(station)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Erro ao salvar a estação, na requisição PayRecharge, para a estação de id %d\n", stationID)
			return responseMessage, conn
		}
		// TODO simular pagamento
		responseMessage.Car = message.Car
		err = s.saveCarToFile(message.Car)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Erro ao salvar o carro, na requisição PayRecharge, para o carro de id %d\n", message.Car.CarID)
			return responseMessage, conn
		}
	} else {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não disponível, na requisição PayRecharge, para a estação de id %d\n", stationID)
	}

	return responseMessage, conn
}

func (s *Server) HandleRechargeCompleteFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.RechargeComplete, Status: types.Success}
	car, stationID := message.Car, message.Car.ReservedStation

	// Obter a estação do arquivo
	station, err := s.getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada para o ID %d\n", stationID)
		return responseMessage, conn
	}

	// Validar se a estação está em uso pelo carro correto
	if station.InUseBy != car.CarID {
		responseMessage.Status = types.Error
		fmt.Printf("Estação %d não está em uso pelo carro %d\n", stationID, car.CarID)
		return responseMessage, conn
	}
	// Atualizar os dados do carro e da estação
	car.ReservedStation, car.RecomendedStation = 0, 0
	station.InUseBy = 0
	car.BatteryLevel = 100

	err = s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar a estação %d\n", station.StationID)
		return responseMessage, conn
	}

	err = s.saveCarToFile(car)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar o carro %d\n", car.CarID)
		return responseMessage, conn
	}

	fmt.Printf("Recarga concluída para o carro %d na estação %d.\n", car.CarID, station.StationID)
	responseMessage.Car = car
	return responseMessage, conn
}
func (s *Server) HandleStartRechargeFromStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.StartRecharge, Status: types.Success}
	car, stationID := message.Car, message.Car.ReservedStation

	// Obter a estação do arquivo
	station, err := s.getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada para o ID %d\n", stationID)
		return responseMessage, conn
	}

	// Marcar a estação como em uso pelo carro
	station.InUseBy = car.CarID
	err = s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar a estação %d\n", stationID)
		return responseMessage, conn
	}

	fmt.Printf("Recarga iniciada para o carro %d na estação %d.\n", car.CarID, station.StationID)
	//TODO o cliente ainda consegue enviar mensagens no terminal no periodo. elas são processadas em sequencia apos o tempo
	responseMessage.Car = car
	return responseMessage, conn
}

func (s *Server) HandleStationUpdate(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.StationUpdate, Status: types.Success}
	station := message.Station
	err := s.saveStationToFile(station)
	if err != nil {
		fmt.Println("Em HandleStationUpdate:", err)
		responseMessage.Status = types.Error
	}
	// nil para não enviar a mensagem para o cliente
	return responseMessage, nil
}
