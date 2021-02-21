import { MenuItem, Select } from "@material-ui/core";
import { Box, Drawer, InputAdornment, TextField } from "@material-ui/core";
import React from "react";
import { useDispatch, useSelector } from "react-redux";
import { showSettings, bitRateChange, setResolution } from "../store/actions/rootActions";

interface SettingsBarProps {
}

export const SettingsBar: React.FC<SettingsBarProps> = ({ }) => {
    const showBar = useSelector((state: any) => state.showSettings);
    const bitrate = useSelector((state: any) => state.bitrate);
    const resolution = useSelector((state: any) => state.streamResolution);
    const dispatch = useDispatch();
    let resolutions = [
        { name: "240p", value: { x: 426, y: 240 } },
        { name: "360p", value: { x: 640, y: 360 } },
        { name: "480p", value: { x: 854, y: 480 } },
        { name: "720p", value: { x: 1280, y: 720 } },
        { name: "1080p", value: { x: 1920, y: 1080 } },
        { name: "1440p", value: { x: 2560, y: 1440 } },
        { name: "2160p", value: { x: 3840, y: 2160 } }
    ]

    function handleBitrateChange(event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) {
        let val = parseInt(event.currentTarget.value);
        dispatch(bitRateChange(val));
    }

    function handleResolutionChange(event: React.ChangeEvent<{ name?: string | undefined; value: unknown; }>) {
        let val = event.target.value;
        dispatch(setResolution(JSON.parse(val as string)));
    }

    console.log(resolution);
    return (<React.Fragment >
        <Drawer anchor="left" open={showBar} onClose={() => { dispatch(showSettings(false)) }}>
            <Box m={2}>
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
            <Box m={2}>
                <Select
                    labelId="demo-simple-select-label"
                    id="demo-simple-select"
                    value={JSON.stringify(resolution)}
                    onChange={(e) => handleResolutionChange(e)}
                >
                    {resolutions.map((resolution, index) => (
                        <MenuItem value={JSON.stringify(resolution.value)} key={index}>{resolution.name}</MenuItem>
                    ))}
                </Select>
            </Box>
        </Drawer>
    </React.Fragment>);
}