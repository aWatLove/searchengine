FROM golang:latest

WORKDIR /app

COPY . .

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN go build cmd/main.go

RUN chmod +x main

CMD ["./main"]