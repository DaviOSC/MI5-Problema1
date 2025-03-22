package main

import (
	"encoding/json"
	"fmt"
	"main/types"
	"net"
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

func HandleListStations() (types.Message, error) {
	return types.Message{
		Req: types.ListStations,
	}, nil
}

// HandleReserveStation lida com a requisição de reserva de estação
func HandleReserveStation(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.ReserveStation,
		Car: car,
	}, nil
}

// HandleRechargeCar lida com a requisição de recarga de carro
func HandleRechargeCar(car types.Car) (types.Message, error) {
	return types.Message{
		Req: types.RechargeCar,
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
		return car, err
	}
	_, err = conn.Write(buf)
	if err != nil {
		fmt.Println("Erro ao enviar a mensagem de login:", err)
		return car, err
	}

	response, err := ReadResponse(conn)
	if err != nil {
		fmt.Println("Erro receber resposta do servidor:", err)
		return car, err
	}

	if response.Status != types.Success {
		fmt.Println("Erro no login:", response.Status)
	} else {
		data, ok := response.Data.([]byte)
		if !ok {
			fmt.Println("Erro ao converter os dados da resposta para []byte")
			return car, err
		}
		json.Unmarshal(data, &car)
		fmt.Println(response.Status.String())
	}
	return car, nil
}

func ReadResponse(conn net.Conn) (types.ResponseMessage, error) {
	decoder := json.NewDecoder(conn)
	responseMessage := types.ResponseMessage{}
	err := decoder.Decode(&responseMessage)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return responseMessage, err
	}
	return responseMessage, nil
}
