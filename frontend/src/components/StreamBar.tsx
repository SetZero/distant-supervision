import React, { useEffect, useRef, useState } from "react";
import { WebRTC } from "../classes/WebRTC";
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';

interface StreamBarProps { }

export const StreamBar: React.FC<StreamBarProps> = () => {
    return (<AppBar position="static">
        <Toolbar>
            <IconButton edge="start" color="inherit" aria-label="menu">
                <MenuIcon />
            </IconButton>
            <Typography variant="h6">
                News
  </Typography>
            <Button color="inherit">Login</Button>
        </Toolbar>
    </AppBar>);
}