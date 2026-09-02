FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o stock .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/stock .

EXPOSE 8081

CMD ["./stock"]