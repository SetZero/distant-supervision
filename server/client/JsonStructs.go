package client

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type MessageType string

const(
	ErrorMessageType MessageType = "error"
	JoinMessageType = "joinMessage"
	JoinRoomSuccessType = "joinedMessage"
)

type ErrorType string

const(
	Unknown ErrorType = "Unknown"
	InvalidMessage = "Invalid Message"
)

type MessageWrapper struct {
	Type MessageType `json:"type"`
	Message json.RawMessage `json:"message"`
}

type JoinRoomMessage struct {
	RoomId string `json:"roomId"`
}

type ErrorMessage struct {
	Type ErrorType
	Description string
}

func (e *ErrorMessage) writeError(conn *websocket.Conn) {
	m, _  := json.Marshal(e)
	sendMessageWrapper(conn, MessageWrapper{Type: ErrorMessageType, Message: m})
}

func sendMessageWrapper(conn *websocket.Conn, message MessageWrapper) {
	m, err  := json.Marshal(message)
	if err != nil {
		fmt.Println("Error during error:", err)
	} else {
		conn.WriteMessage(websocket.TextMessage, m)
	}
}