# Building stage
FROM --platform=linux/amd64 golang:1.15-alpine3.12 as builder

LABEL MAINTAINER="Yalochat <eng@yalochat.com>"

WORKDIR /go/src/yalochat.com/salesforce-integration

COPY go.sum go.mod ./

RUN GOOS=linux GOARCH=386 go mod download

COPY base base

COPY app app

RUN apk add build-base

RUN GOOS=linux GOARCH=386 go build -o /salesforce-integration app/main.go


# Final image
FROM --platform=linux/x86_64 alpine:3.12

LABEL MAINTAINER="Yalochat <eng@yalochat.com>"

ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /zoneinfo.zip
ENV ZONEINFO /zoneinfo.zip

COPY --from=builder /salesforce-integration .

CMD ["./salesforce-integration"]
