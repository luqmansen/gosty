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
import {processData} from "./ProcessData";

//Progress Page V2 use SSE for updating state
const ProgressPageV2 = () => {

    const [data, setData] = useState([])
    let search = window.location.search;
    let params = new URLSearchParams(search);
    let showAnalytics = params.get('analytics');
    if (showAnalytics == null){
        showAnalytics = false
    }

    useEffect(() => {
        (async () => {
            const res = await fetch(APISERVER_HOST + TASK_PROGRESS_ENDPOINT);
            if (res.status === 200) {
                const blocks = await res.json();
                setData(processData(blocks))
            }
        })()

    }, [])

    useEffect(() => {
        let eventSource = new EventSource(`${APISERVER_HOST}${EVENTSTREAM_ENDPOINT}?stream=${TASK_STREAM_NAME}`)
        eventSource.onmessage = (event) => {
            setData(processData(JSON.parse(event.data)))
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

    return (
        <>
            <div className="container">
                <h1>Task Progress</h1>
                {(() => {
                    if (data.length > 0) {
                        return (data.map(v => tableData(v, showAnalytics)))
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