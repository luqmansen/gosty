import React, {useEffect, useState} from 'react';
import {APISERVER_HOST, FILESERVER_HOST} from '../../Constant'
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
    const [activeVideo, setActiveVideo] = useState()

    useEffect(() => {
        console.log("APISERVER: ", APISERVER_HOST)
        console.log("FILESERVER: ", FILESERVER_HOST)

        fetch(APISERVER_HOST + "/playlist").then(response => {
            if (response.status !== 200) {
                console.log('Looks like there was a problem. Status Code: ', response.status);
                return;
            }
            response.json().then(data => {
                data.forEach((item, idx) => {
                    if (idx === 0) {
                        return item.isActive = true
                    }
                    item.isActive = false
                })
                setData(data);
                setActiveVideo(data.find(video => video.isActive))
            });
        }).catch(error => {
            console.log('There has been a problem with your fetch operation: ', error.message);
        });
    }, []);


    if (!activeVideo) {
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
