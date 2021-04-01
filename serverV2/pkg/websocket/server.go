package websocket

import (
	"github.com/SetZero/distant-supervision/pkg/logger"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	JoinChannel chan *Connection
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
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

	s.JoinChannel <- &Connection{conn: conn}
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() {
	if s.JoinChannel != nil {
		panic("Start() should only be called once!")
	}

	defer close(s.JoinChannel)
	s.JoinChannel = make(chan *Connection)

	logger.Info.Println("Started Websocket Server")
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.serveWs(w, r)
	})
	err := http.ListenAndServe(":5501", nil)
	if err != nil {
		logger.Error.Println("ListenAndServe: ", err)
	}
}