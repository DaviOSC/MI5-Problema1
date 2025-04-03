package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"math"
	"net"
	"os"
	"slices"
)

func (s *Server) HandleRegisterCar(message types.Message) types.Message {
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
	return responseMessage
}

func (s *Server) HandleRegisterStation(message types.Message) types.Message {
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
	return responseMessage
}

func (s *Server) HandleUserLogin(message types.Message, conn net.Conn) types.Message {
	loginData := message.Car
	responseMessage := types.Message{Req: types.UserLogin}

	cars, err := s.listCarsFromFile()
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Println("Erro ao acessar os dados dos carros")
		return responseMessage
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
	return responseMessage
}

func (s *Server) HandleListStations(message types.Message) types.Message {
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

	return responseMessage
}

func (s *Server) HandleReserveStation(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.ReserveStation, Status: types.Success}
	var err error
	station := types.Station{}
	stationID := message.Car.RecomendedStation

	// Verifica se o carro possue uma estação recomendada
	if stationID == 0 {
		station, err = s.getBestStation(message.Car)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Estação não encontrada, na requisição ReserveStation, para o carro de id %d\n", message.Car.CarID)
			return responseMessage
		}
	} else {
		station, err = s.getStationFromFile(stationID)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Estação não encontrada, na requisição ReserveStation, para a estação de id %d\n", stationID)
			return responseMessage
		}
	}
	//Verifica se o carro tá na fila
	for _, carID := range station.CarList {
		if carID == message.Car.CarID {
			responseMessage.Status = types.Error
			responseMessage.Car = message.Car
			fmt.Printf("Erro: O carro com ID %d já está na fila da estação com ID %d.\n", message.Car.CarID, station.StationID)
			return responseMessage
		}
	}
	// Atualizando dados da estação e salvando em JSON
	station.CarList = append(station.CarList, message.Car.CarID)
	station.CarsWaiting += 1
	err = s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar a estação, na requisição ReserveStation, para a estação de id %d\n", station.StationID)
		return responseMessage
	}
	message.Car.ReservedStation = station.StationID

	responseMessage.Car = message.Car
	err = s.saveCarToFile(message.Car)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar o carro, na requisição ReserveStation, para o carro de id %d\n", message.Car.CarID)
		return responseMessage
	}

	return responseMessage
}

func (s *Server) HandleSelectStation(message types.Message, conn net.Conn) types.Message {
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
		return responseMessage
	}
	// Adicionar a estação ao mapa de estações conectadas
	//message.Station.Conn = conn.RemoteAddr().String() // Salvar o IP da conexão
	s.connectedStations[stationID] = conn

	responseMessage.Status = types.Success
	responseMessage.Station = message.Station
	fmt.Printf("Estação com ID %d conectada com sucesso. IP: %s\n", responseMessage.Station.StationID, conn.RemoteAddr().String())
	fmt.Printf("Estação %d adicionada com sucesso. IP: %s\n", stationID, conn.RemoteAddr().String())
	return responseMessage
}
func (s *Server) HandleListActiveStations(message types.Message) types.Message {
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
	return responseMessage
}

func (s *Server) HandlePayRecharge(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.PayRecharge, Status: types.Success}
	stationID := message.Car.ReservedStation
	station, err := s.getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada, na requisição PayRecharge, para a estação de id %d\n", stationID)
		return responseMessage
	}
	if message.Car.ReservedStation == 0 {
		responseMessage.Status = types.Error
		responseMessage.Err = "reserve a estação antes de pagar"
		return responseMessage
	}

	if station.InUseBy == 0 && station.CarList[0] == message.Car.CarID {
		station.InUseBy = message.Car.CarID
		station.CarsWaiting -= 1
		station.CarList = station.CarList[1:]
		err := s.saveStationToFile(station)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Erro ao salvar a estação, na requisição PayRecharge, para a estação de id %d\n", stationID)
			return responseMessage
		}
		// TODO simular pagamento
		responseMessage.Car = message.Car
		err = s.saveCarToFile(message.Car)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Erro ao salvar o carro, na requisição PayRecharge, para o carro de id %d\n", message.Car.CarID)
			return responseMessage
		}
	} else {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não disponível, na requisição PayRecharge, para a estação de id %d\n", stationID)
	}

	return responseMessage
}

// func HandlePaymentHistory(message types.Message) types.Message {
// 	car := message.Car
// 	responseMessage := types.Message{Req: types.PaymentHistory}

// 	payments := []types.Payment{}
// 	responseMessage.
// 	return responseMessage

// }

func (s *Server) HandleGetRecommendedStation(message types.Message) types.Message {
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

	return responseMessage
}

