import React, { useEffect, useRef, useState } from "react";
import { WebRTC } from "../classes/WebRTC";

interface CallProps { }
let webRTC: WebRTC

export const CallComponent: React.FC<CallProps> = () => {
    const [callEnabled, setCallEnabled] = useState(false);
    const [finishedLoading, setFinishedLoading] = useState(false);
    const video = React.createRef<HTMLVideoElement>();


    useEffect(() => {
        if(!webRTC) {
            webRTC = new WebRTC(video, setFinishedLoading);
        } else {
            webRTC.setOutputVideo(video);
        }
    });


    return (
        <div>
            {finishedLoading ?
                <div>
                    <video id="localVideo" autoPlay playsInline controls={true} ref={video}></video>
                    <button id="startCall" onClick={() => webRTC.startCall()} disabled={callEnabled}>Call</button>
                </div> : <div>Loading...</div>
            }
        </div>
    )
}