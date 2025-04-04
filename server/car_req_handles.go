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

func (s *Server) HandlePayRecharge(message types.Message, conn net.Conn) (types.Message, net.Conn) {
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
		fmt.Println("Estação não encontrada, na requisição GetRecommendedStation")
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
func (s *Server) HandleStartRecharge(message types.Message, conn net.Conn) (types.Message, net.Conn) {
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
