import React from 'react';
import {FILESERVER_HOST}from './Constant'
import {getMpd} from "./Utils";
import ShakaPlayer from "shaka-player-react";


const Video = (props) => {
    const src = FILESERVER_HOST + "/files/" + getMpd(props.video.dash_file)

    return <ShakaPlayer
        autoPlay={false}
        src={src}
    />
}


export default Video;
