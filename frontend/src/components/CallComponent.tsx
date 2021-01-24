import React, { useEffect, useRef, useState } from "react";
import { WebRTC } from "../classes/WebRTC";
import { Button, Container, Grid, Box, makeStyles, Typography, Chip, TextField } from '@material-ui/core';
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
    const video = React.createRef<HTMLVideoElement>();


    useEffect(() => {
        if (!webRTC) {
            webRTC = new WebRTC(video, setFinishedLoading, setActiveCall, hasActiveCall);
        } else {
            webRTC.setOutputVideo(video);
        }
    });


    return (
        <div>
            {finishedLoading ?
                <div>
                    <Box>
                        <StreamBar webrtc={webRTC} />
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
                            <Grid item xs={12}>
                                <Container className={classes.infoBar}>
                                    <Chip avatar={<PeopleIcon />} label="1.212.413" />
                                    <Chip avatar={<ThumbUpAltIcon />} label="200" />
                                    <Chip avatar={<ThumbDownIcon />} label="0" />
                                </Container>
                            </Grid>
                        </Grid>
                    </Box>
                </div> : <div>Loading...</div>
            }
        </div>
    )
}