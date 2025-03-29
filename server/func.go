package main

import (
	"encoding/json"
	"fmt"
	"main/types"
	"math"
	"net"
	"os"
	"time"
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

func (s *Server) HandleUserLogin(message types.Message) types.Message {
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

			s.sessionsMu.Lock()
			s.loggedCars[car.CarID] = car
			s.sessionsMu.Unlock()

			fmt.Println("Login bem-sucedido")
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
		fmt.Println("Erro ao listar estações")
	} else {
		responseMessage.Status = types.Success
		fmt.Println("Estações listadas com sucesso")
		responseMessage.StationList = stations
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
		station, err = s.getBestStation(message.Car.CarID)
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

	go s.startCarMovement(message.Car, station)

	responseMessage.Car = message.Car
	err = s.saveCarToFile(message.Car)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar o carro, na requisição ReserveStation, para o carro de id %d\n", message.Car.CarID)
		return responseMessage
	}

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
		responseMessage.Err = fmt.Errorf("Reserve a estação antes de pagar")
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
	station, err := s.getBestStation(car.CarID)
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

    fmt.Printf("Recarga iniciada para o carro %d na estação %d. Tempo estimado: 10 segundos.\n", car.CarID, station.StationID)
    time.Sleep(10 * time.Second)
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

func (s *Server)getStationFromFile(id int) (types.Station, error) {
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

func (s *Server) getBestStation(carId int) (types.Station, error) {
	stations, err := s.listStationsFromFile()
	if err != nil {
		return types.Station{}, err
	}
	cars, err := s.listCarsFromFile()
	if err != nil {
		return types.Station{}, err
	}
	var car types.Car
	for _, c := range cars {
		if c.CarID == carId {
			car = c
			break
		}
	}

	//TODO verificação de carros na fila, e de tempo de espera
	var bestStation types.Station
	minDistance := math.MaxFloat64

	for _, station := range stations {
		distance := math.Abs(float64(car.CoordX-station.CoordX)) + math.Abs(float64(car.CoordY-station.CoordY))
		if distance < minDistance {
			minDistance = distance
			bestStation = station
		}
	}
	return bestStation, nil
}

func (s *Server) startCarMovement(car types.Car, station types.Station) {
    var speed = 0.5
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    fmt.Printf("Trajeto do carro %d: para a estação %d\n", car.CarID, station.StationID)
    coordX := float64(car.CoordX)
    coordY := float64(car.CoordY)
    stationX := float64(station.CoordX)
    stationY := float64(station.CoordY)

    for {
        select {
        case <-ticker.C:
            if coordX < stationX {
                coordX += speed
                if coordX > stationX {
                    coordX = stationX
                }
            } else if coordX > stationX {
                coordX -= speed
                if coordX < stationX {
                    coordX = stationX
                }
            }
            if coordY < stationY {
                coordY += speed
                if coordY > stationY {
                    coordY = stationY
                }
            } else if coordY > stationY {
                coordY -= speed
                if coordY < stationY {
                    coordY = stationY
                }
            }
            car.CoordX = int(coordX)
            car.CoordY = int(coordY)
            err := s.saveCarToFile(car)
            if err != nil {
                fmt.Printf("Erro ao salvar a posição do carro %d: %v\n", car.CarID, err)
                return
            }
            if math.Abs(coordX-stationX) < 0.01 && math.Abs(coordY-stationY) < 0.01 {
                fmt.Printf("Carro %d chegou à estação %d.\n", car.CarID, station.StationID)
                return
            }
        }
    }
}


func (s *Server) HandleBatterySync(message types.Message) types.Message {
    s.sessionsMu.Lock()
    defer s.sessionsMu.Unlock()

    if car, exists := s.loggedCars[message.Car.CarID]; exists {
        // Atualiza nível da bateria
        car.BatteryLevel = message.Car.BatteryLevel
        s.loggedCars[message.Car.CarID] = car
        
        // Persiste no arquivo diretamente
        if err := s.saveCarToFile(car); err != nil {
            fmt.Printf("[ERRO] Falha ao salvar bateria do carro %d: %v\n", car.CarID, err)
            return types.Message{Status: types.Error}
        }
        
        fmt.Printf("[SYNC] Bateria do carro %d atualizada para %d%%\n", car.CarID, car.BatteryLevel)
    }

    return types.Message{Status: types.Success}
}

