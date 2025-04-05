FROM golang:1.22

WORKDIR /app

COPY ./client_station /app

COPY ./types /app/types

COPY go.mod /app/

RUN go mod download

RUN go build -o client_station main.go func.go

CMD ["./client_station"]