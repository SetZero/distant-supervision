package client

import (
	"../messages"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type ErrorType string

const(
	Unknown ErrorType = "Unknown"
	InvalidMessage = "Invalid Message"
)

type MessageWrapper struct {
	Type messages.MessageType `json:"type"`
	Message json.RawMessage `json:"message"`
}

type JoinRoomMessage struct {
	RoomId string `json:"roomId"`
}

type ErrorMessage struct {
	Type ErrorType
	Description string
}

type StreamerMessage struct {
	RoomHasStreamer bool `json:"roomHasStreamer"`
}

type ViewerMessage struct {
	Viewers uint32 `json:"viewers"`
}

type StartStreamInfoMessage struct {
	StreamStartSuccess bool `json:"startStreamSuccess"`
}

func (e *ErrorMessage) writeError(conn *websocket.Conn) {
	m, _  := json.Marshal(e)
	sendMessageWrapper(conn, MessageWrapper{Type: messages.ErrorMessageType, Message: m})
}

func sendMessageWrapper(conn *websocket.Conn, message MessageWrapper) {
	m, err  := json.Marshal(message)
	if err != nil {
		fmt.Println("Error during error:", err)
	} else {
		conn.WriteMessage(websocket.TextMessage, m)
	}
}