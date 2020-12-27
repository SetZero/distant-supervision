enum ShareState {
    INITAL,
    IN_PROGRESS,
    STABLE
}


const configuration = {
    configuration: {
        offerToReceiveVideo: true
    },
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
};

const constraints = {
    'video': true,
    'audio': false
}

export class WebRTC {
    private readonly port = (window.location.hostname === "localhost" ? ":5501" : "");
    private readonly protocol = (this.isSecureContext() ? "wss://" : "ws://");
    private conn = new WebSocket(this.protocol + document.location.hostname + this.port + "/ws");;
    private peerConnection = new RTCPeerConnection(configuration);
    private stream: any;
    private video;
    private videoState = ShareState.INITAL;
    private readonly myId = this.generatRandomId(36);

    constructor(video: React.RefObject<HTMLVideoElement>, finishedLoading: React.Dispatch<React.SetStateAction<boolean>>) {
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
            this.conn.onopen = () => resolve(this.conn);
        });
    }

    public async startCall() {
        if (this.videoState !== ShareState.INITAL)
            return;

        this.videoState = ShareState.IN_PROGRESS;
        console.log("start call");
        // @ts-ignore
        this.stream = await navigator.mediaDevices.getDisplayMedia(constraints);
        this.peerConnection = new RTCPeerConnection(configuration);
        this.conn.send(JSON.stringify({ 'id': this.myId, 'incomingCall': true }));
        this.peerConnection.addEventListener('icecandidate', e => this.onIceCandidate(this.peerConnection, e));
        this.peerConnection.addEventListener('iceconnectionstatechange', e => this.onIceStateChange(this.peerConnection, e));
        this.stream.getTracks().forEach((track: MediaStreamTrack) => this.peerConnection.addTrack(track, this.stream));

        const offer = await this.peerConnection.createOffer({
            offerToReceiveAudio: false,
            offerToReceiveVideo: true
        });
        await this.onCreateOfferSuccess(offer);
    }

    private async onCreateOfferSuccess(desc: RTCSessionDescriptionInit) {
        await this.peerConnection.setLocalDescription(desc);
        let data = await JSON.stringify({ 'id': this.myId, 'sendOffer': desc });
        this.conn.send(data);
    }

    private async acceptCall(message: any) {
        console.log("message: ", message);
        if (message.id === this.myId) return;

        if (message.incomingCall && this.videoState === ShareState.INITAL) {
            this.videoState = ShareState.IN_PROGRESS;
            console.log("incoming call!");
            this.peerConnection = new RTCPeerConnection(configuration);
            this.peerConnection.addEventListener('icecandidate', e => this.onIceCandidate(this.peerConnection, e));
            this.peerConnection.addEventListener('iceconnectionstatechange', e => this.onIceStateChange(this.peerConnection, e));
            this.peerConnection.addEventListener('track', e => this.gotRemoteStream(e));
        } else if (message.sendOffer && this.videoState === ShareState.IN_PROGRESS) {
            console.log("Got Offer!");
            await this.peerConnection.setRemoteDescription(message.sendOffer);
            const answer = await this.peerConnection.createAnswer();
            await this.onCreateAnswerSuccess(answer);
        } else if (message.sendAnswer && this.videoState === ShareState.IN_PROGRESS) {
            this.peerConnection.setRemoteDescription(message.sendAnswer);
            console.log("Got Answer!");
        } else if (message.sendIceCanidate && this.videoState === ShareState.IN_PROGRESS) {
            this.peerConnection.addIceCandidate(message.sendIceCanidate);
            console.log("Got ICE Canidate!");
        }
    }

    private async onCreateAnswerSuccess(desc: RTCSessionDescriptionInit) {
        await this.peerConnection.setLocalDescription(desc);
        this.conn.send(JSON.stringify({ 'id': this.myId, 'sendAnswer': desc }));
    }

    private gotRemoteStream(e: RTCTrackEvent) {
        this.videoState = ShareState.STABLE;
        if (this.video.current && this.video.current.srcObject !== e.streams[0]) {
            this.video.current.srcObject = e.streams[0];
            console.log('received remote stream');
        }
    }

    private async onIceCandidate(pc: RTCPeerConnection, event: RTCPeerConnectionIceEvent) {
        console.log("called ice!", pc);
        this.conn.send(JSON.stringify({ 'id': this.myId, 'sendIceCanidate': event.candidate }));
    }

    private onIceStateChange(pc: RTCPeerConnection, event: Event) {
        if (pc) {
            console.log(`${(pc)} ICE state: ${pc.iceConnectionState}`);
            console.log('ICE state change event: ', event);
            if (((event.target) as RTCPeerConnection)?.iceConnectionState === "disconnected") {
                this.videoState = ShareState.INITAL;
                console.log(((event.target) as RTCPeerConnection)?.iceConnectionState);
            }
        }
    }
}