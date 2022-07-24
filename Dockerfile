FROM golang:1.18-alpine3.16 AS build
RUN apk --no-cache add ca-certificates
WORKDIR /telegram-webhook
COPY . .
RUN go build -o main

FROM alpine:3.6 AS run
WORKDIR /bin
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /telegram-webhook/main /bin/main
ENTRYPOINT ["/bin/main"]