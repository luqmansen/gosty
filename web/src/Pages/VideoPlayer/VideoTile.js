import React, { Component } from 'react';
import '../../style/VideoTile.css';
import {getMpd} from "../../Utils";


class VideoTile extends Component {
    constructor(props) {
        super(props);
        this.onClickVideoTile = this.onClickVideoTile.bind(this);
        this.mpd = getMpd(this.props.dash)
    }
    onClickVideoTile() {
        // lift state up, by calling `activeVideo`
        // passed in `props` by a parent component
        this.props.setActiveVideo(this.props.video);
    }
    //TODO: add small thumbnail for each video
    render() {
        return (
            <div className='videoTile' onClick={this.onClickVideoTile}>
                <div className='videoTile__title'>
                    <div className='videoTile__title__text'>
                        {'Video:'}
                    </div>
                    <div className='videoTile__title__value'>
                        {this.props.title}
                    </div>
                </div>
                {/*<div className='videoTile__views'>*/}
                {/*    <div className='videoTile__views__text'>*/}
                {/*        {'Views:'}*/}
                {/*    </div>*/}
                {/*    <div className='videoTile__views__value'>*/}
                {/*        {this.mpd}*/}
                {/*    </div>*/}
                {/*</div>*/}
            </div>
        );
    }
}

export default VideoTile;
