FROM golang:1.23 AS builder

COPY ../go.mod go.sum ./

COPY . .

RUN go mod download

RUN go build -o app ./cmd/main.go

COPY ../scripts/wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

EXPOSE 8080

CMD ["./wait-for-it.sh", "db:5432", "--", "./app"]