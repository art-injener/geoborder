FROM golang:1.16-alpine3.14 AS builder

RUN go version

COPY . /github.com/X-Keeper/geoborder
WORKDIR /github.com/X-Keeper/geoborder

RUN go mod download

RUN GOOS=linux  go build -o ./bin/server ./cmd/rtree/main.go

FROM alpine:latest
LABEL maintainer="Artem Danilchenko <danilchenko.a@x-keeper.ru>"

WORKDIR /root/

COPY --from=0 /github.com/X-Keeper/geoborder/bin/server .
COPY --from=0 /github.com/X-Keeper/geoborder/configs configs/

RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime

EXPOSE 7071

ENTRYPOINT ["./server"]
