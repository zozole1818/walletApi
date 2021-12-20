FROM golang:1.17.5-alpine

RUN apk add bash
RUN apk add postgresql-client

COPY ./. /app
RUN go get github.com/swaggo/swag/cmd/swag@v1.7.6

WORKDIR /app
RUN swag init -d cmd/,controller/,model/

WORKDIR /app/cmd
RUN go build -o walletApi

EXPOSE 8000

CMD ["/app/devops/web/entrypoint.sh"]
