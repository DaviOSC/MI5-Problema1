package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Car struct {
	CarID             int    `json:"car_id"`
	User              string `json:"user"`
	Password          string `json:"password"`
	CoordX            int    `json:"coord_x"`
	CoordY            int    `json:"coord_y"`
	BatteryLevel      int    `json:"battery_level"`
	RecomendedStation int    `json:"recomended_station"`
	ReservedStation   int    `json:"reserved_station"`
}

type Station struct {
	StationID  int    `json:"station_id"`
	Name       string `json:"name"`
	CoordX     int    `json:"coord_x"`
	CoordY     int    `json:"coord_y"`
	IsReserved bool   `json:"is_reserved"`
}

var (
	cars     []Car
	stations []Station
	mu       sync.Mutex
)

func loadData() {
	carFile, err := os.ReadFile("cars.json")
	if err != nil {
		fmt.Println("Erro ao ler cars.json:", err)
		return
	}
	json.Unmarshal(carFile, &cars)

	stationFile, err := os.ReadFile("posts.json")
	if err != nil {
		fmt.Println("Erro ao ler posts.json:", err)
		return
	}
	json.Unmarshal(stationFile, &stations)
}

func saveCars() {
	file, _ := json.MarshalIndent(cars, "", "  ")
	os.WriteFile("cars.json", file, 0644)
}

func saveStations() {
	file, _ := json.MarshalIndent(stations, "", "  ")
	os.WriteFile("posts.json", file, 0644)
}

func calculateDistance(x1, y1, x2, y2 int) float64 {
	return math.Sqrt(math.Pow(float64(x2-x1), 2) + math.Pow(float64(y2-y1), 2))
}

func findNearestStation(car Car) (int, string) {
	minDist := math.MaxFloat64
	nearestID := -1
	nearestName := ""

	for _, station := range stations {
		if !station.IsReserved {
			dist := calculateDistance(car.CoordX, car.CoordY, station.CoordX, station.CoordY)
			if dist < minDist {
				minDist = dist
				nearestID = station.StationID
				nearestName = station.Name
			}
		}
	}
	return nearestID, nearestName
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Autenticação
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Erro na autenticação:", err)
		return
	}

	creds := strings.Split(string(buffer[:n]), ":")
	if len(creds) != 2 {
		conn.Write([]byte("Formato inválido. Use user:password"))
		return
	}

	mu.Lock()
	var authCar *Car
	for i := range cars {
		if cars[i].User == creds[0] && cars[i].Password == creds[1] {
			authCar = &cars[i]
			break
		}
	}
	mu.Unlock()

	if authCar == nil {
		conn.Write([]byte("Autenticação falhou"))
		return
	}
	conn.Write([]byte("Autenticado com sucesso!"))

	// Processar comandos
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Conexão perdida:", err)
			return
		}

		cmd := strings.TrimSpace(string(buffer[:n]))
		switch {
		case cmd == "recommend":
			mu.Lock()
			nearestID, nearestName := findNearestStation(*authCar)
			mu.Unlock()
		
			conn.Write([]byte(fmt.Sprintf("Posto recomendado: ID %d, Nome %s", nearestID, nearestName)))
			
		case strings.HasPrefix(cmd, "reserve"):
			parts := strings.Split(cmd, " ")
			if len(parts) != 2 {
				conn.Write([]byte("Formato inválido. Use reserve [station_id]"))
				continue
			}

			stationID, err := strconv.Atoi(parts[1])
			if err != nil {
				conn.Write([]byte("ID do posto inválido"))
				continue
			}

			mu.Lock()
			var target *Station
			for i := range stations {
				if stations[i].StationID == stationID {
					target = &stations[i]
					break
				}
			}

			if target == nil {
				mu.Unlock()
				conn.Write([]byte("Posto não encontrado"))
				continue
			}

			if target.IsReserved {
				mu.Unlock()
				conn.Write([]byte("Posto já reservado"))
				continue
			}

			target.IsReserved = true
			authCar.ReservedStation = stationID
			saveStations()
			saveCars()
			mu.Unlock()
			conn.Write([]byte(fmt.Sprintf("Posto %d reservado com sucesso!", stationID)))

		case cmd == "exit":
			conn.Write([]byte("Desconectando..."))
			return

		default:
			conn.Write([]byte("Comando inválido"))
		}
	}
}

func main() {
	loadData()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Erro ao iniciar servidor:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Servidor ouvindo na porta 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Erro ao aceitar conexão:", err)
			continue
		}
		go handleConnection(conn)
	}
}