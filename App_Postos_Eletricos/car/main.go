package main

import (
	"encoding/json"
	"fmt"
	"main/types"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
	}

	defer conn.Close()

	for {
		var carID, coordX, coordY, batteryLevel, recomendedStation, reservedStation int
		fmt.Print("Digite o ID do carro, coordenada X, coordenada Y, nível da bateria, estação recomendada e estação reservada: ")
		fmt.Scanln(&carID, &coordX, &coordY, &batteryLevel, &recomendedStation, &reservedStation)

		car := types.Car{
			CarID:             carID,
			CoordX:            coordX,
			CoordY:            coordY,
			BatteryLevel:      batteryLevel,
			RecomendedStation: recomendedStation,
			ReservedStation:   reservedStation,
		}

		buf, err := json.Marshal(car)

		if err != nil {
			fmt.Println("Erro serializar dados:", err)
			break
		}
		_, err = conn.Write(buf)
		if err != nil {
			fmt.Println("Erro ao enviar dados:", err)
			break
		}
	}
}
