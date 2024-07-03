FROM golang:1.22 as build
WORKDIR /src
COPY . .
RUN go build -o /bin/llmwhisperer ./main.go

FROM scratch
COPY --from=build /bin/llmwhisperer /bin/llmwhisperer
CMD ["/bin/llmwhisperer", "http"]