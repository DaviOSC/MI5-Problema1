package main

import (
	"fmt"
	"log"
	"main/types"
	"net"
)

func NewStationClient() *StationClient {
	// Conectar ao servidor
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Erro ao conectar ao servidor:", err)
	}

	return &StationClient{
		conn: conn,
	}
}

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
		Req:     types.RegisterStation,
		Station: station,
	}, nil

}

func RegisterStationResponse(responseMessage types.Message) (types.Station, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Estação registrada com sucesso.")
		return responseMessage.Station, nil
	} else {
		return responseMessage.Station, fmt.Errorf("erro ao registrar estação")
	}
}

func ChooseStation(station types.Station) (types.Message, error) {
	return types.Message{
		Req:     types.ListStations,
		Station: station,
	}, nil
}

func SelectStation(station types.Station) (types.Message, error) {
	return types.Message{
		Req:     types.SelectStation,
		Station: station,
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

func ChooseStationResponse(responseMessage types.Message) (types.Station, error) {
	if responseMessage.Status != types.Success {
		return types.Station{}, fmt.Errorf("falha ao listar estações: %v", responseMessage.Err)
	}

	// Exibir estações disponíveis
	fmt.Println("Estações disponíveis:")
	for _, station := range responseMessage.StationList {
		fmt.Printf("ID: %d, Coordenadas: (%d, %d)\n", station.StationID, station.CoordX, station.CoordY)
	}

	// Escolher uma estação
	var stationID int
	fmt.Print("Escolha o ID da estação: ")
	fmt.Scanln(&stationID)

	for _, station := range responseMessage.StationList {
		if station.StationID == stationID {
			fmt.Printf("Estação escolhida: ID %d, Coordenadas: (%d, %d)\n", station.StationID, station.CoordX, station.CoordY)
			return station, nil
		}
	}

	return types.Station{}, fmt.Errorf("estação inválida")
}

func (c *StationClient) HandleGetStationInfo(message types.Message, station types.Station) types.Message {
	return types.Message{
		Req:     types.GetStationInfo,
		Station: station,
	}
}

func (c *StationClient) HandleGetReservedStation(message types.Message) types.Message {
	return types.Message{
		Req: types.GetReservedStation,
	}
}

func (c *StationClient) HandleStartRecharge(message types.Message, station types.Station) types.Message {
	return types.Message{
		Req:     types.StartRecharge,
		Station: station,
	}
}
