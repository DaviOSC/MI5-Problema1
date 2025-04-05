package main

import (
	"encoding/json"
	"fmt"
	"main/types"
	"math"
	"net"
	"os"
)

// Enviar mensagem para um cliente
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

// Salvar um carro em um arquivo JSON
func (s *Server) saveCarToFile(car types.Car) error {
	// Listar carros que já existem no arquivo
	cars, err := s.listCarsFromFile()
	if err != nil {
		cars = []types.Car{}
	}

	found := false
	// Verificar se o carro a ser salvo já existe no arquivo
	for i, existingCar := range cars {
		if existingCar.CarID == car.CarID {
			// Caso encontre, o carro é atualizado
			cars[i] = car
			found = true
			break //Interrompe o loop após encontrar o carro
		}
	}

	if !found {
		// Se o carro for novo, é inserido na lista
		cars = append(cars, car)
	}
	// Lista de carros é salva em JSON
	return s.saveJSONToFile("../data/cars.json", cars)
}

// Salvar um posto em um arquivo JSON
func (s *Server) saveStationToFile(station types.Station) error {
	// Listar Postos que já existem no arquivo
	stations, err := s.listStationsFromFile()
	if err != nil {
		stations = []types.Station{}
	}

	// Verificar se o Posto a ser salvo já existe no arquivo
	for i, existingStation := range stations {
		if existingStation.StationID == station.StationID {
			// Caso encontre, o Posto é atualizado
			stations[i] = station
			fmt.Println("Tentativa de alterar uma estação existente")
			return s.saveJSONToFile("../data/stations.json", stations)
		}
	}
	// Se o Posto for novo, é inserido na lista
	stations = append(stations, station)
	return s.saveJSONToFile("../data/stations.json", stations)
}

// Função para salvar uma estrutura em um arquivo JSON
func (s *Server) saveJSONToFile(fileName string, v interface{}) error {
	// Estrutura é serializada
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Erro ao serializar JSON:", err)
		return err
	}
	// Dados são escritos no arquivo
	err = os.WriteFile(fileName, jsonData, 0644)
	if err != nil {
		fmt.Println("Erro ao salvar o arquivo JSON:", err)
		return err
	}
	return nil
}

// Retorna uma lista com todos os carros no arquivo JSON
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

// Retorna uma lista com todos os Postos no arquivo JSON
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

// Retorna um carro especifico, atraves do ID, do arquivo JSON
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

// Retorna um posto especifico, atraves do ID, do arquivo JSON
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

// Calcula e retorna o melhor posto para o carro passado como parâmetro
func (s *Server) getBestStation(car types.Car) (types.Station, error) {

	var bestStation types.Station
	minScore := math.MaxFloat64

	for stationID, conn := range s.connectedStations {
		station, err := s.GetStationInfo(stationID, conn)
		if err != nil {
			fmt.Printf("Erro ao obter informações da estação %d: %v\n", stationID, err)
			continue
		}
		// Calcula distância
		distance := math.Abs(float64(car.CoordX-station.CoordX)) + math.Abs(float64(car.CoordY-station.CoordY))
		// Calcula o tempo de espera baseado na quantidade de carros na fila
		waitTime := float64(station.CarsWaiting) * types.RechargeTime
		// Formula do Score
		score := (distance * types.CarSpeed) + waitTime
		// O posto com menor score é escolhido
		if score < minScore {
			minScore = score
			bestStation = station
		}
	}

	if minScore == math.MaxFloat64 {
		return types.Station{}, fmt.Errorf("nenhuma estação disponível")
	}

	return bestStation, nil
}

// Envia uma requisição ao posto para que atualize suas informações com o servidor
func (s *Server) GetStationInfo(id int, conn net.Conn) (types.Station, error) {
	message := types.Message{Req: types.StationUpdate}
	// Envia a requisição
	err := s.SendResponse(conn, message)
	if err != nil {
		fmt.Println("Erro em GetStationInfo:", err)
	}
	// Recebe os dados do arquivo
	station, err := s.getStationFromFile(id)
	if err != nil {
		fmt.Println("Erro em GetStationInfo:", err)
	}
	return station, err
}
