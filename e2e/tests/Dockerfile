FROM golang:1.25

WORKDIR /tests

COPY go.mod ./

RUN go mod download

COPY . .

CMD ["go", "test", "./...", "-v"]
