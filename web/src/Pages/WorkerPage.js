import {Component} from "react";
import 'react-virtualized/styles.css';
import {Column, Table} from 'react-virtualized';
import {APISERVER_HOST, WORKER_STATUS_ENDPOINT, WORKER_STATUS} from "../Constant";

class WorkerPage extends Component {

    state = {
        data: []
    }


    //TODO: this stupid, need to apply websocket or sse
    // instead of requesting every 100ms
    async componentDidMount() {
        try {
            setInterval(async () => {
                const res = await fetch(APISERVER_HOST + WORKER_STATUS_ENDPOINT);
                const blocks = await res.json();
                blocks.map(w => {
                    w.status = WORKER_STATUS[w.status]
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
                <h1>Worker List</h1>
                <Table
                    rowClassName='table-row'
                    headerHeight={40}
                    width={900}
                    height={300}
                    rowHeight={40}
                    rowCount={this.state.data.length}
                    rowGetter={({index}) => this.state.data[index]}
                >
                    <Column
                        label='Id'
                        dataKey='id'
                        width={200}
                    />
                    <Column
                        label='Worker Name'
                        dataKey='worker_pod_name'
                        width={250}
                    />
                    <Column
                        label='Status'
                        dataKey='status'
                        width={100}
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
    }
}

export default WorkerPage;