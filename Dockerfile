#docker build -t go-debit .
#docker run -dit --name go-debit -p 5000:5000 go-debit

FROM golang:1.21 As builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-debit -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-debit .

CMD ["/app/go-debit"]