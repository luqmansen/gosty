import {useEffect, useState} from "react";
import 'react-virtualized/styles.css';
import {Column, Table} from 'react-virtualized';
import {
    APISERVER_HOST,
    WORKER_STATUS,
    EVENTSTREAM_ENDPOINT,
    WORKER_STREAM_NAME, TASK_PROGRESS_ENDPOINT, WORKER_STATUS_ENDPOINT, WORKER_STATUS_TERMINATED
} from "../../Constant";

const WorkerPageV2 = () => {

    const [data, setData] = useState([])

    useEffect(() => {
        (async () => {
            const res = await fetch(APISERVER_HOST + WORKER_STATUS_ENDPOINT);
            if (res.status === 200) {
                const blocks = await res.json();
                processData(blocks)
            }
        })()

    }, [])

    useEffect(() => {
        let eventSource = new EventSource(`${APISERVER_HOST}${EVENTSTREAM_ENDPOINT}?stream=${WORKER_STREAM_NAME}`)
        eventSource.onmessage = (event) => {
            let d = JSON.parse(event.data)
            if ((d) && (d.length > 0)){
                d.sort(
                    (a, b) => {
                        if (a.status < b.status) {
                            return -1
                        }
                        if (a.status > b.status) {
                            return 1
                        }
                        return 0;
                    }
                )
                processData(d)
            }

        }
        eventSource.onerror = e => {
            console.log(e)
        }

        return (() => {
            eventSource.close()
        })
    }, [])

    const processData = (blocks) => {
        let filtered = []

        if (blocks.length > 0) {
            blocks.map(w => {
                w.status = WORKER_STATUS[w.status]
            })
            blocks.map(w => {
                if (w.status !== WORKER_STATUS_TERMINATED) {
                    filtered.push(w)
                }
            })
        }
        if (filtered.length > 0){
            setData(filtered)
        } else {
            setData(blocks) // if no data, display previously terminated worker
        }
    }


    if (data.length > 0) {
        return (
            <div className="container">
                <h1>Worker List</h1>
                <Table
                    rowClassName='table-row'
                    headerHeight={40}
                    width={10000}
                    height={data.length * 80}
                    rowHeight={40}
                    rowCount={data.length}
                    rowGetter={({index}) => data[index]}
                >
                    <Column
                        label='Id'
                        dataKey='id'
                        width={250}
                    />
                    <Column
                        label='Worker Name'
                        dataKey='worker_pod_name'
                        width={250}

                    />
                    <Column
                        label='IP Address'
                        dataKey='ip_address'
                        width={130}
                    />
                    <Column
                        label='Status'
                        dataKey='status'
                        width={130}
                    />
                    <Column
                        label='Working On'
                        dataKey='working_on'
                        width={200}
                    />
                    <Column
                        label='Updated'
                        dataKey='updated_at'
                        width={300}
                    />
                </Table>
            </div>
        )
    } else {
        return (
            <>
                <div className="container">
                    <h1>Worker List</h1>
                    <p>No Worker Available</p>
                </div>
            </>
        )
    }


}

export default WorkerPageV2;