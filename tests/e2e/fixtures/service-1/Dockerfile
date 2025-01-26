FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go build -o main cmd/service-1/main.go

CMD ["./main"]
