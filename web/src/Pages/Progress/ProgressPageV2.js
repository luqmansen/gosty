import {Component, useEffect, useState} from "react";
import 'react-virtualized/styles.css';
import '../../style/style.css';
import {Column, Table} from 'react-virtualized';
import {
    APISERVER_HOST,
    EVENTSTREAM_ENDPOINT,
    TASK_KIND,
    TASK_PROGRESS_ENDPOINT,
    TASK_STATUS, TASK_STREAM_NAME,
    WORKER_STREAM_NAME
} from "../../Constant";
import {msToTime} from "../../Utils";
import {tableData} from "./Tabledata";

//Progress Page V2 use SSE for updating state
const ProgressPageV2 = () => {

    const [data, setData] = useState([])

    useEffect(() => {
        (async () => {
            const res = await fetch(APISERVER_HOST + TASK_PROGRESS_ENDPOINT);
            if (res.status === 200) {
                const blocks = await res.json();
                processData(blocks)
            }
        })()

    }, [])

    useEffect(() => {
        let eventSource = new EventSource(`${APISERVER_HOST}${EVENTSTREAM_ENDPOINT}?stream=${TASK_STREAM_NAME}`)
        eventSource.onmessage = (event) => {
            processData(JSON.parse(event.data))
        }
        eventSource.onerror = e => {
            eventSource.close()
            console.log(e)
        }

        return () => {
            console.log("PROGRESS STREAM CLOSED")
            eventSource.close()
        }
    }, [])

    const processData = (blocks) => {
        blocks.map(w => {
            w.task_list.map(t => {
                t.kind = TASK_KIND[t.kind]
            })
        })
        blocks.map(w => {
            w.task_list.map(t => {
                t.status = TASK_STATUS[t.status]
            })
        })
        blocks.map(w => {
            w.task_list.map(t => {
                if (t.task_transcode != null) {
                    t.target = t.task_transcode.target_res
                } else if (t.task_split != null) {
                    if (t.task_split.splited_video != null) {
                        t.target = t.task_split.splited_video.length
                    }
                } else if (t.task_merge != null) {
                    if (t.task_merge.list_video != null) {
                        t.target = t.task_merge.list_video.length
                    }
                }
            })
        })
        blocks.map(w => {
            w.task_list.map(t => {
                t.task_duration = msToTime(t.task_duration / 1e+6)
            })
        })
        blocks.map(w => {
            w.task_list.map((t, idx) => {
                t.no = idx + 1
            })
        })
        setData(blocks.reverse())
    }

    return (
        <>
            <div className="container">
                <h1>Task Progress</h1>
                {(() => {
                    if (data.length > 0) {
                        return (data.map(v => tableData(v)))
                    } else {
                        return (
                            <p>No Task</p>
                        )
                    }
                })()}
            </div>
        </>
    )
}

export default ProgressPageV2;