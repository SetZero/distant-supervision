let initialState = {
    bitrate: localStorage.getItem('settings/bitrate') ?? 2000,
    showSettings: false,
    streamStarted: false,
    streamRole: -1,
    streamResolution: JSON.parse(localStorage.getItem('stream/resolution') ?? "null") ?? { x: 1280, y: 720 },
    finishedLoading: false,
    darkMode: localStorage.getItem("settings/darkMode") ?? false
};

export function rootReducer(state = initialState, action: any) {
    switch (action.type) {
        case 'settings/bitrate':
            localStorage.setItem('settings/bitrate', action.payload.bitrate);
            return { ...state, bitrate: action.payload.bitrate };
        case 'settings/show':
            return { ...state, showSettings: action.payload.showSettings };
        case 'stream/started':
            return { ...state, streamStarted: action.payload.streamStarted };
        case 'stream/role':
            return { ...state, streamRole: action.payload.streamRole };
        case 'stream/resolution':
            localStorage.setItem('stream/resolution', JSON.stringify(action.payload.streamResolution));
            return { ...state, streamResolution: action.payload.streamResolution };
        case 'stream/finishedLoading':
            return { ...state, finishedLoading: action.payload.finishedLoading };
        case 'settings/darkMode':
            localStorage.setItem("settings/darkMode", action.payload.darkMode)
            return { ...state, darkMode: action.payload.darkMode };
        default:
            return state;
    }
}