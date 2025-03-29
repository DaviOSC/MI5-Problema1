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
		fmt.Println("Erro ao decodificar JSON:", err)
		return responseMessage, err
	}
	return responseMessage, nil
}

func HandleRegisterCarResponse(responseMessage types.Message) (types.Car, error) {
	if responseMessage.Status == types.Success {
		fmt.Println("Carro registrado com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro ao registrar carro")
	}
}
func HandlePaymentHistoryResponse(responseMessage types.Message) (types.Car, error) {
    if responseMessage.Status == types.Success {
        fmt.Println("Histórico de pagamentos:")
        for _, payment := range responseMessage.Car.PaymentHistory {
            fmt.Printf(
                `ID do pagamento: %d
De (CarID): %d
Para (StationID): %d
Valor: %d
Timestamp: %d`, payment.PaymentID, payment.From, payment.To,
				payment.Value, payment.TimeStamp,
            )
        }
        return responseMessage.Car, nil
    } else {
        return responseMessage.Car, fmt.Errorf("erro ao listar pagamentos")
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
        return responseMessage.Car, fmt.Errorf("erro ao recomendar estação")
    }
}
func HandleReserveStationResponse(responseMessage types.Message) (types.Car, error){
	if responseMessage.Status == types.Success {
		fmt.Println("Estação reservada com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro ao reservar estação")
	}
}

func HandleRechargeCompleteResponse(responseMessage types.Message) (types.Car, error){
	if responseMessage.Status == types.Success {
		fmt.Println("Recarga completa com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro ao completar recarga")
	}
}
func HandleStartRechargeResponse(client *Client, responseMessage types.Message) (types.Car, error) {
    if responseMessage.Status == types.Success {
        fmt.Println("Recarga iniciada com sucesso.")

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
        return responseMessage.Car, fmt.Errorf("erro ao iniciar recarga: %v", responseMessage.Status)
    }
}

func HandlePayRechargeResponse(responseMessage types.Message) (types.Car, error){
	if responseMessage.Status == types.Success {
		fmt.Println("Recarga paga com sucesso.")
		return responseMessage.Car, nil
	} else {
		return responseMessage.Car, fmt.Errorf("erro ao pagar recarga")
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

// Monitoramento da bateria
func (c *Client) batteryMonitor() {
    for {
        select {
        case <-c.batteryTicker.C:
            c.batteryMutex.Lock()
            if c.car.BatteryLevel > 0 {
                c.car.BatteryLevel--
                fmt.Printf("🔋 Bateria atual: %d%%\n", c.car.BatteryLevel)
                
                // Alerta aos 15%
                if c.car.BatteryLevel == 15 {
                    fmt.Println("⚠️  Bateria crítica! Procure uma estação!")
                }
            }
            c.batteryMutex.Unlock()
        
        case <-c.syncTicker.C:
            c.syncBatteryWithServer()
        }
    }
}

// Sincroniza com servidor
func (c *Client) syncBatteryWithServer() {
    c.batteryMutex.Lock()
    defer c.batteryMutex.Unlock()

    msg := types.Message{
        Req: types.BatterySync,
        Car: c.car,
    }

    if err := c.SendMessage(msg); err != nil {
        fmt.Println("Erro ao sincronizar bateria:", err)
        return
    }

    fmt.Println("🔄 Sincronizando bateria com servidor...")
}
