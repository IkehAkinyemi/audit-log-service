FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -o=/log-service ./cmd/service

EXPOSE 8081

RUN chmod 666 -R /var/

CMD [ "/log-service" ]

