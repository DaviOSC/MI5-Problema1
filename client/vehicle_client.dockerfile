FROM golang:latest
WORKDIR /app
COPY go.mod .
COPY vehicle_client.go .
RUN go build -o vehicle_client .
CMD ["./vehicle_client"]