package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/types"
	"math"
	"net"
	"os"
)

type Server struct {
	address  string // Porta do TCP
	listener net.Listener
	quitch   chan struct{}
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		quitch:  make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		fmt.Println("Erro ao iniciar o servidor:", err)
		return err
	}
	defer ln.Close()
	s.listener = ln
	fmt.Println("Servidor iniciado na porta", s.address)

	go s.acceptLoop()
	<-s.quitch
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		fmt.Println("Conexão aceita de", conn.RemoteAddr())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	for {
		decoder := json.NewDecoder(conn)
		message := types.Message{}
		var err = decoder.Decode(&message)
		if err != nil {
			fmt.Println("Erro ao decodificar JSON:", err)
			break
		}

		responseMessage := types.ResponseMessage{}

		switch message.Req {

		case types.RegisterCar:
			car := message.Car
			// if !ok {
			// 	log.Fatal("Erro ao converter os dados da mensagem para types.Car")
			// }
			// err := json.Unmarshal(data, &car)
			// if err != nil {
			// 	responseMessage.Status = types.Error
			// } else {
			// Salvar o carro em arquivo JSON
			err = s.saveCarToFile(car)
			if err != nil {
				responseMessage.Status = types.Error
			} else {
				responseMessage.Data = car
				responseMessage.Status = types.Success
			}
		case types.RegisterStation:
			station := message.Station
			// err := json.Unmarshal(data, &station)
			// if err != nil {
			// 	responseMessage.Status = types.Error
			// 	fmt.Print("Erro ao decodificar o JSON")
			// } else {
			// Salvar a estação em arquivo JSON
			err = s.saveStationToFile(station)
			if err != nil {
				responseMessage.Status = types.Error
				fmt.Println("Erro ao salvar a estação")
			} else {
				responseMessage.Data = station
				responseMessage.Status = types.Success
				fmt.Println("Estação registrada com sucesso")
			}
		case types.UserLogin:
			loginData := message.Car
			// err := json.Unmarshal(data, &loginData)
			// if err != nil {
			// 	responseMessage.Status = types.Error
			// 	fmt.Println("Erro ao decodificar dados de login")
			// } else {
			cars, err := s.listCarsFromFile()
			if err != nil {
				responseMessage.Status = types.Error
				fmt.Println("Erro ao acessar os dados dos carros")
			} else {
				valid := false
				for _, car := range cars {
					if car.User == loginData.User && car.Password == loginData.Password {
						valid = true
						responseMessage.Status = types.Success
						fmt.Println("Login bem-sucedido")
						responseMessage.Data = car
						break
					}
				}
				if !valid {
					responseMessage.Status = types.Error
					fmt.Println("Usuário ou senha inválidos")
				}
			}
		case types.ListStations:
			stations, err := s.listStationsFromFile()
			if err != nil {
				responseMessage.Status = types.Error
				fmt.Println("Erro ao listar estações")
			} else {
				responseMessage.Status = types.Success
				fmt.Println("Estações listadas com sucesso")
				responseMessage.Data = stations
			}

		case types.ReserveStation:

		case types.RechargeCar:

		case types.PayRecharge:

		case types.GetRecommendedStation:
			car := message.Car
			// err := json.Unmarshal(data, &car)
			// if err != nil {
			// 	responseMessage.Status = types.Error
			// 	fmt.Println("Erro ao decodificar dados do carro")
			// }
			// fmt.Println(car.CarID)
			station, err := s.getBestStation(car.CarID)
			if err != nil {
				responseMessage.Status = types.Error
				fmt.Println("Estação não encontrada, na requisição GetRecommendedStation")
			} else {
				responseMessage.Data = station
				responseMessage.Status = types.Success
				fmt.Println("Estação encontrada")
			}

		default:
			responseMessage.Status = types.Error
			fmt.Println("Requisição inválida")
		}

		// Enviar a resposta para o cliente
		responseBuf, err := json.Marshal(responseMessage)
		if err != nil {
			fmt.Println("Erro ao serializar resposta:", err)
			return
		}
		_, err = conn.Write(responseBuf)
		if err != nil {
			fmt.Println("Erro ao enviar resposta:", err)
			return
		}
	}
}

// Funções para salvar e ler os dados em arquivos JSON
func (s *Server) saveCarToFile(car types.Car) error {
	cars, err := s.listCarsFromFile()
	if err != nil {
		cars = []types.Car{}
	}
	for i, existingCar := range cars {
		if existingCar.CarID == car.CarID {
			cars[i] = car
			fmt.Println("Tentativa de alterar um carro existente")

			return s.saveJSONToFile("../data/cars.json", cars)
		}
	}
	cars = append(cars, car)
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
	return types.Car{}, fmt.Errorf("Carro não encontrado")
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
	return types.Station{}, fmt.Errorf("Estação não encontrada")
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

func main() {
	server := NewServer(":8080")
	log.Fatal(server.Start())
}
