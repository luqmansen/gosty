import {Component} from "react";
import 'react-virtualized/styles.css';
import '../../style/style.css';
import {Column, Table} from 'react-virtualized';
import {APISERVER_HOST, TASK_KIND, TASK_PROGRESS_ENDPOINT, TASK_STATUS} from "../../Constant";
import {msToTime} from "../../Utils";
import {tableData} from "./Tabledata";
import {processData} from "./ProcessData";

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
                    this.setState({
                        data: processData(blocks),
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