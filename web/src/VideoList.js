import React, { Component } from 'react';
import VideoTile from './VideoTile.js';
import Loader from './Loader.js';

class VideoList extends Component {
    constructor(props) {
        super(props);
    }
    render() {
        let data = this.props.data;
        let listVideoTiles = data.length === 0 ? <Loader /> :
            data.map((video, index) =>
                <VideoTile
                    key={index}
                    id={video.id}
                    title={video.file_name}
                    dash={video.dash_file}
                    video={video}
                    setActiveVideo={this.props.setActiveVideo}
                />
            );
        return (
            <div>
                <div>{listVideoTiles}</div>
            </div>
        );
    }
}

export default VideoList;
