package main

import (
	"encoding/json"
	"errors"
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
			var car types.Car
			err := json.Unmarshal(message.Data, &car)
			if err != nil {
				responseMessage.Status = "Erro ao decodificar dados do carro"
			} else {
				// Salvar o carro em arquivo JSON
				err = s.saveCarToFile(car)
				if err != nil {
					responseMessage.Status = "Erro ao salvar o carro"
				} else {
					responseMessage.Status = "Carro registrado com sucesso"
				}
			}

		case types.RegisterStation:
			var station types.Station
			err := json.Unmarshal(message.Data, &station)
			if err != nil {
				responseMessage.Status = "Erro ao decodificar dados da estação"
			} else {
				// Salvar a estação em arquivo JSON
				err = s.saveStationToFile(station)
				if err != nil {
					responseMessage.Status = "Erro ao salvar a estação"
				} else {
					responseMessage.Status = "Estação registrada com sucesso"
				}
			}
		case types.UserLogin:
			var loginData map[string]string
			err := json.Unmarshal(message.Data, &loginData)
			if err != nil {
				responseMessage.Status = "Erro ao decodificar dados de login"
			} else {
				cars, err := s.listCarsFromFile()
				if err != nil {
					responseMessage.Status = "Erro ao acessar os dados dos carros"
				} else {
					valid := false
					for _, car := range cars {
						if car.User == loginData["user"] && car.Password == loginData["password"] {
							valid = true
							responseMessage.Status = "Login bem-sucedido"
							responseMessage.Data, _ = json.Marshal(car)
							break
						}
					}
					if !valid {
						responseMessage.Status = "Usuário ou senha inválidos"
					}
				}
			}
		case types.ReserveStation:

		case types.RechargeCar:

		case types.PayRecharge:

		// case "get_car":
		// 	// Consultar o carro no arquivo JSON
		// 	car, err := s.getCarFromFile(message.ID)
		// 	if err != nil {
		// 		responseMessage.Status = "Carro não encontrado"
		// 	} else {
		// 		responseMessage.Status = "Carro encontrado"
		// 		responseMessage.Data = car
		// 	}

		// case "get_station":
		// 	// Consultar a estação no arquivo JSON
		// 	station, err := s.getStationFromFile(message.ID)
		// 	if err != nil {
		// 		responseMessage.Status = "Estação não encontrada"
		// 	} else {
		// 		responseMessage.Status = "Estação encontrada"
		// 		responseMessage.Data = station
		// 	}

		// case "list_cars":
		// 	// Listar todos os carros salvos em arquivos JSON
		// 	cars, err := s.listCarsFromFile()
		// 	if err != nil {
		// 		responseMessage.Status = "Erro ao listar carros"
		// 	} else {
		// 		responseMessage.Status = "Lista de carros"
		// 		responseMessage.Data = cars
		// 	}

		// case "list_stations":
		// 	// Listar todas as estações salvas em arquivos JSON
		// 	stations, err := s.listStationsFromFile()
		// 	if err != nil {
		// 		responseMessage.Status = "Erro ao listar estações"
		// 	} else {
		// 		responseMessage.Status = "Lista de estações"
		// 		responseMessage.Data = stations
		// 	}
		// case "teste":
		// 	// Consultar o carro no arquivo JSON
		// 	car, err := s.getCarFromFile(message.ID)
		// 	fmt.Printf("Carro: %+v\n X: %+v\n Y: %+v\n", car.CarID, car.CoordX, car.CoordY)
		// 	if err != nil {
		// 		responseMessage.Status = "Carro não encontrado"
		// 	} else {
		// 		responseMessage.Status = "Carro encontrado"
		// 		responseMessage.Data = car
		// 	}
		case types.GetRecommendedStation:
			car := types.Car{}
			err := json.Unmarshal(message.Data, &car)
			if err != nil {
				responseMessage.Status = "Erro ao decodificar dados do carro"
			}
			station, err := s.getBestStation(car.CarID)
			if err != nil {
				responseMessage.Status = "Estação não encontrada, na requisição GetRecommendedStation"
			} else {
				responseMessage.Status = "Estação encontrada"
				responseMessage.Data, err = json.Marshal(station)
				if err != nil {
					responseMessage.Status = "Erro ao serializar estação, na requisição GetRecommendedStation"
				}
			}

		default:
			responseMessage.Status = "Escolha desconhecida"
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
	return types.Car{}, fmt.Errorf("carro não encontrado")
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
	return types.Station{}, fmt.Errorf("estação não encontrada")
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

	if car.CarID == 0 {
		return types.Station{}, errors.New("car not found")
	}

	// Inicializar as variáveis para calcular a estação mais próxima
	var bestStation types.Station
	minDistance := math.MaxFloat64 // Um valor muito grande, que será substituído

	// Função para calcular a distância entre dois pontos geográficos
	// Usando a fórmula de Haversine
	haversine := func(lat1, lon1, lat2, lon2 float64) float64 {
		const R = 6371 // Raio da Terra em km
		lat1Rad := lat1 * math.Pi / 180
		lon1Rad := lon1 * math.Pi / 180
		lat2Rad := lat2 * math.Pi / 180
		lon2Rad := lon2 * math.Pi / 180

		dlat := lat2Rad - lat1Rad
		dlon := lon2Rad - lon1Rad

		a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
		c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

		return R * c // Distância em km
	}

	// Calcular a distância entre o carro e todas as estações
	for _, station := range stations {
		distance := haversine(float64(car.CoordX), float64(car.CoordY), float64(station.CoordX), float64(station.CoordY))

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
