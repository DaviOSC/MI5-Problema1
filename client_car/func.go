package main

import (
	"encoding/json"
	"fmt"
	"main/types"
	"net"
	"time"
)

// HandleRegisterCar lida com a requisição de registro de carro
func HandleRegisterCar() (types.Message, error) {
	var car types.Car
	fmt.Print("Digite o ID do carro: ")
	fmt.Scanln(&car.CarID)
	fmt.Print("Digite o Usuário do carro: ")
	fmt.Scanln(&car.User)
	fmt.Print("Digite a senha: ")
	fmt.Scanln(&car.Password)
	fmt.Print("Digite a coordenada X do carro: ")
	fmt.Scanln(&car.CoordX)
	fmt.Print("Digite a coordenada Y do carro: ")
	fmt.Scanln(&car.CoordY)
	fmt.Print("Digite o nível da bateria: ")
	fmt.Scanln(&car.BatteryLevel)

	return types.Message{
		Req: types.RegisterCar,
		Car: car,
	}, nil
}

// HandleGetRecommendedStation lida com a requisição de recomendação de estação
func HandleGetRecommendedStation(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.GetRecommendedStation,
		Car: car,
	}, nil
}
func HandleGetReservedStation(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.GetReservedStation,
		Car: car,
	}, nil
}

func HandleListActiveStations() (types.Message, error) {
	return types.Message{
		Req: types.ListActiveStations,
	}, nil
}
func HandleListActiveStationsResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Estações conectadas:")
		for _, station := range responseMessage.StationList {
			fmt.Printf("ID: %d, Coordenadas: (%d, %d)\n", station.StationID, station.CoordX, station.CoordY)
		}
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro ao listar estações")
	}
}

// HandleReserveStation lida com a requisição de reserva de estação
func HandleReserveStation(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.ReserveStation,
		Car: car,
	}, nil
}

// HandleRechargeComplete lida com a requisição de recarga de carro
func HandleRechargeComplete(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.RechargeComplete,
		Car: car,
	}, nil
}

func HandleStartRecharge(car types.Car) (types.Message, error) {
	fmt.Printf("Solicitando início da recarga para o carro %d na estação %d...\n", car.CarID, car.ReservedStation)
	return types.Message{
		Req: types.StartRecharge,
		Car: car,
	}, nil
}

func HandlePaymentHistory(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.PaymentHistory,
		Car: car,
	}, nil
}

// HandlePayRecharge lida com a requisição de pagamento de recarga
func HandlePayRecharge(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.PayRecharge,
		Car: car,
	}, nil
}
func HandleExit(car types.Car) types.Message {
	return types.Message{
		Req: types.ExitCar,
		Car: car,
	}
}
func HandleLogin(conn net.Conn) (types.Car, error) {
	car := types.Car{}

	fmt.Print("Digite o usuário: ")
	fmt.Scanln(&car.User)
	fmt.Print("Digite a senha: ")
	fmt.Scanln(&car.Password)

	message := types.Message{
		Req: types.UserLogin,
		Car: car,
	}

	buf, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Erro ao serializar a mensagem de login:", err)
		return types.Car{}, err
	}
	_, err = conn.Write(buf)
	if err != nil {
		fmt.Println("Erro ao enviar a mensagem de login:", err)
		return types.Car{}, err
	}

	response, err := ReadResponse(conn)
	if err != nil {
		fmt.Println("Erro receber resposta do servidor:", err)
		return types.Car{}, err
	}

	if response.Status == types.Error {
		return types.Car{}, fmt.Errorf("senha ou usuario invalidos")
	}

	return response.Car, nil
}

func ReadResponse(conn net.Conn) (types.Message, error) {
	decoder := json.NewDecoder(conn)
	responseMessage := types.Message{}
	err := decoder.Decode(&responseMessage)
	if err != nil {
		fmt.Println("erro:", err)
		return responseMessage, err
	}
	return responseMessage, nil
}

func HandleRegisterCarResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Carro registrado com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro: %s", responseMessage.Err)
	}
}

func HandleGetRecommendedStationResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Printf(
			`ID da estação recomendada: %d, 
Coordenadas: (x: %d, y: %d)`,
			responseMessage.Station.StationID,
			responseMessage.Station.CoordX, responseMessage.Station.CoordY,
		)
		fmt.Println()
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro: %s", responseMessage.Err)
	}
}

func HandleReserveStationResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Reserva de estação realizada com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro: %s", responseMessage.Err)
	}
}

func HandleRechargeCompleteResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Recarga concluída com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro: %s", responseMessage.Err)
	}
}

func HandleStartRechargeResponse(client *CarClient, responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		//Espera o tempo de recarga
		rechargeTime := types.RechargeTime * (float64)(100-client.car.BatteryLevel)
		fmt.Printf("Esperando %f segundos para simular a recarga...\n", rechargeTime)
		time.Sleep(time.Second * time.Duration(rechargeTime))

		// Criar a mensagem de conclusão da recarga
		message, err := HandleRechargeComplete(responseMessage.Car)
		if err != nil {
			return responseMessage.Car, fmt.Errorf("erro ao preparar mensagem de conclusão da recarga: %v", err)
		}

		// Enviar a mensagem de conclusão ao servidor
		err = client.SendMessage(message)
		if err != nil {
			return responseMessage.Car, fmt.Errorf("erro ao enviar mensagem de conclusão da recarga: %v", err)
		}

		// Ler a resposta de conclusão da recarga
		completeResponse, err := client.ReadResponse()
		if err != nil {
			return responseMessage.Car, fmt.Errorf("erro ao receber resposta de conclusão da recarga: %v", err)
		}

		// Processar a resposta de conclusão da recarga
		if completeResponse.Status == types.Success {
			fmt.Println("Recarga concluída com sucesso.")
			return completeResponse.Car, nil
		} else {
			return completeResponse.Car, fmt.Errorf("erro ao concluir recarga: %v", completeResponse.Status)
		}
	} else {
		return responseMessage.Car, fmt.Errorf("erro ao iniciar recarga: %v", responseMessage.Err)
	}
}

func HandlePayRechargeResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Pagamento realizado com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf(responseMessage.Err)
	}
}
func HandleListStationsResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Estações:")
		for _, station := range responseMessage.StationList {
			fmt.Printf(
				`ID: %d
Coordenadas: (x: %d, y: %d)`,
				station.StationID, station.CoordX, station.CoordY,
			)
			fmt.Println()
		}
		return types.Car{}, nil
	} else {
		return types.Car{}, fmt.Errorf("erro ao listar estações")
	}
}

/*
Movimenta o carro para a coordenada especificada, levando em consideração
velocidade e a porcentagem da bateria
*/
func MoveCarTo(client *CarClient, x int, y int) types.Car {
	car := client.car
	totalMetersTraveled := 0
	fullyDischargedCar := false

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	fmt.Printf("Trajeto para as coordenadas: (%d, %d)\n", x, y)
	coordX := car.CoordX
	coordY := car.CoordY
	stationX := x
	stationY := y
	speed := types.CarSpeed

	for {
		select {
		case <-ticker.C:
			if car.BatteryLevel == 0 && !fullyDischargedCar {
				fullyDischargedCar = true
				speed = types.CarSpeedWhenFullyDischarge
				fmt.Println("O carro descarregou, o motorista desceu e começou a empurrar o carro")
			}
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
			} else if coordY < stationY {
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
			totalMetersTraveled += speed
			if car.BatteryLevel > 0 && totalMetersTraveled > types.BatteryDischarge {
				totalMetersTraveled -= types.BatteryDischarge
				car.BatteryLevel -= 1
			}

			car.CoordX = coordX
			car.CoordY = coordY

			fmt.Printf("Carro está em (%d, %d)\n", car.CoordX, car.CoordY)
			fmt.Printf("Bateria: %d%% \n", car.BatteryLevel)

			if coordX == stationX && coordY == stationY {
				fmt.Printf("Carro %d chegou nas coordenadas: (%d, %d).\n", car.CarID, x, y)
				client.car = car
				return car
			}
		}
	}
}

func HandleStartCarMovement(client *CarClient, responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("A caminho da estação:", responseMessage.Station.StationID)
		car := MoveCarTo(client, responseMessage.Station.CoordX, responseMessage.Station.CoordY)
		return car, nil
	} else {
		return responseMessage.Car, fmt.Errorf(responseMessage.Err)
	}
}
