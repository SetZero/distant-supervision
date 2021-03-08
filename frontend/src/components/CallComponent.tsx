import React, { useEffect, useRef, useState } from "react";
import { StreamRole, WebRTC } from "../classes/WebRTC";
import { Button, Container, Grid, Box, makeStyles, Typography, Chip, TextField, InputAdornment } from '@material-ui/core';
import { StreamBar } from './StreamBar'
import PeopleIcon from '@material-ui/icons/People';
import ThumbUpAltIcon from '@material-ui/icons/ThumbUpAlt';
import ThumbDownIcon from '@material-ui/icons/ThumbDown';
import { SettingsBar } from "./SettingsBar";
import { useDispatch, useSelector } from "react-redux";
import { showSettings, bitRateChange, startStream, streamRole, finishedLoading } from "../store/actions/rootActions";

const useStyles = makeStyles((theme) => ({
    video: {
        flexGrow: 1,
        maxWidth: '100%'
    },
    videoContainer: {
        backgroundColor: '#000000'
    },
    chatContainer: {
        backgroundColor: '#DDDDDD',
        flexGrow: 1,
        flexShrink: 1,
        flexBasis: 'auto',
        display: 'flex'
    },
    infoBar: {
        display: 'flex',
        justifyContent: 'center',
        flexWrap: 'wrap',
        '& > *': {
            margin: theme.spacing(0.5),
        },
    },
    chatContent: {
        flexGrow: 1
    }
}));

interface CallProps { }
let webRTC: WebRTC;

export const CallComponent: React.FC<CallProps> = () => {
    const classes = useStyles();
    const [hasActiveCall, setActiveCall] = useState(false);
    const [activeViewers, setActiveViewers] = useState(0);
    let video = React.createRef<HTMLVideoElement>();
    const bitrate = useSelector((state: any) => state.bitrate);
    const finished = useSelector((state: any) => state.finishedLoading);
    const resolution = useSelector((state: any) => state.streamResolution);
    const dispatch = useDispatch();
    const loadingFinished = (state: boolean) => { dispatch(finishedLoading(state)) };
    const streamRoleState: StreamRole = useSelector((state: any) => state.streamRole);

    useEffect(() => {
        if (!webRTC) {
            webRTC = new WebRTC(video, loadingFinished);
        } else {
            webRTC.setOutputVideo(video);
        }
    });

    if (!!webRTC) {
        webRTC.setBitrate(bitrate);
        webRTC.setVideoResolution(resolution.x, resolution.y)
    }

    function streamStartHandler() {
        if (webRTC) {
            console.log("ok")
            dispatch(streamRole(webRTC.role));
            dispatch(startStream(true));
        }
    }

    let streamInfo = hasActiveCall ?
        (
            <Grid item xs={12} >
                <Container className={classes.infoBar}>
                    <Chip avatar={<PeopleIcon />} label={activeViewers} />
                    <Chip avatar={<ThumbUpAltIcon />} label="200" />
                    <Chip avatar={<ThumbDownIcon />} label="0" />
                </Container>
            </Grid>
        )
        : (<div></div>);

    return (
        <div>
            {finished ?
                <div>
                    <SettingsBar />
                    <Box>
                        <StreamBar webrtc={webRTC} onStreamStart={() => streamStartHandler()} />
                    </Box>
                    <Box>
                        <Grid container>
                            <Grid item xs={12} sm={10} className={classes.videoContainer}>
                                <video id="localVideo" autoPlay playsInline controls={true} ref={video} className={classes.video} muted={streamRoleState === StreamRole.STREAMER}></video>
                            </Grid>
                            <Grid item xs={12} sm={2} className={classes.chatContainer}>
                                <Grid container>
                                    <Grid item xs={12} className={classes.chatContent}>

                                    </Grid>
                                    <Grid item xs={12}>
                                    </Grid>
                                </Grid>
                            </Grid>
                            {streamInfo}
                        </Grid>
                    </Box>
                </div> : <div>Loading...</div>
            }
        </div >
    )
}