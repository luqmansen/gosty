import {APISERVER_HOST, WORKER_SCALE_ENDPOINT} from "../../Constant";
import React, {useState} from "react";

const WORKER_FORMDATA_SCALE_KEY = "replicanum"

const ScaleWorker = () => {
    const [replica, setReplica] = useState("1")

    const submit = e => {
        e.preventDefault()
        const data = new FormData()
        data.append(WORKER_FORMDATA_SCALE_KEY, replica)
        fetch(APISERVER_HOST + WORKER_SCALE_ENDPOINT, {method: 'POST', body: data})
            .then(res => res.json())
            .then(json => console.log(json))
    }

    return (
            <form style={{paddingTop: 10, paddingBottom: 10}} onSubmit={submit}>
                <label>
                    Scale Worker Number :
                <input
                    type="text"
                    name="replica[num]"
                    onChange={e => setReplica(e.target.value)}
                />
                </label>
                <input type="submit" name="Submit"/>
            </form>
    )
}


export default ScaleWorker

