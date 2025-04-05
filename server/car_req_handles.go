package main

import (
	"fmt"
	"main/types"
	"net"
)

func (s *Server) HandleRegisterCar(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	car := message.Car
	responseMessage := types.Message{Req: types.RegisterCar}

	err := s.saveCarToFile(car)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Println("Erro ao salvar o carro")
	} else {
		responseMessage.Status = types.Success
		responseMessage.Car = car
		fmt.Printf("Carro com id %d registrado com sucesso", car.CarID)
	}
	return responseMessage, conn
}

func (s *Server) HandleUserLogin(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	loginData := message.Car
	responseMessage := types.Message{Req: types.UserLogin}

	cars, err := s.listCarsFromFile()
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Println("Erro ao acessar os dados dos carros")
		return responseMessage, conn
	}

	valid := false
	for _, car := range cars {
		if car.User == loginData.User && car.Password == loginData.Password {
			valid = true
			responseMessage.Status = types.Success
			responseMessage.Car = car
			fmt.Printf("%d/n", responseMessage.Car.CarID)
			s.sessionsMu.Lock()
			s.loggedCars[car.CarID] = conn
			s.sessionsMu.Unlock()

			fmt.Println("Login bem-sucedido.")
			break
		}
	}

	if !valid {
		responseMessage.Status = types.Error
		fmt.Println("Usuário ou senha inválidos")
	}
	return responseMessage, conn
}

func (s *Server) HandleListStations(message types.Message, conn net.Conn) (types.Message, net.Conn) {
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

func (s *Server) HandleReserveStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.ReserveStation, Status: types.Success}
	var err error
	station := types.Station{}
	stationID := message.Car.RecomendedStation

	// Verifica se o carro possue uma estação recomendada
	if stationID == 0 {
		station, err = s.getBestStation(message.Car)
		if err != nil {
			responseMessage.Status = types.Error
			responseMessage.Err = "Não foi possível encontrar uma estação recomendada"
			fmt.Printf("Estação não encontrada, na requisição ReserveStation, para o carro de id %d\n", message.Car.CarID)
			return responseMessage, conn
		}
		stationID = station.StationID
	}

	responseMessage.Car = message.Car

	// Enviar a mensagem para a estação concluir a requisição da reserva
	return responseMessage, s.connectedStations[stationID]
}

func (s *Server) HandleListActiveStations(message types.Message, conn net.Conn) (types.Message, net.Conn) {
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

/*
Verifica se o carro possui uma estação reservada
Então envia mensagem para a estação reservada
Se não, retorna erro
*/
func (s *Server) HandlePayRecharge(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.PayRecharge, Status: types.Success}
	if message.Car.ReservedStation == 0 {
		responseMessage.Status = types.Error
		responseMessage.Err = "reserve a estação antes de pagar"
		return responseMessage, conn
	}

	return message, s.connectedStations[message.Car.ReservedStation]
}

// func HandlePaymentHistory(message types.Message, conn net.Conn) (types.Message, net.Conn) {
// 	car := message.Car
// 	responseMessage := types.Message{Req: types.PaymentHistory}

// 	payments := []types.Payment{}
// 	responseMessage.
// 	return responseMessage

// }

func (s *Server) HandleGetRecommendedStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	car := message.Car
	responseMessage := types.Message{Req: types.GetRecommendedStation}
	fmt.Printf("Requisição GetRecommendedStation para o carro de id %d\n", car.CarID)
	station, err := s.getBestStation(car)
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = err.Error()
		fmt.Println("Estação não encontrada, na requisição GetRecommendedStation")
		return responseMessage, conn

	} else {
		responseMessage.Status = types.Success
		responseMessage.Station = station
		car.RecomendedStation = station.StationID
		responseMessage.Car = car
		err = s.saveCarToFile(car)
		if err != nil {
			responseMessage.Status = types.Success
		}
		fmt.Printf("Estação com id %d encontrada", station.StationID)
	}

	return responseMessage, conn
}

func (s *Server) HandleGetReservedStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.GetReservedStation}
	car, err := s.getCarFromFile(message.Car.CarID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao obter o carro com ID %d: %v\n", message.Car.CarID, err)
		return responseMessage, conn
	}
	if car.ReservedStation == 0 {
		responseMessage.Status = types.Error
		responseMessage.Err = "Carro não possui uma estação reservada"
		fmt.Printf("Carro %d não possui uma estação reservada.\n", car.CarID)

		return responseMessage, conn
	}
	station, err := s.getStationFromFile(car.ReservedStation)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao obter a estação reservada com ID %d: %v\n", car.ReservedStation, err)
		return responseMessage, conn
	}

	fmt.Printf("Retornando a estação %d para o carro %d", station.StationID, car.CarID)
	responseMessage.Status = types.Success
	responseMessage.Car = car
	responseMessage.Station = station
	return responseMessage, conn
}

func (s *Server) HandleRechargeComplete(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	if message.Car.ReservedStation == 0 {
		return types.Message{
			Status: types.Error,
			Err:    "Carro não possui uma estação reservada"}, conn
	}
	return message, s.connectedStations[message.Car.ReservedStation]
}

func (s *Server) HandleStartRecharge(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	if message.Car.ReservedStation == 0 {
		return types.Message{
			Status: types.Error,
			Err:    "Carro não possui uma estação reservada"}, conn
	}
	return message, s.connectedStations[message.Car.ReservedStation]
}

func (s *Server) HandleCarUpdate(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	if _, exists := s.loggedCars[message.Car.CarID]; exists {

		// Persiste no arquivo diretamente
		if err := s.saveCarToFile(message.Car); err != nil {
			fmt.Printf("[ERRO] Falha ao salvar bateria do carro %d: %v\n", message.Car.CarID, err)
			return types.Message{Status: types.Error}, conn
		}

		fmt.Printf("[SYNC] Bateria do carro %d atualizada para %d%%\n", message.Car.CarID, message.Car.BatteryLevel)
	}

	return types.Message{Status: types.Success}, conn
}
