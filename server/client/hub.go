package client

import "fmt"

type RoomMessage struct {
	room             string
	broadcastMessage []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients to rooms.
	broadcast map[string]chan []byte

	// Register requests from the clients.
	register chan *RoomJoin

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(map[string]chan []byte),
		register:   make(chan *RoomJoin),
		unregister: make(chan *Client),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	agg := make(chan RoomMessage)
	for {
		select {
		case roomJoin := <-h.register:
			if h.rooms[roomJoin.roomId] == nil {
				h.rooms[roomJoin.roomId] = make(map[*Client]bool)
				h.broadcast[roomJoin.roomId] = make(chan []byte)
				h.rebuildChannelAggregator(&agg)
				fmt.Println("Created New Room:", roomJoin.roomId)
			}
			h.rooms[roomJoin.roomId][roomJoin.client] = true
			fmt.Println("Two:", agg)
		case client := <-h.unregister:
			if _, ok := h.rooms[client.room][client]; ok {
				delete(h.rooms[client.room], client)
				close(client.send)
				fmt.Println("Left")
			}
		case roomMessages := <-agg:
			for client := range h.rooms[roomMessages.room] {
				select {
				case client.send <- roomMessages.broadcastMessage:
				default:
					close(client.send)
					delete(h.rooms[roomMessages.room], client)
				}
			}
		}
	}
}

func (h *Hub) rebuildChannelAggregator(agg *chan RoomMessage) {
	for roomId, ch := range h.broadcast {
		go func(c chan []byte, roomId string) {
			for msg := range c {
				*agg <- RoomMessage{roomId, msg}
			}
		}(ch, roomId)
	}
}
