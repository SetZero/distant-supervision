package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"testing"
)

func TestIncomingMessage(t *testing.T) {
	// g := gomega.NewGomegaWithT(t)

	u := url.URL{Scheme: "ws", Host: "localhost:8082", Path: "/echo"}

	s := NewServer()
	go s.Start(u)

	fmt.Println("Connect to: ", u.String())
	c, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)
	defer c.Close()
	c.WriteMessage(websocket.TextMessage, []byte("ABC"))
}