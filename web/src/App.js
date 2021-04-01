import React, {useEffect, useState} from 'react';
import {APISERVER_HOST} from './Constant'
import Header from './Header';
import VideoList from './VideoList';
import 'shaka-player/dist/controls.css';
import './App.css';
import Loader from "./Loader";
import Video from "./Video";

const App = () => {
    const [ui, setUI] = useState({
        header: 'DASH Video Player',
        headerVideoList: '',
    })

    const [data, setData] = useState([])
    const [activeVideo, setActiveVideo] = useState()

    useEffect(() => {
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
                <Header title={ui.header}/>
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
export default App;