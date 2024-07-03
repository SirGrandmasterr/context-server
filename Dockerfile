FROM golang:1.22 as build
WORKDIR /src
COPY . .
RUN go build -o /bin/llm-whisperer ./main.go
ENTRYPOINT  ["/bin/llm-whisperer", "http"]

