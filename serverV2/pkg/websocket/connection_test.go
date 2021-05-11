package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/onsi/gomega"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestIncomingMessage(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	u := url.URL{Scheme: "ws", Host: "localhost:8082", Path: "/echo"}

	s := NewServer()
	go s.Start(u)

	fmt.Println("Connect to: ", u.String())
	c, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)
	defer c.Close()
	c.WriteMessage(websocket.TextMessage, []byte("ABC"))

	var wg sync.WaitGroup
	testChannel := make(chan []byte)

	wg.Add(1)
	go func() {
		defer wg.Done()

		client := <- s.JoinChannel
		testChannel<- <-client.Recv
	}()
	g.Eventually(testChannel, 2 * time.Second).Should(gomega.Receive(gomega.Equal([]byte("ABC"))))

	wg.Wait()
}