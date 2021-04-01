package data

import (
	"github.com/SetZero/distant-supervision/pkg/websocket"
)

type Hub struct {
}

func (h Hub) WaitForUser(channel chan *websocket.Connection) {

}

func NewHub() *Hub {
	return &Hub{}
}