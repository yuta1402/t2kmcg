FROM golang:1.14-alpine as builder

ENV CGO_ENABLED 0

WORKDIR /go/src/github.com/yuta1402/t2kmpg

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

COPY . .

RUN go build -a -o t2kmpg .


FROM alpine:latest

WORKDIR /root

RUN apk add --no-cache \
    udev \
    ttf-freefont \
    chromium \
    chromium-chromedriver

COPY --from=builder /go/src/github.com/yuta1402/t2kmpg/t2kmpg .

CMD ["./t2kmpg"]
