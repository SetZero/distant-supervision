import { off } from "process";
import React, { useState } from "react";

enum ShareState {
    INITAL,
    IN_PROGRESS,
    STABLE
}

export enum StreamRole {
    STREAMER,
    VIEWER
}

const configuration = {
    configuration: {
    },
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
};

export class WebRTC {
    /*private readonly hostnameOverwrite = "rtc.magical.rocks";
    private readonly protocolOverwrite = "wss://";
    private readonly portOverwrite = false;*/
    private readonly hostnameOverwrite = undefined;
    private readonly protocolOverwrite = undefined;
    private readonly portOverwrite = undefined;

    private readonly port = ((window.location.hostname === "localhost" && this.portOverwrite) ? ":5501" : "");
    private readonly protocol = (this.isSecureContext() ? "wss://" : "ws://");
    private conn = new WebSocket((this.protocolOverwrite || this.protocol) + (this.hostnameOverwrite || document.location.hostname) + this.port + "/ws");;
    private peerConnection = new RTCPeerConnection(configuration);
    private stream: any;
    private video;
    private videoState = ShareState.INITAL;
    private readonly myId = this.generatRandomId(36);
    private myRoom = window.location.hash || "DEFAULT";
    private webRtcStarted = false;
    private streamRole = StreamRole.VIEWER;
    private constraints = {
        video: {
            width: { ideal: 1280 },
            height: { ideal: 1024 },
            cursor: "never",
            facingMode: "environment"
        },
        'audio': true
    };
    public callOnConnected: () => any = () => { };

    constructor(video: React.RefObject<HTMLVideoElement>,
        finishedLoading: (state: boolean) => void) {

        console.log("created object")
        this.video = video;
        this.startSignaling(window.location.hash).then((c) => {
            c.addEventListener('message', async message => {
                try {
                    let jsonArr = message.data.trim().split("\n").filter((el: string) => el !== "");
                    jsonArr.forEach((element: string) => {
                        this.acceptCall(JSON.parse(element));
                    });
                } catch (e) {
                    console.error(e);
                    console.log(message.data);
                }
            });
            finishedLoading(true);
        });
    }

    public setOutputVideo(video: React.RefObject<HTMLVideoElement>) {
        this.video = video;
    }

    private isSecureContext() {
        return window.location.protocol === 'https:';
    }

    public setVideoResolution(x: number, y: number) {
        this.constraints.video.width.ideal = x;
        this.constraints.video.height.ideal = y;
    }
    private generatRandomId(length: Number) {
        var result = '';
        var characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        var charactersLength = characters.length;
        for (var i = 0; i < length; i++) {
            result += characters.charAt(Math.floor(Math.random() * charactersLength));
        }
        return result;
    }

    private startSignaling(userId: String) {
        return new Promise<WebSocket>((resolve, reject) => {
            this.conn.onopen = () => {
                this.conn.send(JSON.stringify({ type: "joinMessage", message: { roomId: this.myRoom } }))
                resolve(this.conn);
            }
        });
    }

    public async startCall() {
        if (this.videoState !== ShareState.INITAL)
            return;

        this.streamRole = StreamRole.STREAMER;
        this.videoState = ShareState.IN_PROGRESS;
        console.log("start call");
        // @ts-ignore
        this.stream = await navigator.mediaDevices.getDisplayMedia(this.constraints);
        this.showOwnVideo();
        this.conn.send(JSON.stringify({ type: "startStream" }))
    }

    public get role() {
        return this.streamRole;
    }

    public async startWebRTC() {
        this.peerConnection = new RTCPeerConnection(configuration);
        this.peerConnection.addEventListener('icecandidate', e => this.onIceCandidate(this.peerConnection, e));
        this.peerConnection.addEventListener('iceconnectionstatechange', e => this.onIceStateChange(this.peerConnection, e));
        this.peerConnection.addEventListener('connectionstatechange', e => this.connectionStateChange(e));
        this.peerConnection.addEventListener('track', e => this.gotRemoteStream(e));
        if (this.stream)
            this.stream.getTracks().forEach((track: MediaStreamTrack) => this.peerConnection.addTrack(track, this.stream));

        let offer = await this.peerConnection.createOffer({
            offerToReceiveAudio: true,
            offerToReceiveVideo: true
        });
        if (offer.sdp) {
            await this.onCreateOfferSuccess(offer);
        }
    }

