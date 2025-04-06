package main

import (
	"fmt"
	"main/types"
	"slices"
	"time"
)

func RegisterStation() (types.Message, error) {
	// Registrar uma estação
	station := types.Station{}
	fmt.Print("Digite o ID da estação: ")
	fmt.Scanln(&station.StationID)
	fmt.Print("Digite a coordenada X da estação: ")
	fmt.Scanln(&station.CoordX)
	fmt.Print("Digite a coordenada Y da estação: ")
	fmt.Scanln(&station.CoordY)

	return types.Message{
		Req:        types.RegisterStation,
		Station:    station,
		ClientType: types.StationClientType,
	}, nil

}

func RegisterStationResponse(responseMessage types.Message) (types.Station, error) {
	if responseMessage.Status == types.Success {
		return responseMessage.Station, nil
	} else {
		return responseMessage.Station, fmt.Errorf(responseMessage.Err)
	}
}

func (c *StationClient) HandleRegisterStation() error {
	// Gerar mensagem com requisição para registrar posto
	message, err := RegisterStation()
	if err != nil {
		return err
	}
	// Enviar mensagem
	err = c.SendMessage(message)
	if err != nil {
		return err
	}

	// Receber a resposta do servidor
	responseMessage, err := c.ReadResponse()
	if err != nil {
		return err
	}

	// Tratar resposta do servidor
	_, err = RegisterStationResponse(responseMessage)
	if err != nil {
		return err
	}
	return nil
}

func HandleListStations(station types.Station) (types.Message, error) {
	return types.Message{
		Req:        types.ListStations,
		Station:    station,
		ClientType: types.StationClientType,
	}, nil
}

func SelectStation(station types.Station) (types.Message, error) {
	return types.Message{
		ClientType: types.StationClientType,
		Req:        types.SelectStation,
		Station:    station,
	}, nil
}
func SelectStationResponse(responseMessage types.Message) (types.Station, error) {
	if responseMessage.Status == types.Success {
		fmt.Printf("Estação adicionada com sucesso: ID %d, Coordenadas: (%d, %d)\n",
			responseMessage.Station.StationID, responseMessage.Station.CoordX, responseMessage.Station.CoordY)
		return responseMessage.Station, nil
	} else {
		return types.Station{}, fmt.Errorf("erro ao adicionar estação: %v", responseMessage.Err)
	}
}

func HandleListStationsResponse(responseMessage types.Message) (types.Station, error) {
	if responseMessage.Status != types.Success {
		return types.Station{}, fmt.Errorf(responseMessage.Err)
	}

	// Exibir estações disponíveis
	fmt.Println("Estações disponíveis:")
	for _, station := range responseMessage.StationList {
		fmt.Printf("ID: %d, Coordenadas: (%d, %d)\n", station.StationID, station.CoordX, station.CoordY)
	}

	// Escolher uma estação
	var stationID int
	for {
		fmt.Print("Escolha o ID da estação: ")
		fmt.Scanln(&stationID)
		if stationID > 0 {
			break
		}
		fmt.Println("ID inválido")
	}

	for _, station := range responseMessage.StationList {
		if station.StationID == stationID {
			fmt.Printf("Estação escolhida: ID %d, Coordenadas: (%d, %d)\n", station.StationID, station.CoordX, station.CoordY)
			return station, nil
		}
	}

	return types.Station{}, fmt.Errorf("estação inválida")
}

func (c *StationClient) HandleListAndChooseStation() (types.Station, error) {
	message := types.Message{Req: types.ListStations}
	station := types.Station{}

	// Enviar a mensagem para listar estações
	err := c.SendMessage(message)
	if err != nil {
		return station, err
	}

	// Receber a resposta do servidor
	responseMessage, err := c.ReadResponse()
	if err != nil {
		return station, err
	}

	// Processar a resposta e permitir que o usuário escolha uma estação
	station, err = HandleListStationsResponse(responseMessage)
	if err != nil {
		return station, err
	}

	// Informar ao servidor qual estação foi escolhida
	message = types.Message{
		Req:     types.SelectStation,
		Station: station,
	}

	err = c.SendMessage(message)
	if err != nil {
		return station, err
	}

	responseMessage, err = c.ReadResponse()
	if err != nil {
		return station, err
	}

	station, err = SelectStationResponse(responseMessage)

	return station, err
}

