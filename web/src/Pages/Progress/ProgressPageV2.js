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
        setData(blocks)
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


const tableData = (v) => {
    let data = ""
    if (v.task_list.length > 0) {
        data = (<Table
            rowClassName='table-row'
            headerHeight={40}
            width={1000}
            height={v.task_list.length * 40}
            rowHeight={40}
            rowCount={v.task_list.length}
            rowGetter={({index}) => v.task_list[index]}
        >
            <Column
                label='Task Kind'
                dataKey='kind'
                width={300}
            />
            <Column
                label='Target'
                dataKey='target'
                width={300}
            />
            <Column
                label='Status'
                dataKey='status'
                width={250}
            />
            <Column
                label='Worker'
                dataKey='worker'
                width={250}
            />
            <Column
                label='task_submitted'
                dataKey='task_submitted'
                width={300}
            />
            <Column
                label='task_completed'
                dataKey='task_completed'
                width={300}
            />
            <Column
                label='duration'
                dataKey='task_duration'
                width={300}
            />
        </Table>)
    } else {
        data = <p>File on queue</p>
    }
    return (
        <div>
            <p><b>File : {v.origin_video.file_name}</b></p>
            <p>Total Duration: {msToTime(v.total_duration / 1e+6)}</p>
            {data}
        </div>
    )
}

export default ProgressPageV2;