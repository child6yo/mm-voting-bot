FROM golang:1.23

WORKDIR /votingbot

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o votingbot ./cmd/main.go

CMD [ "./votingbot" ]