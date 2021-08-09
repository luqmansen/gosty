import React from "react";
import ScaleWorker from "./ScaleWorker";
import WorkerListV2 from "./WorkerListV2";


const WorkerPage = () => {
    return (
        <>
            <div className="container">
                <h1>Worker List</h1>
                <div>
                    <ScaleWorker/>
                </div>
                <div>
                    <WorkerListV2/>
                </div>
            </div>
        </>
    )
}
export default WorkerPage
