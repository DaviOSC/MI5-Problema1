FROM golang:latest
WORKDIR /app
COPY go.mod .
COPY post_client.go .
RUN go build -o post_client .
CMD ["./post_client"]