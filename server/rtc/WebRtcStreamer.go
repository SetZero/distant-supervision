package rtc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc"
	"time"
)

type WebRTCStreamer struct {
	peerConnection *webrtc.PeerConnection
	Send           chan []byte
	Recv           chan []byte
	WebRtcStream   chan *rtp.Packet
}

func NewWebRTCStreamer() *WebRTCStreamer {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	m := webrtc.MediaEngine{}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&m))
	peerConnection, err := api.NewPeerConnection(config)
	rtc := &WebRTCStreamer{Send: make(chan []byte, 16384), Recv: make(chan []byte, 16384), peerConnection: peerConnection}

	outputTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
	rtpSender, err := peerConnection.AddTrack(outputTrack)
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				fmt.Println("Error in rtp!")
				return
			}
		}
	}()

	if err != nil {
		return nil
	} else {
		return rtc
	}
}

func Decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
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
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected {

			fmt.Println("Done writing media files")
		}
	})

	for {
		webRTCMessage := <-r.Recv

		var m messageWrapper
		err := json.Unmarshal(webRTCMessage, &m)
		if err == nil {
			switch m.Type {
			case webRTCOffer:
				var offerMessage webRtcOffer
				json.Unmarshal(m.Message, &offerMessage)
				offer := webrtc.SessionDescription{}
				Decode(offerMessage.Offer, &offer)
				err = r.peerConnection.SetRemoteDescription(offer)
				if err != nil {
					fmt.Println("Error!")
				}
				answer, err := r.peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}
				jsonAnswer, err := json.Marshal(answer)
				if err != nil {
					panic(err)
				}
				r.Send <- jsonAnswer

				//gatherComplete := webrtc.GatheringCompletePromise(r.peerConnection)
				if err = r.peerConnection.SetLocalDescription(answer); err != nil {
					panic(err)
				}
				// <-gatherComplete
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
