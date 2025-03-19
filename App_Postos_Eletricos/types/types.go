package types

type Payment struct {
	PaymentID int `json:"payment_id"`
	From      int `json:"from"` // CarID que pagou a estação
	To        int `json:"to"`   // StationID que recebeu o pagamento
	Value     int `json:"value"`
	TimeStamp int `json:"timestamp"`
}

type Car struct {
	CarID             int   `json:"car_id"`
	CoordX            int   `json:"coord_x"`
	CoordY            int   `json:"coord_y"`
	BatteryLevel      int   `json:"battery_level"`      // 0-100
	RecomendedStation int   `json:"recomended_station"` // StationID
	ReservedStation   int   `json:"reserved_station"`   // StationID
	PaymentHistory    []int `json:"payment_history"`    // Slice de PaymentID
}

type Station struct {
	StationID  int   `json:"station_id"`
	CoordX     int   `json:"coord_x"`
	CoordY     int   `json:"coord_y"`
	CarsInLine []int `json:"cars_in_line"` // Slice de CarID
}
