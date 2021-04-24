export const getMpd = (dashList) => {
    let d;
    d = "";
    dashList.forEach((item) => {
        let name = item.split(".")
        if (name[name.length - 1] === 'mpd') {
            d = item
        }
    })
    return d
}

export const msToTime = (ms) => {
    let seconds = (ms / 1000).toFixed(1);
    let minutes = (ms / (1000 * 60)).toFixed(1);
    let hours = (ms / (1000 * 60 * 60)).toFixed(1);
    let days = (ms / (1000 * 60 * 60 * 24)).toFixed(1);
    if (seconds < 60) return seconds + " Sec";
    else if (minutes < 60) return minutes + " Min";
    else if (hours < 24) return hours + " Hrs";
    else return days + " Days"
}