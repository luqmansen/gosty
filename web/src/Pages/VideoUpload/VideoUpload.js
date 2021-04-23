import Dropzone from "react-dropzone-uploader";
import "react-dropzone-uploader/dist/styles.css";
import {APISERVER_HOST, VIDEO_UPLOAD_ENDPOINT} from "../../Constant";


export default function VideoUpload() {
    // specify upload params and url for your files
    const getUploadParams = ({ meta }) => {
        return {
            url: APISERVER_HOST + VIDEO_UPLOAD_ENDPOINT,
        };
    };

    // called every time a file's `status` changes
    const handleChangeStatus = ({ meta, file }, status) => {
        console.log(status, meta, file);
    };

    // receives array of files that are done uploading when submit button is clicked
    const handleSubmit = files => {
        console.log(files.map(f => f.meta));
    };

    return (
        <div className="App">
            <h1>Upload Video</h1>
            <Dropzone
                getUploadParams={getUploadParams}
                onChangeStatus={handleChangeStatus}
                onSubmit={handleSubmit}
                accept=".mp4"
                inputContent="Choose video with format MP4"
                inputWithFilesContent="Add more videos"
                submitButtonContent="Done"
            />
        </div>
    );
}
