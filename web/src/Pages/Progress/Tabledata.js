import {msToTime} from "../../Utils";
import {Column, Table} from "react-virtualized";
import {GRAFANA_API_ENDPOINT} from "../../Constant";

const getStartAndEnd = (data) => {
    let end;

    let start = new Date(data.task_list[0].task_started).getTime()

    let last = data.task_list[data.task_list.length - 1];

    if (last.task_completed === "0001-01-01T00:00:00Z") {
        end = Date.now()
    } else {
        end = new Date(last.task_completed).getTime()
    }
    return [start, end]
}

const calculateElapsedTime = (data) => {
    let res = getStartAndEnd(data)
    let start = res[0]
    let end = res[1]
    let et = Math.abs(end - start)
    return msToTime(et)
}

const getWorkerNumber = (data) => {
    let workerList = []
    data.task_list.forEach((j, idx) => {
        if (j.worker !== "") {
            workerList.push(j.worker)
        }
    })
    return new Set(workerList).size
}

const getAvgCpuUsagePerNamespace = (data) => {
    let startAndEnd = getStartAndEnd(data)
    let start = startAndEnd[0].toString().slice(0, -3)
    console.log("Start from avg CPU: ", start)
    let end = startAndEnd[1].toString().slice(0, -3)
    let query = `/query_range?query=sum (rate (container_cpu_usage_seconds_total{image!="",kubernetes_io_hostname=~"^.*$"}[1m])) by (namespace)&start=${start}&end=${end}&step=10`
    var req = new XMLHttpRequest();
    req.open("GET", GRAFANA_API_ENDPOINT + query, false)
    req.setRequestHeader("x-powered-by", "CORS Anywhere")
    req.send()

    let jsonData = JSON.parse(req.response)
    let resultData = jsonData.data.result
    if (resultData.length > 0) {
        let timeSeriesData = resultData[0].values
        let aggregate = 0;
        timeSeriesData.forEach((val, idx) => {
            aggregate += parseFloat(val[1])
        })
        return (aggregate / timeSeriesData.length).toFixed(2)
    }
}

const getAvgCpuPerDeployment = (data) => {
    let startAndEnd = getStartAndEnd(data)
    let start = startAndEnd[0].toString().slice(0, -3)
    let end = startAndEnd[1].toString().slice(0, -3)
    let query = `/query_range?query=sum (rate (container_cpu_usage_seconds_total{image!="",kubernetes_io_hostname=~"^.*$",namespace="gosty"}[1m])) by (container)&start=${start}&end=${end}&step=10`
    let req = new XMLHttpRequest();
    req.open("GET", GRAFANA_API_ENDPOINT + query, false)
    req.setRequestHeader("x-powered-by", "CORS Anywhere")
    req.send()

    let jsonData = JSON.parse(req.response)
    let resultData = jsonData.data.result
    if (resultData.length > 0) {
        let avgApiServer = 0;
        resultData[1].values.forEach((val, _) => avgApiServer += parseFloat(val[1]))
        let avgFileServer = 0;
        resultData[2].values.forEach((val, _) => avgFileServer += parseFloat(val[1]))
        let avgWeb = 0;
        resultData[3].values.forEach((val, _) => avgWeb += parseFloat(val[1]))
        let avgWorker = 0;
        resultData[4].values.forEach((val, _) => avgWorker += parseFloat(val[1]))

        return <ul>
            <li>Apiserver: {(avgApiServer / resultData[1].values.length).toFixed(4)}</li>
            <li>Filserver: {(avgFileServer / resultData[2].values.length).toFixed(4)}</li>
            <li>Web: {(avgWeb / resultData[3].values.length).toFixed(4)}</li>
            <li>Worker: {(avgWorker / resultData[4].values.length).toFixed(4)}</li>
        </ul>
    }
}

const getAvgMemoryPerDeployment = (data) => {
    let startAndEnd = getStartAndEnd(data)
    let start = startAndEnd[0].toString().slice(0, -3)
    let end = startAndEnd[1].toString().slice(0, -3)
    console.log("Start from avg memory: ", start)
    let query = `/query_range?query=sum (container_memory_working_set_bytes{image!="",kubernetes_io_hostname=~"^.*$",namespace="gosty"}) by (container)&start=${start}&end=${end}&step=10`

    let req = new XMLHttpRequest();
    req.open("GET", GRAFANA_API_ENDPOINT + query, false)
    req.setRequestHeader("x-powered-by", "CORS Anywhere")
    req.send()

    let jsonData = JSON.parse(req.response)
    if (jsonData.data) {
        let resultData = jsonData.data.result
        console.log(resultData)
        if (resultData.length > 0) {
            let avgApiServer = 0;
            resultData[1].values.forEach((val, _) => avgApiServer += parseFloat(val[1]))
            let avgFileServer = 0;
            resultData[2].values.forEach((val, _) => avgFileServer += parseFloat(val[1]))
            let avgWeb = 0;
            resultData[3].values.forEach((val, _) => avgWeb += parseFloat(val[1]))
            let avgWorker = 0;
            resultData[4].values.forEach((val, _) => avgWorker += parseFloat(val[1]))
            console.log("raw sum :", avgWorker)
            console.log("data length :", resultData[4].values.length)
            return <ul>
                <li>Apiserver: {((avgApiServer/resultData[1].values.length) * 1e-6).toFixed()} MB</li>
                <li>Filserver: {((avgFileServer/resultData[2].values.length) * 1e-6).toFixed()} MB</li>
                <li>Web: {((avgWeb/resultData[3].values.length) * 1e-6).toFixed()} MB</li>
                <li>Worker: {((avgWorker / resultData[4].values.length) * 1e-6).toFixed()} MB</li>
            </ul>
        }
    }
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
                width={350}
            />
            <Column
                label='task_submitted'
                dataKey='task_submitted'
                width={350}
            />
            <Column
                label='task_started'
                dataKey='task_started'
                width={350}
            />
            <Column
                label='task_completed'
                dataKey='task_completed'
                width={350}
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

    let analytics = ""
    if (window.location.pathname === "/progress/analytics") {
        analytics = <div>
            <p>Resource Summary:</p>
            <li>Worker Number: {getWorkerNumber(v)}</li>
            <li>Average CPU Usage Namespace: {getAvgCpuUsagePerNamespace(v)}</li>
            <li>Average CPU Usage per Deployment: <ul>{getAvgCpuPerDeployment(v)}</ul></li>
            <li>Average Memory Usage per Deployment: <ul>{getAvgMemoryPerDeployment(v)}</ul></li>
        </div>
    }

    return (
        <div>
            <p><b>File : {v.origin_video.file_name}</b></p>
            <p>Elapsed Time: {calculateElapsedTime(v)}</p>
            <p>Accumulated Worker Time: {msToTime(v.total_duration / 1e+6)}</p>
            {analytics}
            {data}
        </div>
    )
}

