export const bitRateChange = (newBitrate) => ({
    type: "settings/bitrate",
    payload: {bitrate: newBitrate}
});

export const darkMode = (darkMode) => ({
    type: "settings/darkMode",
    payload: {darkMode: darkMode}
});

export const showSettings = (show) => ({
    type: "settings/show",
    payload: {showSettings: show}
});

export const startStream = (started) => ({
    type: "stream/started",
    payload: {streamStarted: started}
});

export const streamRole = (role) => ({
    type: "stream/role",
    payload: {streamRole: role}
});

export const setResolution = (resolution) => ({
    type: "stream/resolution",
    payload: {streamResolution: resolution}
});

export const finishedLoading = (finished) => ({
    type: "stream/finishedLoading",
    payload: {finishedLoading: finished}
});

