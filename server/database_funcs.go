package main

import (
	"encoding/json"
	"fmt"
	"main/types"
	"math"
	"net"
	"os"
	"slices"
)

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

func (s *Server) GetStationInfo(id int, conn net.Conn) (types.Station, error) {
	message := types.Message{Req: types.StationUpdate}
	err := s.SendResponse(conn, message)
	if err != nil {
		fmt.Println("Erro em GetStationInfo:", err)
	}

	station, err := s.getStationFromFile(id)
	if err != nil {
		fmt.Println("Erro em GetStationInfo:", err)
	}
	return station, err
}
