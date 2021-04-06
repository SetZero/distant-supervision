package websocket

import (
	"net/url"
	"testing"
	"time"
)
import "github.com/onsi/gomega"

func TestIllegalAddr(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	u := url.URL{Scheme: "http", Host: "localhost:8082", Path: "/echo"}

	s := NewServer()
	g.Expect(func () { s.Start(u) }).Should(gomega.PanicWith("Scheme should be wss or ws, but is: http"), "Expected Panic on illegal address, got none")
}

func TestMultipleStart(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s := NewServer()

	u := url.URL{Scheme: "ws", Host: "localhost:8082", Path: "/echo"}
	go s.Start(u)
	time.Sleep(400 * time.Millisecond)

	g.Expect(func () { s.Start(u) }).Should(gomega.PanicWith("Start() should only be called once!"), "Expected Panic on multiple start calls, got none")
}