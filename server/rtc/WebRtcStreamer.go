package rtc

import (
	"encoding/json"
	"fmt"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc"
	"time"
)

type WebRTCStreamer struct {
	peerConnection *webrtc.PeerConnection
	send           chan []byte
	recv           chan []byte
	WebRtcStream   chan *rtp.Packet
}

func NewWebRTCStreamer() *WebRTCStreamer {
	connectionInfo, err := createPeerConnection(true)
	if err == nil && connectionInfo != nil {
		rtc := &WebRTCStreamer{send: make(chan []byte, 16384), recv: make(chan []byte, 16384), peerConnection: connectionInfo.peerConnection, WebRtcStream: make(chan *rtp.Packet)}
		return rtc
	} else {
		return nil
	}
}

func (r *WebRTCStreamer) Send() chan []byte {
	return r.send
}

func (r *WebRTCStreamer) Recv() chan []byte {
	return r.recv
}

func (r *WebRTCStreamer) Start() {
	r.peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := r.peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), track.Codec().MimeType)
		codec := track.Codec()
		if codec.MimeType == "video/VP8" {
			for {
				rtpPkg, _, readErr := track.ReadRTP()
				if readErr != nil {
					panic(readErr)
				}
				r.WebRtcStream <- rtpPkg
			}
		}
	})

	r.peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("[Streamer] Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected {

			fmt.Println("Done writing media files")
		}
	})

	for {
		webRTCMessage := <-r.recv

		var m messageWrapper
		err := json.Unmarshal(webRTCMessage, &m)
		if err == nil {
			switch m.Type {
			case webRTCOffer:
				fmt.Println("check")
				var offerMessage webRtcOffer
				json.Unmarshal(m.Message, &offerMessage)
				offer := webrtc.SessionDescription{}
				Decode(offerMessage.Offer, &offer)
				err = r.peerConnection.SetRemoteDescription(offer)
				if err != nil {
					fmt.Println("[Streamer] Error: ", err)
				}
				answer, err := r.peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}
				jsonAnswer, err := json.Marshal(answer)
				if err != nil {
					panic(err)
				}
				r.send <- jsonAnswer

				if err = r.peerConnection.SetLocalDescription(answer); err != nil {
					panic(err)
				}
				break
			case iceCandidate:
				var iceandidate webrtc.ICECandidateInit
				json.Unmarshal(m.Message, &iceandidate)
				err := r.peerConnection.AddICECandidate(iceandidate)
				if err != nil {
					panic(err)
				}
				break
			}
		} else {
			fmt.Println("Error: ", err)
		}
	}
}
