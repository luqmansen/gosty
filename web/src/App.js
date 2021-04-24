import React from "react";
import {
    BrowserRouter as Router,
    Switch,
    Route,
} from "react-router-dom";
import PlayerPage from "./Pages/VideoPlayer/PlayerPage";
import WorkerPage from "./Pages/Worker/WorkerPage";
import ProgressPage from "./Pages/Progress/ProgressPage";
import Header from "./Components/Header";
import VideoUpload from "./Pages/VideoUpload/VideoUpload";
import WorkerPageV2 from "./Pages/Worker/WorkerPageV2";
import ProgressPageV2 from "./Pages/Progress/ProgressPageV2";

export default function BasicExample() {
    return (
        <>
        <Header/>
        <Router>
            <div>
                <Switch>
                    <Route exact path="/">
                        <PlayerPage />
                    </Route>
                    <Route path="/worker">
                        <WorkerPageV2 />
                    </Route>
                    <Route path="/workerv1">
                        <WorkerPage />
                    </Route>
                    <Route path="/progressv1">
                        <ProgressPage />
                    </Route>
                    <Route path="/progress">
                        <ProgressPageV2 />
                    </Route>
                    <Route path="/upload">
                        <VideoUpload/>
                    </Route>

                </Switch>
            </div>
        </Router>
        </>
    );
}


