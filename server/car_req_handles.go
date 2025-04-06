package main

import (
	"fmt"
	"main/types"
	"maps"
	"net"
	"slices"
)

/*
Registra um carro, recebe as informações basicas do cliente,
salva em um arquivo e retorna o mesmo carro em carro de sucesso
*/
func (s *Server) HandleRegisterCar(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	car := message.Car
	responseMessage := types.Message{Req: types.RegisterCar}

	err := s.saveCarToFile(car)
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = "Erro ao salvar o carro"
	} else {
		responseMessage.Status = types.Success
		responseMessage.Car = car
	}
	return responseMessage, conn
}

/*
Realiza o login de um usuário,
*/
func (s *Server) HandleUserLogin(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	loginData := message.Car
	responseMessage := types.Message{Req: types.UserLogin}

	// Recebe a lista de todos os carros já registrados no servidor
	cars, err := s.listCarsFromFile()
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = "Erro ao acessar os dados dos carros"
		return responseMessage, conn
	}

	// Compara com as informações recebidas na mensagem
	valid := false
	for _, car := range cars {
		if car.User == loginData.User && car.Password == loginData.Password {
			valid = true
			responseMessage.Status = types.Success
			responseMessage.Car = car

			s.sessionsMu.Lock()
			s.loggedCars[car.CarID] = conn
			s.sessionsMu.Unlock()

			break
		}
	}

	// Caso nenhuma das informações coincidam
	if !valid {
		responseMessage.Status = types.Error
		responseMessage.Err = "Usuário ou senha inválidos"
	}
	return responseMessage, conn
}

/*
Retorna a lista de todas as estações salvas no servidor
*/
func (s *Server) HandleListStations(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	stations, err := s.listStationsFromFile()
	responseMessage := types.Message{Req: types.ListStations}

	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = "Erro ao listar estações"
	} else {
		responseMessage.Status = types.Success
		responseMessage.StationList = stations
	}

	return responseMessage, conn
}

/*
Reservar um posto para um cliente carro
*/
func (s *Server) HandleReserveStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.ReserveStation, Status: types.Success}
	var err error
	station := types.Station{}
	stationID := message.Car.RecomendedStation

	// Verifica se o carro possui um posto recomendado
	if stationID == 0 {
		// Caso não, um posto é imediatamente buscado
		station, err = s.getBestStation(message.Car)
		if err != nil {
			responseMessage.Status = types.Error
			responseMessage.Err = "Não foi possível encontrar uma estação recomendada"
			return responseMessage, conn
		}
		stationID = station.StationID
	}

	responseMessage.Car = message.Car
	// Envia a mensagem, com as informações do carro,
	// para a estação que se quer reservar
	return responseMessage, s.connectedStations[stationID]
}

/*
Retorna uma lista, com os postos ativos atualmente, para o cliente
*/
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
			continue
		}
		activeStations = append(activeStations, station)
	}

	// Adicionar a lista de estações conectadas à resposta
	responseMessage.StationList = activeStations
	return responseMessage, conn
}

/*
Verifica se o carro possui uma estação reservada
Então envia mensagem para a estação reservada
Se não, retorna erro
*/
func (s *Server) HandlePayRecharge(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{
		Req:    types.PayRecharge,
		Status: types.Success}
	if message.Car.ReservedStation == 0 {
		responseMessage.Status = types.Error
		responseMessage.Err = "reserve a estação antes de pagar"
		return responseMessage, conn
	}

	return message, s.connectedStations[message.Car.ReservedStation]
}

/*
Recebe informações de um cliente carro e retorna o posto mais
viavel para recarga
*/
func (s *Server) HandleGetRecommendedStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	car := message.Car
	responseMessage := types.Message{Req: types.GetRecommendedStation}

	station, err := s.getBestStation(car)
	if err != nil {
		// estação
		responseMessage.Status = types.Error
		responseMessage.Err = "nenhuma estação foi encontrada"
		return responseMessage, conn

	} else {
		// estação encontrada
		responseMessage.Status = types.Success
		responseMessage.Station = station
		car.RecomendedStation = station.StationID
		responseMessage.Car = car
		s.saveCarToFile(car)
	}

	return responseMessage, conn
}

