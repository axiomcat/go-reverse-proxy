FROM golang:1.23-bookworm AS base
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main
CMD ["/build/main"]
