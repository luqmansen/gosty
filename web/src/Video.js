import React, {useEffect, useRef} from 'react';
import {FILESERVER_HOST}from './Constant'
import indigoPlayer from "indigo-player";
import "indigo-player/lib/indigo-theme.css";
import {getMpd} from "./Utils";


const Video = (props) => {
    const ref = useRef(null)

    useEffect(() => {
            indigoPlayer.init(ref.current, {
                sources: [
                    {
                        type: "dash",
                        src: FILESERVER_HOST + "/files/" + getMpd(props.video.dash_file),
                    },
                ],
            });
    });

    return <div ref={ref}/>
}


export default Video;
