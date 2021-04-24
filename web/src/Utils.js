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