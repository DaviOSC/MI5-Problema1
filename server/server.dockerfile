FROM golang:1.22

WORKDIR /app

COPY ./server /app

COPY ./types /app/types

COPY ./data /app/data

COPY go.mod /app/

RUN go mod download

RUN go build -o server_bin main.go database_funcs.go car_req_handles.go station_req_handles.go

EXPOSE 8080

CMD ["./server_bin"]