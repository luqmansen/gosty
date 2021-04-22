import {Component} from "react";
import 'react-virtualized/styles.css';
import {Column, Table} from 'react-virtualized';

class ProgressPage extends Component {

    state = {
        data: []
    }

    TASK_KIND = ["NEW", "SPLIT", "MERGE", "TRANSCODE", "DASH"]
    TASK_STATUS = ["QUEUED", "DONE", "ON PROGRESS", "FAILED"]

    //TODO: this stupid, need to apply websocket or sse
    // instead of requesting every 100ms
    async componentDidMount() {
        try {
            setInterval(async () => {
                const res = await fetch('http://localhost:8000/progress');
                const blocks = await res.json();
                blocks.map(w => {
                    w.task_list.map(t => {
                        t.kind = this.TASK_KIND[t.kind]
                    })
                })
                blocks.map(w => {
                    w.task_list.map(t => {
                        t.status = this.TASK_STATUS[t.status]
                    })
                })
                blocks.map(w => {
                    w.task_list.map(t => {
                        if (t.task_transcode != null) {
                            t.target = t.task_transcode.target_res
                        } else if (t.task_split != null) {
                            t.target = t.task_split.splited_video.length
                        } else if (t.task_merge != null){
                            t.target = t.task_merge.list_video.length
                        }
                            })
                })
                blocks.map(w => {
                    w.task_list.map(t => {
                        t.task_duration = msToTime(t.task_duration / 1e+6)
                    })
                })
                this.setState({
                    data: blocks,
                })
            }, 100);

        } catch (e) {
            console.log(e);
        }
    }

    render() {
        return (
            <div class="container">
                {this.state.data.map(v => tableData(v))}
            </div>
        )
    }
}

function msToTime(ms) {
    let seconds = (ms / 1000).toFixed(1);
    let minutes = (ms / (1000 * 60)).toFixed(1);
    let hours = (ms / (1000 * 60 * 60)).toFixed(1);
    let days = (ms / (1000 * 60 * 60 * 24)).toFixed(1);
    if (seconds < 60) return seconds + " Sec";
    else if (minutes < 60) return minutes + " Min";
    else if (hours < 24) return hours + " Hrs";
    else return days + " Days"
}

const tableData = (v) => {
    return (
        <div>
            <p>File Name: {v.origin_video.file_name}</p>

            <Table
                rowClassName='table-row'
                headerHeight={40}
                width={900}
                height={300}
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
                    label='task_duration'
                    dataKey='task_duration'
                    width={300}
                />
            </Table>
        </div>
    )
}

export default ProgressPage;