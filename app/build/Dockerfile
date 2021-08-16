# Building stage
FROM golang:1.15-alpine3.12 as builder

LABEL MAINTAINER="Yalochat <eng@yalochat.com>"

WORKDIR /go/src/yalochat.com/salesforce-integration

COPY go.sum go.mod ./

RUN go mod download

COPY base base

COPY app app

RUN apk add build-base 

RUN go build -o /salesforce-integration app/main.go 


# Final image
FROM alpine:3.12

LABEL MAINTAINER="Yalochat <eng@yalochat.com>"

COPY --from=builder /salesforce-integration .

CMD ["./salesforce-integration"]
