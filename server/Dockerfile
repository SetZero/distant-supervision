FROM golang:1.15.6
WORKDIR /workdir
COPY . .
RUN go get -d -v github.com/gorilla/websocket \
  && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o backend .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /workdir/backend .

EXPOSE 5501
CMD ["./backend"]