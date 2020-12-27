const constraints = {
    'video': true,
    'audio': false
}
var conn;
const myId = generatRandomId(36);
const configuration = {
    configuration: {
        offerToReceiveVideo: true
    },
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
};
var peerConnection;
var stream;

const callButton = document.querySelector("#startCall");
const video = document.querySelector('video');

function generatRandomId(length) {
    var result = '';
    var characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    var charactersLength = characters.length;
    for (var i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}

function startSignaling(userId) {
    let socketAddr = "ws://" + document.location.hostname + ":5501/ws";
    console.log(socketAddr);
    conn = new WebSocket(socketAddr);
    return new Promise(function (resolve, reject) {
        conn.onopen = () => resolve(conn);
    });
}

async function startCall() {
    console.log("start call");
    callButton.disabled = true;
    stream = await navigator.mediaDevices.getDisplayMedia(constraints);
    peerConnection = new RTCPeerConnection(configuration);
    conn.send(JSON.stringify({ 'id': myId, 'incomingCall': true }));
    peerConnection.addEventListener('icecandidate', e => onIceCandidate(peerConnection, e));
    peerConnection.addEventListener('iceconnectionstatechange', e => onIceStateChange(peerConnection, e));
    stream.getTracks().forEach(track => peerConnection.addTrack(track, stream));

    const offer = await peerConnection.createOffer({
        offerToReceiveAudio: 0,
        offerToReceiveVideo: 1
    });
    await onCreateOfferSuccess(offer);
}

async function onCreateOfferSuccess(desc) {
    await peerConnection.setLocalDescription(desc);
    let data = await JSON.stringify({ 'id': myId, 'sendOffer': desc });
    conn.send(data);
}

async function acceptCall(message) {
    console.log("message: ", message);
    if (message.id === myId) return;

    if (message.incomingCall) {
        console.log("incoming call!");
        callButton.disabled = true;
        peerConnection = new RTCPeerConnection(configuration);
        peerConnection.addEventListener('icecandidate', e => onIceCandidate(peerConnection, e));
        peerConnection.addEventListener('iceconnectionstatechange', e => onIceStateChange(peerConnection, e));
        peerConnection.addEventListener('track', gotRemoteStream);
    } else if (message.sendOffer) {
        console.log("Got Offer!");
        await peerConnection.setRemoteDescription(message.sendOffer);
        const answer = await peerConnection.createAnswer();
        await onCreateAnswerSuccess(answer);
    } else if (message.sendAnswer) {
        peerConnection.setRemoteDescription(message.sendAnswer);
        console.log("Got Answer!");
    } else if (message.sendIceCanidate) {
        peerConnection.addIceCandidate(message.sendIceCanidate);
        console.log("Got ICE Canidate!");
    }
}

async function onCreateAnswerSuccess(desc) {
    await peerConnection.setLocalDescription(desc);
    conn.send(JSON.stringify({ 'id': myId, 'sendAnswer': desc }));
}

function gotRemoteStream(e) {
    if (video.srcObject !== e.streams[0]) {
        video.srcObject = e.streams[0];
        console.log('received remote stream');
    }
}

async function onIceCandidate(pc, event) {
    console.log("called ice!", pc);
    conn.send(JSON.stringify({ 'id': myId, 'sendIceCanidate': event.candidate }));
}

function onIceStateChange(pc, event) {
    if (pc) {
      console.log(`${(pc)} ICE state: ${pc.iceConnectionState}`);
      console.log('ICE state change event: ', event);
    }
  }

(function () {
    startSignaling(window.location.hash).then((c) => {
        callButton.addEventListener("click", e => startCall());
        c.addEventListener('message', async message => acceptCall(JSON.parse(message.data)));
    });
})();