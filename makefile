build_server:
	sudo docker build -t servidor -f server/server.dockerfile .
run_server: 
	sudo docker run -p 8080:8080 servidor
build_client_car:
	sudo docker build -t client_car -f client_car/client_car.dockerfile .
build_client_station:
	sudo docker build -t client_station -f client_station/client_station.dockerfile .
run_client_car:
	sudo docker run -it --network="host" client_car
run_client_station:
	sudo docker run -it --network="host" client_station
all_server: 
	make build_server
	make run_server
all_client_car: 
	make build_client_car
	make run_client_car
all_client_station: 
	make build_client_station
	make run_client_station
all: 
	make build_server
	make run_server
	make build_client_car
	make run_client_car
	make build_client_station
	make run_client_station
