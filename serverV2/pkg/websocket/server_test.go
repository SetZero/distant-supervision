package websocket

import "testing"
import "github.com/onsi/gomega"

func TestMultipleStart(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s := NewServer()
	go s.Start()
	g.Expect(s.Start).To(gomega.Panic(), "Expected Panic on multiple start calls, got none")
}