func (c *StationClient) HandleGetStationInfo(message types.Message) types.Message {
	return types.Message{
		Req:     types.StationUpdate,
		Station: c.station,
	}
}

func (c *StationClient) HandleReserveStation(message types.Message) types.Message {
	responseMessage := types.Message{Req: types.ReserveStation}
	car := message.Car

	if slices.Contains(c.station.CarList, car.CarID) {
		responseMessage.Status = types.Error
		responseMessage.Err = "Carro já está na fila da estação"
		responseMessage.Car = car
		responseMessage.Station = c.station
		return responseMessage
	}

	c.station.CarList = append(c.station.CarList, car.CarID)
	c.station.CarsWaiting += 1
	car.ReservedStation = c.station.StationID
	responseMessage.Status = types.Success
	responseMessage.Car = car
	responseMessage.Station = c.station

	return responseMessage
}

func (c *StationClient) HandleExit() types.Message {
	return types.Message{
		Req:     types.ExitStation,
		Station: c.station,
	}
}

func (c *StationClient) HandleStartRecharge(message types.Message) types.Message {
	responseMessage := types.Message{
		Req:     types.StartRecharge,
		Status:  types.Success,
		Station: c.station,
		Car:     message.Car,
	}
	if !message.Car.PaidReservedStation {
		responseMessage.Status = types.Error
		responseMessage.Err = "É necessario pagar antes de iniciar a recarga"
		return responseMessage
	} else {
		return responseMessage
	}
}

func (c *StationClient) HandleRechargeComplete(message types.Message) types.Message {
	responseMessage := types.Message{
		Req:     types.RechargeComplete,
		Status:  types.Success,
		Station: c.station,
		Car:     message.Car,
	}

	if message.Car.CarID != c.station.InUseBy {
		responseMessage.Status = types.Error
		responseMessage.Err = "ID do carro não corresponde ao carro na estação"
		return responseMessage
	}
	c.station.InUseBy = 0
	responseMessage.Car.BatteryLevel = 100
	responseMessage.Car.ReservedStation = 0
	responseMessage.Car.RecomendedStation = 0
	responseMessage.Car.PaidReservedStation = false
	return responseMessage
}

func (c *StationClient) HandlePayRecharge(message types.Message) types.Message {
	responseMessage := types.Message{
		Req:     types.PayRecharge,
		Status:  types.Success,
		Car:     message.Car,
		Station: c.station}

	if message.Car.CoordX != c.station.CoordX || message.Car.CoordY != c.station.CoordY {
		responseMessage.Status = types.Error
		responseMessage.Err = "Carro não está na estação"
		return responseMessage
	}

	if c.station.InUseBy != 0 {
		responseMessage.Status = types.Error
		responseMessage.Err = "o posto está em uso"
		return responseMessage
	}

	if c.station.CarList[0] != message.Car.CarID {
		responseMessage.Status = types.Error
		responseMessage.Err = "Carro não é o proximo da fila"
		return responseMessage
	}
	paymentIDs += 1

	payment := types.Payment{
		PaymentID: paymentIDs,
		From:      c.station.StationID,
		To:        message.Car.CarID,
		Value:     100.0,
		TimeStamp: time.Now(),
	}

	if message.Car.PixCode != 0 {
		responseMessage.Car = c.DoPaymentPix(message.Car, payment)
	} else if message.Car.CreditCardNumber != 0 {
		responseMessage.Car = c.DoPaymentCreditCard(message.Car, payment)
	} else {
		responseMessage.Status = types.Error
		responseMessage.Err = "Nenhuma forma de pagamento cadastrada"
		return responseMessage
	}

	return responseMessage
}

func (c *StationClient) DoPaymentPix(car types.Car, payment types.Payment) types.Car {
	car.PaymentHistory = append(car.PaymentHistory, payment)
	car.PaidReservedStation = true
	c.station.CarsWaiting -= 1
	c.station.CarList = c.station.CarList[1:]
	c.station.InUseBy = car.CarID
	return car
}

func (c *StationClient) DoPaymentCreditCard(car types.Car, payment types.Payment) types.Car {
	car.PaymentHistory = append(car.PaymentHistory, payment)
	car.PaidReservedStation = true
	c.station.CarsWaiting -= 1
	c.station.CarList = c.station.CarList[1:]
	c.station.InUseBy = car.CarID
	return car
}