func (s *Server) HandleGetReservedStation(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.GetReservedStation}
	car, err := s.getCarFromFile(message.Car.CarID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao obter o carro com ID %d: %v\n", message.Car.CarID, err)
		return responseMessage
	}
	if car.ReservedStation == 0 {
		responseMessage.Status = types.Error
		fmt.Printf("Carro %d não possui uma estação reservada.\n", car.CarID)
		return responseMessage
	}
	station, err := s.getStationFromFile(car.ReservedStation)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao obter a estação reservada com ID %d: %v\n", car.ReservedStation, err)
		return responseMessage
	}

	fmt.Printf("Retornando a estação %d para o carro %d", station.StationID, car.CarID)
	responseMessage.Status = types.Success
	responseMessage.Car = car
	responseMessage.Station = station
	return responseMessage
}

func (s *Server) HandleRechargeComplete(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.RechargeComplete, Status: types.Success}
	car, stationID := message.Car, message.Car.ReservedStation

	// Obter a estação do arquivo
	station, err := s.getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada para o ID %d\n", stationID)
		return responseMessage
	}

	// Validar se a estação está em uso pelo carro correto
	if station.InUseBy != car.CarID {
		responseMessage.Status = types.Error
		fmt.Printf("Estação %d não está em uso pelo carro %d\n", stationID, car.CarID)
		return responseMessage
	}
	// Atualizar os dados do carro e da estação
	car.ReservedStation, car.RecomendedStation = 0, 0
	station.InUseBy = 0
	car.BatteryLevel = 100

	err = s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar a estação %d\n", station.StationID)
		return responseMessage
	}

	err = s.saveCarToFile(car)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar o carro %d\n", car.CarID)
		return responseMessage
	}

	fmt.Printf("Recarga concluída para o carro %d na estação %d.\n", car.CarID, station.StationID)
	responseMessage.Car = car
	return responseMessage
}
func (s *Server) HandleStartRecharge(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.StartRecharge, Status: types.Success}
	car, stationID := message.Car, message.Car.ReservedStation

	// Obter a estação do arquivo
	station, err := s.getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada para o ID %d\n", stationID)
		return responseMessage
	}

	// Marcar a estação como em uso pelo carro
	station.InUseBy = car.CarID
	err = s.saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar a estação %d\n", stationID)
		return responseMessage
	}

	fmt.Printf("Recarga iniciada para o carro %d na estação %d.\n", car.CarID, station.StationID)
	//TODO o cliente ainda consegue enviar mensagens no terminal no periodo. elas são processadas em sequencia apos o tempo
	responseMessage.Car = car
	return responseMessage
}
func (s *Server) SendResponse(conn net.Conn, responseMessage types.Message) error {
	responseBuf, err := json.Marshal(responseMessage)
	if err != nil {
		return fmt.Errorf("erro ao serializar resposta: %w", err)
	}

	_, err = conn.Write(responseBuf)
	if err != nil {
		return fmt.Errorf("erro ao enviar resposta: %w", err)
	}
	return nil
}

// Funções para salvar e ler os dados em arquivos JSON
func (s *Server) saveCarToFile(car types.Car) error {
	cars, err := s.listCarsFromFile()
	if err != nil {
		cars = []types.Car{}
	}

	found := false
	for i, existingCar := range cars {
		if existingCar.CarID == car.CarID {
			cars[i] = car
			found = true
			break //Interrompe o loop após encontrar o carro
		}
	}

	if !found {
		cars = append(cars, car)
	}

	return s.saveJSONToFile("../data/cars.json", cars)
}

func (s *Server) saveStationToFile(station types.Station) error {
	stations, err := s.listStationsFromFile()
	if err != nil {
		stations = []types.Station{}
	}
	for i, existingStation := range stations {
		if existingStation.StationID == station.StationID {
			stations[i] = station
			fmt.Println("Tentativa de alterar uma estação existente")
			return s.saveJSONToFile("../data/stations.json", stations)
		}
	}
	stations = append(stations, station)
	return s.saveJSONToFile("../data/stations.json", stations)
}

func (s *Server) saveJSONToFile(fileName string, v interface{}) error {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Erro ao serializar JSON:", err)
		return err
	}
	err = os.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar o arquivo JSON:", err)
		return err
	}
	return nil
}

