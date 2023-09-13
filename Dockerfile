FROM golang:1.18.4-alpine3.16

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /search-engine-indexer

ENTRYPOINT [ "/search-engine-indexer" ]