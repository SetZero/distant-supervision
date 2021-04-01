package main

import (
	"github.com/SetZero/distant-supervision/pkg/data"
	"github.com/SetZero/distant-supervision/pkg/websocket"
)


func main() {
	socketServer := websocket.NewServer()
	hub := data.NewHub()
	hub.WaitForUser(socketServer.JoinChannel)

	go socketServer.Start()
}