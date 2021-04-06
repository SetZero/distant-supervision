package websocket

import (
	"github.com/SetZero/distant-supervision/pkg/logger"
	"github.com/gorilla/websocket"
	"time"
)

type Connection struct {
	conn *websocket.Conn
	Input chan []byte
	Output chan []byte

	/* Client connection Handling */
}

func (c Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case _, ok := <-c.Output:
			// todo
			if !ok {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c Connection) readPump() {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Info.Println("Error while processing request: ", err)
			}
			break
		}
		logger.Info.Println("Incoming Message: ", message)
	}
}

func NewConnection(conn *websocket.Conn) *Connection {
	internalConnection := &Connection{conn: conn}
	go internalConnection.writePump()
	go internalConnection.readPump()

	return internalConnection
}