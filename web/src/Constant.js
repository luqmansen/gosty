export const APISERVER_HOST = "api"
// export const APISERVER_HOST = "http://34.72.201.178/api"
export const FILESERVER_HOST = "files"
// export const FILESERVER_HOST = "http://34.72.201.178/files"
export const GRAFANA_API_ENDPOINT = "http://luqmansen-cors-anywhere.herokuapp.com/http://34.134.106.216:8084/grafana/api/datasources/proxy/1/api/v1"

export const EVENTSTREAM_ENDPOINT = "/events"
export const WORKER_STREAM_NAME = "worker"
export const TASK_STREAM_NAME = "task"

export const VIDEO_UPLOAD_ENDPOINT = "/video/upload"
export const VIDEO_PLAYLIST_ENDPOINT = "/video/playlist"
export const WORKER_STATUS_ENDPOINT = "/worker/"
export const WORKER_SCALE_ENDPOINT = "/worker/scale"
export const TASK_PROGRESS_ENDPOINT = "/scheduler/progress"

export const WORKER_STATUS_READY = "READY"
export const WORKER_STATUS_WORKING = "WORKING"
export const WORKER_STATUS_TERMINATED = "TERMINATED"
export const WORKER_STATUS_UNREACHABLE = "UNREACHABLE"
export const WORKER_STATUS = [WORKER_STATUS_READY, WORKER_STATUS_WORKING, WORKER_STATUS_TERMINATED, WORKER_STATUS_UNREACHABLE]


export const TASK_KIND = ["NEW", "SPLIT", "MERGE", "TRANSCODE", "DASH"]
export const TASK_STATUS = ["QUEUED", "DONE", "ON PROGRESS", "FAILED"]
