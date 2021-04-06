package websocket

import (
	"fmt"
	"github.com/SetZero/distant-supervision/pkg/logger"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
)

type Server struct {
	JoinChannel chan *Connection
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

// serveWs handles websocket requests from the peer.
func (s *Server) serveWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	logger.Info.Println("New User Joined")
	s.JoinChannel <- NewConnection(conn)
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start(addr url.URL) {
	if s.JoinChannel != nil {
		panic("Start() should only be called once!")
	}
	if addr.Scheme != "ws" && addr.Scheme != "wss" {
		panic(fmt.Sprintf("Scheme should be wss or ws, but is: %s", addr.Scheme))
	}

	defer close(s.JoinChannel)
	s.JoinChannel = make(chan *Connection)

	logger.Info.Println("Started Websocket Server")
	http.HandleFunc("/", serveHome)
	http.HandleFunc(addr.Path, func(w http.ResponseWriter, r *http.Request) {
		s.serveWs(w, r)
	})
	err := http.ListenAndServe(addr.Host, nil)
	if err != nil {
		logger.Error.Println("ListenAndServe: ", err)
	}
}