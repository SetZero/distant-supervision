package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"./client"
)

var addr = flag.String("addr", ":5501", "http service address")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  16384,
	WriteBufferSize: 16384,
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Called Home")
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
func serveWs(hub *client.Hub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := client.NewClient(hub, conn)
	// client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	// go client.writePump()
	go client.ReadPump()
}

func main() {
	fmt.Println("Started!")
	flag.Parse()
	hub := client.NewHub()
	go hub.Run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(":5501", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}