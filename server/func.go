package main

import (
	"encoding/json"
	"fmt"
	"main/types"
	"math"
	"net"
	"os"
)

func HandleRegisterCar(message types.Message) types.Message {
	car := message.Car
	responseMessage := types.Message{Req: types.RegisterCar}

	err := saveCarToFile(car)
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

func HandleRegisterStation(message types.Message) types.Message {
	station := message.Station
	responseMessage := types.Message{Req: types.RegisterStation}

	err := saveStationToFile(station)
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

func HandleUserLogin(message types.Message) types.Message {
	loginData := message.Car
	responseMessage := types.Message{Req: types.UserLogin}

	cars, err := listCarsFromFile()
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Println("Erro ao acessar os dados dos carros")

	} else {
		valid := false
		for _, car := range cars {
			if car.User == loginData.User && car.Password == loginData.Password {
				valid = true
				responseMessage.Status = types.Success
				responseMessage.Car = car
				fmt.Println("Login bem-sucedido")
				break
			}
		}
		if !valid {
			responseMessage.Status = types.Error
			fmt.Println("Usuário ou senha inválidos")
		}
	}
	return responseMessage
}

func HandleListStations(message types.Message) types.Message {
	stations, err := listStationsFromFile()
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

func HandleReserveStation(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.ReserveStation, Status: types.Success}
	var err error
	station := types.Station{}
	stationID := message.Car.RecomendedStation

	// Verifica se o carro possue uma estação recomendada
	if stationID == 0 {
		station, err = getBestStation(message.Car.CarID)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Estação não encontrada, na requisição ReserveStation, para o carro de id %d\n", message.Car.CarID)
			return responseMessage
		}
	} else {
		station, err = getStationFromFile(stationID)
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
	err = saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
        fmt.Printf("Erro ao salvar a estação, na requisição ReserveStation, para a estação de id %d\n", station.StationID)
		return responseMessage
	}
	message.Car.ReservedStation = station.StationID
	
	responseMessage.Car = message.Car
	err = saveCarToFile(message.Car)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Erro ao salvar o carro, na requisição ReserveStation, para o carro de id %d\n", message.Car.CarID)
			return responseMessage
		}
		
	return responseMessage
}

func HandlePayRecharge(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.PayRecharge, Status: types.Success}
	stationID := message.Car.ReservedStation
	station, err := getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada, na requisição PayRecharge, para a estação de id %d\n", stationID)
		return responseMessage
	}
	fmt.Println(station.CarList)
	fmt.Println(len(station.CarList))

	if station.InUseBy == 0 && station.CarList[0] == message.Car.CarID {
		station.InUseBy = message.Car.CarID
		station.CarsWaiting -= 1
		station.CarList = station.CarList[1:]
		err := saveStationToFile(station)
		if err != nil {
			responseMessage.Status = types.Error
			fmt.Printf("Erro ao salvar a estação, na requisição PayRecharge, para a estação de id %d\n", stationID)
			return responseMessage
		}
		// TODO simular pagamento
		responseMessage.Car = message.Car
		err = saveCarToFile(message.Car)
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

func HandlePaymentHistory(message types.Message) types.Message {
	car := message.Car
	responseMessage := types.Message{Req: types.PaymentHistory}

	payments := []types.Payment{}
	responseMessage.Res
	return responseMessage

}

func HandleGetRecommendedStation(message types.Message) types.Message {
	car := message.Car
	responseMessage := types.Message{Req: types.GetRecommendedStation}

	station, err := getBestStation(car.CarID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Println("Estação não encontrada, na requisição GetRecommendedStation")
	} else {
		responseMessage.Status = types.Success
		responseMessage.Station = station
		responseMessage.Car = car
		fmt.Printf("Estação com id %d encontrada", station.StationID)
	}

	return responseMessage
}

func HandleRechargeComplete(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.RechargeComplete, Status: types.Success}
	car, stationID := message.Car, message.Car.ReservedStation
	station, err := getStationFromFile(stationID)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Estação não encontrada, na requisição RechageComplete, para a estação de id %d\n", stationID)
		responseMessage.Car = car
		return responseMessage
	}
	if station.InUseBy != car.CarID {
		responseMessage.Status = types.Error
		fmt.Printf("Informação não compativeis entre o carro e a estação")
		responseMessage.Car = car
		return responseMessage
	}

	car.ReservedStation, car.RecomendedStation = 0, 0
	station.InUseBy = 0
	responseMessage.Car = car
	err = saveStationToFile(station)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar a estação, na requisição RechargeComplete, para a estação de id %d\n", stationID)
		return responseMessage
	}
	err = saveCarToFile(car)
	if err != nil {
		responseMessage.Status = types.Error
		fmt.Printf("Erro ao salvar o carro, na requisição RechargeComplete, para o carro de id %d\n", car.CarID)
		return responseMessage
	}

	return responseMessage
}

func SendResponse(conn net.Conn, responseMessage types.Message) error {
	responseBuf, err := json.Marshal(responseMessage)
	if err != nil {
		fmt.Println("Erro ao serializar resposta:", err)
		return err
	}
	_, err = conn.Write(responseBuf)
	if err != nil {
		fmt.Println("Erro ao enviar resposta:", err)
		return err
	}

	return nil
}

// Funções para salvar e ler os dados em arquivos JSON
func saveCarToFile(car types.Car) error {
	cars, err := listCarsFromFile()
	if err != nil {
		cars = []types.Car{}
	}
	for i, existingCar := range cars {
		if existingCar.CarID == car.CarID {
			cars[i] = car
			fmt.Println("Tentativa de alterar um carro existente")

			return saveJSONToFile("../data/cars.json", cars)
		}
	}
	cars = append(cars, car)
	return saveJSONToFile("../data/cars.json", cars)
}

func saveStationToFile(station types.Station) error {
	stations, err := listStationsFromFile()
	if err != nil {
		stations = []types.Station{}
	}
	for i, existingStation := range stations {
		if existingStation.StationID == station.StationID {
			stations[i] = station
			fmt.Println("Tentativa de alterar uma estação existente")
			return saveJSONToFile("../data/stations.json", stations)
		}
	}
	stations = append(stations, station)
	return saveJSONToFile("../data/stations.json", stations)
}

func saveJSONToFile(fileName string, v interface{}) error {
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

func listCarsFromFile() ([]types.Car, error) {
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

func listStationsFromFile() ([]types.Station, error) {
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

func getCarFromFile(id int) (types.Car, error) {
	cars, err := listCarsFromFile()
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

func getStationFromFile(id int) (types.Station, error) {
	stations, err := listStationsFromFile()
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

func getBestStation(carId int) (types.Station, error) {
	stations, err := listStationsFromFile()
	if err != nil {
		return types.Station{}, err
	}
	cars, err := listCarsFromFile()
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
		distance := math.Sqrt(math.Pow(float64(car.CoordX-station.CoordX), 2) + math.Pow(float64(car.CoordY-station.CoordY), 2))

		if distance < minDistance {
			minDistance = distance
			bestStation = station
		}
	}
	return bestStation, nil
}
