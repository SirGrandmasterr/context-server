# builder image
FROM golang:1.22.3-alpine3.20 as builder
RUN mkdir /build
COPY . /build/
WORKDIR /build
#RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o llm-whisperer .


# generate clean, final image for end users
FROM alpine
COPY --from=builder /build/.env /
COPY --from=builder /build/llm-whisperer /

# executable
ENTRYPOINT [ "/llm-whisperer", "websocket" ] 

