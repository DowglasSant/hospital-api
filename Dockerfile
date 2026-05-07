FROM golang:1.25.3

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o hospital-api .

EXPOSE 8080

CMD ["./hospital-api"]