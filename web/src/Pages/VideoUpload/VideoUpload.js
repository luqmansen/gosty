import React, {useRef, useState} from "react";
import "./style.css"
import {APISERVER_HOST, VIDEO_UPLOAD_ENDPOINT} from "../../Constant";

export const VideoUpload = (props) => {
    const { width, height } = props;

    const inputRef = useRef();
    const [source, setSource] = useState();
    const [file, setFile] = useState();

    const handleFileChange = (event) => {
        const file = event.target.files[0];
        const url = URL.createObjectURL(file);
        setSource(url);
        setFile(file)
    };

    const handleChoose = (event) => {
        inputRef.current.click();
    };

    const handleSubmit = (event) => {
        event.preventDefault()
        let formData = new FormData();
        formData.append("file", file)

        fetch(APISERVER_HOST + VIDEO_UPLOAD_ENDPOINT, {
            method: 'POST',
            body: formData,
        }).then(function(response) {
            console.log(response)
            return response.json();
        });

    }

    return (
        <div className="VideoInput">
            <input
                ref={inputRef}
                className="VideoInput_input"
                type="file"
                onChange={handleFileChange}
                accept=".mov,.mp4"
            />
            {!source && <button onClick={handleChoose}>Choose</button>}
            {source && (
                <video
                    className="VideoInput_video"
                    width="100%"
                    height={height}
                    controls
                    src={source}
                />
            )}
            <div className="VideoInput_footer">{source || "Nothing selected"}</div>
            <button className="submitButton"
                    type="submit"
                    onClick={(e) => handleSubmit(e)}>Upload
            </button>
        </div>
    );
}
