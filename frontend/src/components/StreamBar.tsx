import React, { useEffect, useRef, useState } from "react";
import { WebRTC } from "../classes/WebRTC";
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import { makeStyles } from "@material-ui/core";

const useStyles = makeStyles((theme) => ({
    root: {
        flexGrow: 1,
    },
    menuButton: {
        marginRight: theme.spacing(2),
    },
    title: {
        flexGrow: 1,
    },
}));

interface StreamBarProps {
    webrtc: WebRTC
    onStreamStart: () => void
}

export const StreamBar: React.FC<StreamBarProps> = ({ webrtc, onStreamStart }) => {
    const classes = useStyles();

    if (!webrtc) return (<div />);
    return (<AppBar position="static">
        <Toolbar>
            <IconButton edge="start" color="inherit" aria-label="menu">
                <MenuIcon />
            </IconButton>
            <Typography variant="h6" className={classes.title}>
                ScreenShare
            </Typography>
            <Button
                id="startCall"
                onClick={() => { webrtc.startCall(); onStreamStart() }}
                disabled={webrtc.isStreamActive()}
                variant="contained"
                color="default"
            >
                {webrtc.isStreamActive() ? "stream in progress..." : "start streaming!"}
            </Button>
        </Toolbar>
    </AppBar>);
}