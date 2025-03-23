package types

type Payment struct {
	PaymentID int `json:"payment_id"`
	From      int `json:"from"` // CarID que pagou a estação
	To        int `json:"to"`   // StationID que recebeu o pagamento
	Value     int `json:"value"`
	TimeStamp int `json:"timestamp"`
}

type Car struct {
	CarID             int    `json:"car_id"`
	User              string `json:"user"`
	Password          string `json:"password"`
	CoordX            int    `json:"coord_x"`
	CoordY            int    `json:"coord_y"`
	BatteryLevel      int    `json:"battery_level"`      // 0-100
	RecomendedStation int    `json:"recomended_station"` // StationID
	ReservedStation   int    `json:"reserved_station"`   // StationID
	PaymentHistory    []int  `json:"payment_history"`    // Slice de PaymentID
}

type Station struct {
	StationID  int   `json:"station_id"`
	CoordX     int   `json:"coord_x"`
	CoordY     int   `json:"coord_y"`
	CarsInLine []int `json:"cars_in_line"` // Slice de CarID
}

type Requests int

const (
	RegisterCar           Requests = iota // Registrar carro, passando ID, coordenadas e nível da bateria: carro -> server
	RegisterStation                       // Registrar estação, passando ID e coordenadas: estação -> server
	GetRecommendedStation                 // Obter a estação recomendada para um carro, passando suas coordenadas: carro -> server
	IsStationAvailable                    // Verificar se uma estação está disponível, passando seu ID: server -> estações
	ReserveStation                        // Reservar uma estação, passando o ID do carro e o ID da estação: carro -> server
	RechargeCar                           // Recarregar um carro, passando o ID do carro e o ID da estação: carro -> server
	GeneratePayment                       // Gerar um pagamento, passando o ID do pagamento e o valor: server -> carro
	PayRecharge                           // Pagar uma recarga, passando o ID do pagamento: carro -> server(Obs:fazer antes da reserva o pagamento)
	UserLogin
	ListStations
)

var RequestsNames = map[Requests]string{
	RegisterCar:           "register_car",
	RegisterStation:       "register_station",
	GetRecommendedStation: "get_recommended_station",
	IsStationAvailable:    "is_station_available",
	ReserveStation:        "reserve_station",
	RechargeCar:           "recharge_car",
	GeneratePayment:       "generate_payment",
	PayRecharge:           "pay_recharge",
	UserLogin:             "user_login",
	ListStations:          "list_stations",
}

type ResponseStatus int

const (
	Success ResponseStatus = iota
	Error
	Fatal
)

var ResposeStatusNames = map[ResponseStatus]string{
	Success: "success",
	Error:   "error",
	Fatal:   "fatal",
}

func (r ResponseStatus) String() string {
	return ResposeStatusNames[r]
}

func (r Requests) String() string {
	return RequestsNames[r]
}

type Message struct {
	Req         Requests       `json:"request,omitempty"`
	Status      ResponseStatus `json:"status,omitempty"`
	Car         Car            `json:"car,omitempty"`
	Station     Station        `json:"station,omitempty"`
	Payment     Payment        `json:"payment,omitempty"`
	CarList     []Car          `json:"car_list,omitempty"`
	StationList []Station      `json:"station_list,omitempty"`
	PaymentList []Payment      `json:"payment_list,omitempty"`
}
