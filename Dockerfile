FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOPROXY https://goproxy.cn,direct

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /home/metrics-go

ADD go.mod .
ADD go.sum .
RUN go mod tidy
COPY . .
RUN go build -o app .


EXPOSE 2112
CMD ["./app"]