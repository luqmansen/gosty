import {Component} from "react";
import 'react-virtualized/styles.css';
import '../../style/style.css';
import {Column, Table} from 'react-virtualized';
import {APISERVER_HOST, TASK_KIND, TASK_PROGRESS_ENDPOINT, TASK_STATUS} from "../../Constant";
import {msToTime} from "../../Utils";
import {tableData} from "./Tabledata";

class ProgressPage extends Component {

    state = {
        data: []
    }

    //  This page is deprecated, use V2
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
                <div className="container">
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





export default ProgressPage;