build_vehicle_client:
	sudo docker build -t vehicle_cliente -f vehicle_client.dockerfile .
run_vehicle_client:
	sudo docker run -it --network="host" vehicle_cliente
build_post_client:
	sudo docker build -t post_cliente -f post_client.dockerfile .
run_post_client:
	sudo docker run -it --network="host" post_cliente