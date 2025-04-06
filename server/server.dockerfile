FROM golang:1.22

WORKDIR /app

COPY ./server /app/server

COPY ./types /app/types

COPY ./data /app/data

COPY go.mod /app/

RUN go mod download

RUN go build -o server.bin server/main.go server/database_funcs.go server/car_req_handles.go server/station_req_handles.go

EXPOSE 8080

CMD ["./server/server.bin"]