    public async setBitrate(bitrateMbps: number) {
        this.conn.send(JSON.stringify({ type: "bitrateChange", message: { bitrate: bitrateMbps } }));
    }

    private async onCreateOfferSuccess(desc: RTCSessionDescriptionInit) {
        await this.peerConnection.setLocalDescription(desc);
        let data = await JSON.stringify({ 'type': "webRtcOffer", message: { 'offer': btoa(JSON.stringify(desc)) } });
        this.conn.send(data);
    }

    private async acceptCall(message: any) {
        switch (message.type) {
            case "error":
                console.log("There was an error while performing websocket communication: ", message.message.Type);
                break;
            case "joinedMessage":
                //this.setActiveCall(message.message.roomHasStreamer)
                if (message.message.roomHasStreamer && !this.webRtcStarted) {
                    this.webRtcStarted = true;
                    this.startWebRTC();
                }
                break;
            case "startStream":
                if (message.message.startStreamSuccess && !this.webRtcStarted) {
                    this.webRtcStarted = true;
                    this.startWebRTC();
                } else {
                    console.log("someone else is streaming...")
                }
                break;
            case "answer":
                let t = message.message;

                /* screw webrtc specs */
                var arr = t.sdp.split('\r\n');
                arr.forEach((str: string, i: number) => {
                    if (/^a=fmtp:\d*/.test(str)) {
                        arr[i] = str + ';x-google-max-bitrate=1000000;x-google-min-bitrate=600000;x-google-start-bitrate=1200000';
                    } else if (/^a=mid:(1|video)/.test(str)) {
                        arr[i] += '\r\nb=AS:100000';
                    }
                });
                t = new RTCSessionDescription({
                    type: 'answer',
                    sdp: arr.join('\r\n'),
                });
                this.peerConnection.setRemoteDescription(t);
                break;
            case "newIceCandidate":
                if (message.message === null) break;
                try {
                    console.log("New Ice Candidate: ", message.message)
                    this.peerConnection.addIceCandidate(message.message);
                } catch (e) {
                    console.error("Failed to add ICE candidate: ", e)
                }
                break;
            case "currentViewers":
                if (message.message === null) break;
                //this.setActiveViewers(message.message.viewers);
                break;
        }
    }

    private gotRemoteStream(e: RTCTrackEvent) {
        console.log("Got Stream!");
        this.videoState = ShareState.STABLE;
        if (this.streamRole === StreamRole.VIEWER && this.video.current && this.video.current.srcObject !== e.streams[0]) {
            this.video.current.srcObject = e.streams[0];
            console.log('received remote stream', e.streams);
        }
    }

    private showOwnVideo() {
        if (this.video.current) {
            console.log(this.stream.getVideoTracks());
            this.video.current.srcObject = this.stream;
        }
    }

    private async onIceCandidate(pc: RTCPeerConnection, event: RTCPeerConnectionIceEvent) {
        console.log("called ice!", event.candidate);
        this.conn.send(JSON.stringify({ 'type': 'iceCandidate', 'message': event.candidate }));
    }

    private onIceStateChange(pc: RTCPeerConnection, event: Event) {
        if (pc) {
            console.log(this.peerConnection)
            console.log(`${(pc)} ICE state: ${pc.iceConnectionState}`);
            console.log('ICE state change event: ', event);
            if (((event.target) as RTCPeerConnection)?.iceConnectionState === "disconnected") {
                this.videoState = ShareState.INITAL;
                console.log(((event.target) as RTCPeerConnection)?.iceConnectionState);
            }
        }
    }

    private connectionStateChange(event: Event) {
        console.log("event state change: ", event);
        if ((event.currentTarget as any).connectionState === "connected") {
            this.callOnConnected();
        }
    }
}