package main

import (
	"encoding/json"
	"fmt"
	"main/types"
)

// HandleRegisterCar lida com a requisição de registro de carro
func HandleRegisterCar() (types.Message, types.Car, error) {
	var car types.Car
	fmt.Print("Digite o ID do carro: ")
	fmt.Scanln(&car.CarID)
	fmt.Print("Digite a coordenada X do carro: ")
	fmt.Scanln(&car.CoordX)
	fmt.Print("Digite a coordenada Y do carro: ")
	fmt.Scanln(&car.CoordY)
	fmt.Print("Digite o nível da bateria: ")
	fmt.Scanln(&car.BatteryLevel)

	data, err := json.Marshal(car)
	if err != nil {
		return types.Message{}, types.Car{}, err
	}

	return types.Message{
		Req:  types.RegisterCar,
		Data: data,
	}, car, nil
}

// HandleGetRecommendedStation lida com a requisição de recomendação de estação
func HandleGetRecommendedStation(car types.Car) (types.Message, error) {
	data, err := json.Marshal(car)
	if err != nil {
		return types.Message{}, err
	}
	return types.Message{
		Req:  types.GetRecommendedStation,
		Data: data,
	}, nil
}

// HandleReserveStation lida com a requisição de reserva de estação
func HandleReserveStation(car types.Car) (types.Message, error) {
	data, err := json.Marshal(car)
	if err != nil {
		return types.Message{}, err
	}
	return types.Message{
		Req:  types.ReserveStation,
		Data: data,
	}, nil
}

// HandleRechargeCar lida com a requisição de recarga de carro
func HandleRechargeCar(car types.Car) (types.Message, error) {
	data, err := json.Marshal(car)
	if err != nil {
		return types.Message{}, err
	}
	return types.Message{
		Req:  types.RechargeCar,
		Data: data,
	}, nil
}

// HandlePayRecharge lida com a requisição de pagamento de recarga
func HandlePayRecharge(car types.Car) (types.Message, error) {
	data, err := json.Marshal(car)
	if err != nil {
		return types.Message{}, err
	}
	return types.Message{
		Req:  types.PayRecharge,
		Data: data,
	}, nil
}
