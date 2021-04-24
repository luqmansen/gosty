import React, {Component, useEffect} from "react";
import 'react-virtualized/styles.css';
import {Column, Table} from 'react-virtualized';
import {
    APISERVER_HOST,
    WORKER_STATUS,
    EVENTSTREAM_ENDPOINT,
    WORKER_STREAM_NAME
} from "../../Constant";

const WorkerPageV2 = () => {

    const [data, setData] = React.useState([])

    useEffect(() => {
        let eventSource = new EventSource(`${APISERVER_HOST}${EVENTSTREAM_ENDPOINT}?stream=${WORKER_STREAM_NAME}`)
        eventSource.onmessage = (event) => {
            processData(event.data)
        }
    }, [])


    const processData = (eventData) => {
        let blocks = JSON.parse(eventData)
        if (blocks.length > 0) {
            blocks.map(w => {
                w.status = WORKER_STATUS[w.status]
            })
            setData(blocks)
        } else {
            setData([])
        }
    }


    if (data.length > 0) {
        return (
            <div className="container">
                <h1>Worker List</h1>
                <Table
                    rowClassName='table-row'
                    headerHeight={40}
                    width={900}
                    height={300}
                    rowHeight={40}
                    rowCount={data.length}
                    rowGetter={({index}) => data[index]}
                >
                    <Column
                        label='Id'
                        dataKey='id'
                        width={200}
                    />
                    <Column
                        label='Worker Name'
                        dataKey='worker_pod_name'
                        width={170}
                    />
                    <Column
                        label='Status'
                        dataKey='status'
                        width={180}
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