import {msToTime} from "../../Utils";
import {Column, Table} from "react-virtualized";

const calculateElapsedTime = (data) => {
    let end;
    let start = new Date(data.task_list[0].task_started)

    let last = data.task_list[data.task_list.length-1];

    if (last.task_completed !== "0001-01-01T00:00:00Z"){
        end = new Date(last.task_completed)
    } else {
        end = new Date(); // now
    }
    let et = Math.abs(end-start)
    console.log(end)
    console.log("complete" , last.task_completed)
    console.log(start)
    console.log(et)
    console.log(msToTime(et))
    return msToTime(et)
}

export const tableData = (v) => {
    let data = ""
    if (v.task_list.length > 0) {
        data = (<Table
            rowClassName='table-row'
            headerHeight={40}
            width={1500}
            height={v.task_list.length * 50}
            rowHeight={40}
            rowCount={v.task_list.length}
            rowGetter={({index}) => v.task_list[index]}
        >
            <Column
                label='No'
                dataKey='no'
                width={40}
            />
            <Column
                label='Task Kind'
                dataKey='kind'
                width={170}
            />
            <Column
                label='Target'
                dataKey='target'
                width={200}
            />

            <Column
                label='Status'
                dataKey='status'
                width={150}
            />
            <Column
                label='Worker'
                dataKey='worker'
                width={300}
            />
            <Column
                label='task_submitted'
                dataKey='task_submitted'
                width={250}
            />
            <Column
                label='task_started'
                dataKey='task_started'
                width={250}
            />
            <Column
                label='task_completed'
                dataKey='task_completed'
                width={250}
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
            <p>Elapsed Time: {calculateElapsedTime(v)}</p>
            <p>Accumulated Worker Time: {msToTime(v.total_duration / 1e+6)}</p>
            {data}
        </div>
    )
}