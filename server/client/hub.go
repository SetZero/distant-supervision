package client

import (
	"../logger"
	"../messages"
	"encoding/json"
	"fmt"
)

type RoomMessage struct {
	room             string
	broadcastMessage []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	rooms map[string]*RoomInfo

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
		rooms:      make(map[string]*RoomInfo),
	}
}

func (h *Hub) Run() {
	agg := make(chan RoomMessage)
	for {
		select {
		case roomJoin := <-h.register:
			func() {
				if h.rooms[roomJoin.roomId] == nil {
					h.rooms[roomJoin.roomId] = &RoomInfo{roomName: roomJoin.roomId, clients: make(map[*Client]bool), streamer: nil}
					h.broadcast[roomJoin.roomId] = make(chan []byte)
					h.rebuildChannelAggregator(&agg)
				}

				fmt.Printf("Using: %p\n", &h.rooms[roomJoin.roomId].mu)
				defer h.rooms[roomJoin.roomId].mu.Unlock()
				defer fmt.Println("Freed Lock...")
				h.rooms[roomJoin.roomId].mu.Lock()

				h.rooms[roomJoin.roomId].clients[roomJoin.client] = true
				m, _ := json.Marshal(StreamerMessage{RoomHasStreamer: h.rooms[roomJoin.roomId].streamer != nil})
				sendMessageWrapper(roomJoin.client.conn, MessageWrapper{Type: messages.JoinRoomSuccessType, Message: m})
			}()
		case client := <-h.unregister:
			if _, ok := h.rooms[client.room]; !ok {
				break
			}
			if _, ok := h.rooms[client.room].clients[client]; ok {
				if h.rooms[client.room].streamer == client {
					h.rooms[client.room].streamer = nil
					m, _ := json.Marshal(StreamerMessage{RoomHasStreamer: false})
					for clients := range h.rooms[client.room].clients {
						sendMessageWrapper(clients.conn, MessageWrapper{Type: messages.JoinRoomSuccessType, Message: m})
					}

				}
				delete(h.rooms[client.room].clients, client)
				close(client.send)
				logger.InfoLogger.Printf("Client left. Clients left: %d\n", len(h.rooms[client.room].clients))
				if len(h.rooms[client.room].clients) == 0 {
					delete(h.rooms, client.room)
					logger.InfoLogger.Printf("Removed Room: %s, remaining rooms: %d\n", client.room, len(h.rooms))
				}
			}
		case roomMessages := <-agg:
			for client := range h.rooms[roomMessages.room].clients {
				select {
				case client.send <- roomMessages.broadcastMessage:
				default:
					close(client.send)
					delete(h.rooms[roomMessages.room].clients, client)
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
