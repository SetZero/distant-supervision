package websocket

import (
	"testing"
	"time"
)
import "github.com/onsi/gomega"

func TestMultipleStart(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s := NewServer()

	go s.Start()
	time.Sleep(400 * time.Millisecond)

	g.Expect(s.Start).Should(gomega.PanicWith("Start() should only be called once!"), "Expected Panic on multiple start calls, got none")
}