# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

RUN apk add curl

WORKDIR /app

RUN mkdir location 
RUN mkdir task 
RUN mkdir stream

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY location/*.go ./location
COPY task/*.go ./task
COPY stream/*.go ./stream

RUN go build -o /task_manager

EXPOSE 8080

CMD [ "/task_manager" ]