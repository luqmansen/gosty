import {Component} from "react";
import 'react-virtualized/styles.css';
import {Column, Table} from 'react-virtualized';

class WorkerPage extends Component {

    state = {
        data: []
    }

    WORKER_STATUS = ["IDLE", "WORKING", "TERMINATED"]

    //TODO: this stupid, need to apply websocket or sse
    // instead of requesting every 100ms
    async componentDidMount() {
        try {
            setInterval(async () => {
                const res = await fetch('http://localhost:8000/worker');
                const blocks = await res.json();
                blocks.map(w => {w.status = this.WORKER_STATUS[w.status]})
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
                    rowGetter={({ index }) => this.state.data[index]}
                >
                    <Column
                        label='Id'
                        dataKey='id'
                        width={300}
                    />
                    <Column
                        label='Worker Name'
                        dataKey='worker_pod_name'
                        width={250}
                    />
                    <Column
                        label='Status'
                        dataKey='status'
                        width={300}
                    />
                    <Column
                        label='Working On'
                        dataKey='working_on'
                        width={300}
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