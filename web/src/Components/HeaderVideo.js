import React, { Component } from 'react';
import '../style/Header.css'

/*
 * Simple header for the application.
*/
class HeaderVideo extends Component {
    constructor(props) {
        super(props);
    }
    render() {
        return (
            <div className='header'>
                <h1>{this.props.title}</h1>
            </div>
        );
    }
};

export default HeaderVideo;
