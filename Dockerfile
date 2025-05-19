FROM golang:1.23

WORKDIR /app 

COPY ./ /app

EXPOSE 3200

RUN go build -o main ./cmd/server

CMD ["./main"]
