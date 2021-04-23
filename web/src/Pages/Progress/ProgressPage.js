import {Component} from "react";
import 'react-virtualized/styles.css';
import '../../style/style.css';
import {Column, Table} from 'react-virtualized';
import {APISERVER_HOST, TASK_KIND, TASK_PROGRESS_ENDPOINT, TASK_STATUS} from "../../Constant";

class ProgressPage extends Component {

    state = {
        data: []
    }

    //TODO: update this stupid function to use websocket or sse
    // instead of requesting every few ms
    async componentDidMount() {
        try {
            setInterval(async () => {
                const res = await fetch(APISERVER_HOST + TASK_PROGRESS_ENDPOINT);
                if (res.status === 200) {
                    const blocks = await res.json();
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
                    this.setState({
                        data: blocks,
                    })
                } else {
                    this.setState({
                        data: [],
                    })
                }

            }, 500);

        } catch (e) {
            console.log(e);
        }
    }

    render() {
        return (
            <>
                <div class="container">
                    <h1>Task Progress</h1>
                    {(() => {
                        if (this.state.data.length > 0) {
                            return (this.state.data.map(v => tableData(v)))
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
            <p><b>File : {v.origin_video.file_name}</b></p>
            <p>Total Duration: {msToTime(v.total_duration / 1e+6)}</p>

            <Table
                rowClassName='table-row'
                headerHeight={40}
                width={1000}
                height={v.task_list.length * 50}
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
            </Table>
        </div>
    )
}

export default ProgressPage;