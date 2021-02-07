import { Drawer } from "@material-ui/core";
import React from "react";

interface SettingsBarProps {
    showBar: boolean
}

export const SettingsBar: React.FC<SettingsBarProps> = ({ showBar }) => {
    return (<React.Fragment >
        <Drawer anchor="left" open={showBar} onClose={() => {}}>
            {}
        </Drawer>
    </React.Fragment>);
}