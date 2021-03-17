import React, {useEffect, useState} from 'react';
import {API_SERVER_HOST}from './Constant'
import Header from './Header';
import VideoList from './VideoList';
import Video from './Video';
import './App.css';
import Loader from "./Loader";

const App = () => {
  const [ui, setUI] = useState({
      header: 'DASH Video Player',
      headerVideoList: 'Top 10s',
    })

  const [data, setData] = useState([])
  const [activeVideo, setActiveVideo] = useState()
  // const [player, setPlayer] = useState(<div/>)
  // useEffect(() => {
  //   let currentActiveVideo = data.find(video => video.isActive);
  //   currentActiveVideo.isActive = false;
  //
  //   let index = data.findIndex(video => video.id === id);
  //   data[index].isActive = true;
  //
  //   setActiveVideo(data.find(video => video.isActive))
  // })


  useEffect( () =>{
    fetch(API_SERVER_HOST + "/playlist" ).then(response => {
      if (response.status !== 200) {
        console.log('Looks like there was a problem. Status Code: ', response.status);
        return;
      }
      response.json().then(data => {
        data.forEach((item, idx) => {
          if (idx === 0){
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

    if (!activeVideo){
      console.log("asu not loaded")
      return <Loader/>
    }

    return (
      <div className='app'>
        <div className='app__video'>
          <Header title={ui.header}/>
          <div>
            <Video video={activeVideo} />
          </div>

        </div>
        <div className='app__videoList'>
          <VideoList
            data={data}
            title={ui.headerVideoList}
            setActiveVideo={setActiveVideo}
          />
        </div >
      </div >
    );

}

export default App;
