import React, { useEffect, useRef, useState } from "react";
import { WebRTC } from "../classes/WebRTC";
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import SettingsIcon from '@material-ui/icons/Settings';
import { makeStyles } from "@material-ui/core";
import { useDispatch, useSelector } from 'react-redux';
import { showSettings } from "../store/actions/rootActions"

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
    const streamStarted = useSelector((state: any) => state.streamStarted);
    const dispatch = useDispatch();

    if (!webrtc) return (<div />);
    return (<AppBar position="static">
        <Toolbar>
            <IconButton edge="start" color="inherit" aria-label="menu" onClick={() => dispatch(showSettings(true))}>
                <SettingsIcon />
            </IconButton>
            <Typography variant="h6" className={classes.title}>
                ScreenShare
            </Typography>
            <Button
                id="startCall"
                onClick={() => { webrtc.startCall(); onStreamStart() }}
                disabled={streamStarted}
                variant="contained"
                color="default"
            >
                {streamStarted ? "stream in progress..." : "start streaming!"}
            </Button>
        </Toolbar>
    </AppBar>);
}