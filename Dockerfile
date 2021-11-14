FROM golang:1.17-alpine as builder

WORKDIR /app
COPY . .
RUN go get -t .
RUN go build -o ./bin/app
WORKDIR /app/bin
RUN chmod +x app

FROM alpine
COPY --from=builder /app/bin/app ./app
CMD ./app -p $PORT
