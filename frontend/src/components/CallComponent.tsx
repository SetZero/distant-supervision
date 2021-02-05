import React, { useEffect, useRef, useState } from "react";
import { StreamRole, WebRTC } from "../classes/WebRTC";
import { Button, Container, Grid, Box, makeStyles, Typography, Chip, TextField, InputAdornment } from '@material-ui/core';
import { StreamBar } from './StreamBar'
import PeopleIcon from '@material-ui/icons/People';
import ThumbUpAltIcon from '@material-ui/icons/ThumbUpAlt';
import ThumbDownIcon from '@material-ui/icons/ThumbDown';

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
let webRTC: WebRTC

export const CallComponent: React.FC<CallProps> = () => {
    const classes = useStyles();
    const [finishedLoading, setFinishedLoading] = useState(false);
    const [hasActiveCall, setActiveCall] = useState(false);
    const [activeViewers, setActiveViewers] = useState(0);
    const [bitrate, setBitrate] = useState(0);
    const [bitrateButton, setBitrateButton] = useState((<div></div>));
    let video = React.createRef<HTMLVideoElement>();


    useEffect(() => {
        if (!webRTC) {
            webRTC = new WebRTC(video, setFinishedLoading, setActiveCall, hasActiveCall, setActiveViewers);
        } else {
            webRTC.setOutputVideo(video);
        }
    });

    function handleBitrateChange(event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) {
        console.log("called!")
        let val = parseInt(event.currentTarget.value);
        if (!!webRTC) {
            webRTC.setBitrate(val);
        }
    }

    function streamStartHandler() {
        if (webRTC && webRTC.role === StreamRole.STREAMER) {
            setBitrateButton(
                <Box>
                    <TextField
                        label="Bandwith"
                        id="standard-start-adornment"
                        type="number"
                        onChange={(e) => handleBitrateChange(e)}
                        InputProps={{
                            endAdornment: <InputAdornment position="end">Kb/s</InputAdornment>,
                        }}
                    />
                </Box>
            );
        }
    }

    //<video id="localVideo" autoPlay playsInline controls={true} ref={video} className={classes.video}></video>
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
            {finishedLoading ?
                <div>
                    <Box>
                        <StreamBar webrtc={webRTC} onStreamStart={() => streamStartHandler()} />
                    </Box>
                    <Box>
                        <Grid container>
                            <Grid item xs={12} sm={10} className={classes.videoContainer}>
                                <video id="localVideo" autoPlay playsInline controls={true} ref={video} className={classes.video}></video>
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
                    {bitrateButton}
                </div> : <div>Loading...</div>
            }
        </div >
    )
}