func (s *Server) listCarsFromFile() ([]types.Car, error) {
	var cars []types.Car
	data, err := os.ReadFile("../data/cars.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cars)
	if err != nil {
		return nil, err
	}
	return cars, nil
}

func (s *Server) listStationsFromFile() ([]types.Station, error) {
	var stations []types.Station
	data, err := os.ReadFile("../data/stations.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &stations)
	if err != nil {
		return nil, err
	}
	return stations, nil
}

func (s *Server) getCarFromFile(id int) (types.Car, error) {
	cars, err := s.listCarsFromFile()
	if err != nil {
		return types.Car{}, err
	}
	for _, car := range cars {
		if car.CarID == id {
			return car, nil
		}
	}
	return types.Car{}, err
}

func (s *Server) getStationFromFile(id int) (types.Station, error) {
	stations, err := s.listStationsFromFile()
	if err != nil {
		return types.Station{}, err
	}
	for _, station := range stations {
		if station.StationID == id {
			return station, nil
		}
	}
	return types.Station{}, err
}

func (s *Server) getBestStation(car types.Car) (types.Station, error) {
	// s.sessionsMu.Lock()
	// defer s.sessionsMu.Unlock()

	var bestStation types.Station
	minDistance := math.MaxFloat64

	for stationID, conn := range s.connectedStations {
		station, err := s.GetStationInfo(stationID, conn)
		if err != nil {
			fmt.Printf("Erro ao obter informações da estação %d: %v\n", stationID, err)
			continue
		}
		distance := math.Abs(float64(car.CoordX-station.CoordX)) + math.Abs(float64(car.CoordY-station.CoordY))
		if distance < minDistance {
			minDistance = distance
			bestStation = station
		}
	}

	if minDistance == math.MaxFloat64 {
		return types.Station{}, fmt.Errorf("nenhuma estação disponível")
	}

	//TODO outros critérios de escolha da estação
	return bestStation, nil
}

func (s *Server) HandleCarUpdate(message types.Message) types.Message {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	if _, exists := s.loggedCars[message.Car.CarID]; exists {

		// Persiste no arquivo diretamente
		if err := s.saveCarToFile(message.Car); err != nil {
			fmt.Printf("[ERRO] Falha ao salvar bateria do carro %d: %v\n", message.Car.CarID, err)
			return types.Message{Status: types.Error}
		}

		fmt.Printf("[SYNC] Bateria do carro %d atualizada para %d%%\n", message.Car.CarID, message.Car.BatteryLevel)
	}

	return types.Message{Status: types.Success}
}

func (s *Server) CloseConnection(message types.Message) {
	fmt.Printf("Carro desconectado: ID %d\n", message.Car.CarID)
	if message.Car.CarID != 0 {
		reservedStationID := message.Car.ReservedStation
		fmt.Printf("Carro %d desconectado da estação %d\n", message.Car.CarID, reservedStationID)
		if reservedStationID != 0 {
			reservedStation, err := s.getStationFromFile(reservedStationID)
			if err != nil {
				fmt.Println("Erro em CloseConnection (Carro):", err)
				return
			}

			reservedStation.CarsWaiting -= 1
			index := slices.Index(reservedStation.CarList, message.Car.CarID)
			if index >= 0 {
				reservedStation.CarList = slices.Delete(reservedStation.CarList, index, index+1)
			}
			if reservedStation.InUseBy == message.Car.CarID {
				reservedStation.InUseBy = 0
			}
			message.Car.RecomendedStation = 0
			message.Car.ReservedStation = 0
			s.saveCarToFile(message.Car)
			s.saveStationToFile(reservedStation)
		}
	} else if message.Station.StationID != 0 {
		fmt.Printf("Estação %d desconectada\n", message.Station.StationID)

		s.sessionsMu.Lock()
		delete(s.connectedStations, message.Station.StationID)
		s.sessionsMu.Unlock()

		err := s.saveStationToFile(message.Station)
		if err != nil {
			fmt.Println("Erro em CloseConnection (Estação):", err)
		}

	} else {
		fmt.Println("Entidade desconhecida ao tentar fechar conexão.")
	}
}

func (s *Server) SendMessageAndReadResponse(conn net.Conn, message types.Message) (types.Message, error) {
	err := json.NewEncoder(conn).Encode(message)
	if err != nil {
		log.Fatal("Erro ao serializar a mensagem:", err)
	}

	var responseMessage types.Message
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&responseMessage)
	if err != nil {
		// Se houver um erro ao decodificar a resposta, o programa será encerrado
		log.Fatal("Erro ao decodificar a resposta:", err)
	}
	return responseMessage, err
}

func (s *Server) GetStationInfo(id int, conn net.Conn) (types.Station, error) {
	message := types.Message{Req: types.StationUpdate}
	responseMessage, err := s.SendMessageAndReadResponse(conn, message)
	if err != nil {
		fmt.Println("Erro em GetStationInfo:", err)
	}

	if responseMessage.Status == types.Success {
		station, err := s.getStationFromFile(id)
		if err != nil {
			fmt.Println("Erro em GetStationInfo:", err)
		}
		return station, err
	} else {
		err = fmt.Errorf(responseMessage.Err)
		return types.Station{}, err
	}
}

func (s *Server) HandleStationUpdate(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.StationUpdate, Status: types.Success}
	station := message.Station
	err := s.saveStationToFile(station)
	if err != nil {
		fmt.Println("Em HandleStationUpdate:", err)
		responseMessage.Status = types.Error
	}

	return responseMessage
}
