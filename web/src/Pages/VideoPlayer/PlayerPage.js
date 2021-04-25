import React, {useEffect, useState} from 'react';
import {APISERVER_HOST, FILESERVER_HOST, VIDEO_PLAYLIST_ENDPOINT} from '../../Constant'
import HeaderVideo from '../../Components/HeaderVideo';
import VideoList from './VideoList';
import 'shaka-player/dist/controls.css';
import '../../style/App.css';
import Loader from "../../Components/Loader";
import Video from "./Video";

const PlayerPage = () => {
    const [ui, setUI] = useState({
        header: 'DASH Video Player',
        headerVideoList: '',
    })

    const [data, setData] = useState([])
    const [isEmpty, setIsEmpty] = useState(true)
    const [activeVideo, setActiveVideo] = useState()

    useEffect(() => {
        console.log("APISERVER: ", APISERVER_HOST)
        console.log("FILESERVER: ", FILESERVER_HOST)

        fetch(APISERVER_HOST + VIDEO_PLAYLIST_ENDPOINT).then(response => {
            if (response.status === 204) {
                setIsEmpty(true)
                return;
            } else if (response.status === 200) {
                response.json().then(data => {
                    data.forEach((item, idx) => {
                        if (idx === 0) {
                            return item.isActive = true
                        }
                        item.isActive = false
                    })
                    setData(data);
                    setActiveVideo(data.find(video => video.isActive))
                    setIsEmpty(false)
                });
            } else {
                console.log('Looks like there was a problem. Status Code: ', response.status);
                return;
            }
        }).catch(error => {
            console.log('There has been a problem with your fetch operation: ', error.message);
        });
    }, []);

    if (isEmpty) {
        return (
            <>
                <div className="container">
                    <h1>DASH Video Player</h1>
                    <p>Uploaded video will be shown here</p>
                </div>
            </>
        )
    }

    if (!activeVideo) {
        console.log("no active video")
        return <Loader/>
    }
    return (
        <div className='app'>
            <div className='app__video'>
                <HeaderVideo title={ui.header}/>
                <Video video={activeVideo}/>
            </div>
            <div className='app__videoList'>
                <VideoList
                    data={data}
                    title={ui.headerVideoList}
                    setActiveVideo={setActiveVideo}
                />
            </div>
        </div>
    );
}
export default PlayerPage;
