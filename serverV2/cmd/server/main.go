package main

import (
	"flag"
	"github.com/SetZero/distant-supervision/pkg/data"
	"github.com/SetZero/distant-supervision/pkg/websocket"
	"net/url"
)


func main() {
	var addr = flag.String("addr", "localhost:8080", "http service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}

	socketServer := websocket.NewServer()
	hub := data.NewHub()
	hub.WaitForUser(socketServer.JoinChannel)

	go socketServer.Start(u)
}