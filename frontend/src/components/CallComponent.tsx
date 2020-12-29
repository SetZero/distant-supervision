import React, { useEffect, useRef, useState } from "react";
import { WebRTC } from "../classes/WebRTC";
import { Button } from '@material-ui/core';

interface CallProps { }
let webRTC: WebRTC

export const CallComponent: React.FC<CallProps> = () => {
    const [callEnabled, setCallEnabled] = useState(true);
    const [finishedLoading, setFinishedLoading] = useState(false);
    const [hasActiveCall, setActiveCall] = useState(false);
    const video = React.createRef<HTMLVideoElement>();


    useEffect(() => {
        if(!webRTC) {
            webRTC = new WebRTC(video, setFinishedLoading, setActiveCall);
        } else {
            webRTC.setOutputVideo(video);
        }
    });


    return (
        <div>
            {finishedLoading ?
                <div>
                    <video id="localVideo" autoPlay playsInline controls={true} ref={video}></video>
                    <Button id="startCall" onClick={() => webRTC.startCall()} disabled={!callEnabled || hasActiveCall}>{hasActiveCall ? "stream in progress..." : "start streaming!"}</Button>
                </div> : <div>Loading...</div>
            }
        </div>
    )
}