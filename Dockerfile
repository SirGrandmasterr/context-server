FROM golang:1.22 as build
WORKDIR /src
COPY . .
RUN go build -o llm-whisperer ./main.go
ENTRYPOINT  ["/src/llm-whisperer", "websocket"]