/*
Retorna dados do posto reservado pelo carro que enviou esta requisição
*/
func (s *Server) HandleGetReservedStation(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	responseMessage := types.Message{Req: types.GetReservedStation}
	car, err := s.getCarFromFile(message.Car.CarID)
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintf("Erro ao obter o carro com ID %d: %v\n", message.Car.CarID, err)
		return responseMessage, conn
	}
	if car.ReservedStation == 0 {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintf("Carro %d não possui uma estação reservada.\n", car.CarID)

		return responseMessage, conn
	}
	station, err := s.getStationFromFile(car.ReservedStation)
	if err != nil {
		responseMessage.Status = types.Error
		responseMessage.Err = fmt.Sprintf("Erro ao obter a estação reservada com ID %d: %v\n", car.ReservedStation, err)
		return responseMessage, conn
	}

	responseMessage.Status = types.Success
	responseMessage.Car = car
	responseMessage.Station = station
	return responseMessage, conn
}

/*
Indica ao posto que o carro completou a recarga
*/
func (s *Server) HandleRechargeComplete(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	if message.Car.ReservedStation == 0 {
		return types.Message{
			Status: types.Error,
			Err:    "Carro não possui uma estação reservada"}, conn
	}
	return message, s.connectedStations[message.Car.ReservedStation]
}

/*
Inicia a recarga no posto
*/
func (s *Server) HandleStartRecharge(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	fmt.Println("Aqui, reservedStation: ", message.Car.ReservedStation)
	if message.Car.ReservedStation == 0 {
		return types.Message{
			Status: types.Error,
			Err:    "Carro não possui uma estação reservada"}, conn
	}
	return message, s.connectedStations[message.Car.ReservedStation]
}

/*
Atualiza as informações do carro no servidor
*/
func (s *Server) HandleCarUpdate(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	if _, exists := s.loggedCars[message.Car.CarID]; exists {

		// Persiste no arquivo diretamente
		if err := s.saveCarToFile(message.Car); err != nil {
			return types.Message{
				Status: types.Error,
				Err:    fmt.Sprintf("Erro ao salvar bateria do carro %d: %v\n", message.Car.CarID, err),
			}, conn
		}
	}

	return types.Message{Status: types.Success}, conn
}

/*
Requisição acionada quando um cliente carro desconecta do servidor
Qualquer requisição anterior não concluida é finalizada
*/
func (s *Server) HandleExitCar(message types.Message, conn net.Conn) (types.Message, net.Conn) {
	reservedStationID := message.Car.ReservedStation
	// Caso o carro possua um posto reservado
	if reservedStationID != 0 {

		// Busca esse posto nos arquivos
		reservedStation, err := s.getStationFromFile(reservedStationID)
		if err != nil {
			message.Status = types.Error
			message.Err = fmt.Sprintln("Erro em HandleExitCar:", err)
		}

		// Retira esse carro da fila do posto
		reservedStation.CarsWaiting -= 1
		index := slices.Index(reservedStation.CarList, message.Car.CarID)
		if index >= 0 {
			reservedStation.CarList = slices.Delete(reservedStation.CarList, index, index+1)
		}

		// Retira o carro caso estivesse no meio de uma recarga
		if reservedStation.InUseBy == message.Car.CarID {
			reservedStation.InUseBy = 0
		}

		message.Car.RecomendedStation = 0
		message.Car.ReservedStation = 0

		// Retira o carro da lista de carros conectados
		maps.DeleteFunc(s.loggedCars, func(k int, v net.Conn) bool {
			return k == message.Car.CarID
		})

		// Salva as novas informações do carro e do posto
		s.saveCarToFile(message.Car)
		s.saveStationToFile(reservedStation)
	}
	return message, conn
}
