build_server:
	sudo docker build -t servidor -f server/servidor.dockerfile .
run_server: 
	sudo docker run -p 8080:8080 servidor
build_client:
	sudo docker build -t cliente -f client.dockerfile .
run_client:
	sudo docker run -it --network="host" cliente
all: 
	make build_server
	make run_server
	make build_server
	make run_server