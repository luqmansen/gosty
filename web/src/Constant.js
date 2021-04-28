export const APISERVER_HOST = process.env.REACT_APP_APISERVER_HOST || "http://192.168.56.107:30000"
// export const APISERVER_HOST = process.env.REACT_APP_APISERVER_HOST || "http://localhost:8000"
export const FILESERVER_HOST = process.env.REACT_APP_FILESERVER_HOST || "http://192.168.56.107:30001"
// export const FILESERVER_HOST = process.env.REACT_APP_FILESERVER_HOST || "http://localhost:8001"

export const EVENTSTREAM_ENDPOINT = "/events"
export const WORKER_STREAM_NAME = "worker"
export const TASK_STREAM_NAME = "task"

export const VIDEO_UPLOAD_ENDPOINT = "/video/upload"
export const VIDEO_PLAYLIST_ENDPOINT = "/playlist"
export const WORKER_STATUS_ENDPOINT = "/worker"
export const TASK_PROGRESS_ENDPOINT = "/progress"

export const WORKER_STATUS_READY = "READY"
export const WORKER_STATUS_WORKING = "WORKING"
export const WORKER_STATUS_TERMINATED = "TERMINATED"
export const WORKER_STATUS = [WORKER_STATUS_READY, WORKER_STATUS_WORKING, WORKER_STATUS_TERMINATED]


export const TASK_KIND = ["NEW", "SPLIT", "MERGE", "TRANSCODE", "DASH"]
export const TASK_STATUS = ["QUEUED", "DONE", "ON PROGRESS", "FAILED"]
