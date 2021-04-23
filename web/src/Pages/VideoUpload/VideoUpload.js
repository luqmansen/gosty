import React, {useRef, useState} from "react";
import "./style.css"
import {APISERVER_HOST, VIDEO_UPLOAD_ENDPOINT} from "../../Constant";

export const VideoUpload = (props) => {
    const {width, height} = props;

    const inputRef = useRef();
    const [source, setSource] = useState();
    const [file, setFile] = useState();
    const [response, setResponse] = useState({});
    const [uploadStatus, setUploadStatus] = useState(false);

    const handleFileChange = (event) => {
        const file = event.target.files[0];
        const url = URL.createObjectURL(file);
        setSource(url);
        setFile(file)
    };

    const handleChoose = (event) => {
        event.preventDefault()
        inputRef.current.click();
    };

    const handleSubmit = async (event) => {
        event.preventDefault()
        if (file){
            let formData = new FormData();
            formData.append("file", file)

            const resp = await fetch(APISERVER_HOST + VIDEO_UPLOAD_ENDPOINT, {
                method: 'POST',
                body: formData,
            })
            if (resp.status === 201) {
                const body = await resp.json()
                setUploadStatus(true)
                setResponse(body);
                console.log(response)
            }
        }
    }
    if (!uploadStatus) {
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
    } else if (response) {
        return (
            <>
                <div className="container">
                    <h3>Upload success</h3>
                    <p>Video ID :{response.id}</p>
                    <p>Filename :{response.file_name}</p>
                    <p>Size :{response.size}</p>
                    <p>Resolution :{response.width}x{response.height}</p>
                </div>
            </>
        )
    } else {
        return <h1>Empty....</h1>
    }

}
