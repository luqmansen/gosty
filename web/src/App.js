import React from "react";
import {
    BrowserRouter as Router,
    Switch,
    Route,
    Link
} from "react-router-dom";
import PlayerPage from "./Pages/VideoPlayer/PlayerPage";
import WorkerPage from "./Pages/Worker/WorkerPage";
import ProgressPage from "./Pages/Progress/ProgressPage";
import {VideoUpload} from "./Pages/VideoUpload/VideoUpload";
import Header from "./Components/Header";

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
                        <WorkerPage />
                    </Route>
                    <Route path="/progress">
                        <ProgressPage />
                    </Route>
                    <Route path="/upload">
                        <VideoUpload width={400} height={300} />
                    </Route>

                </Switch>
            </div>
        </Router>
        </>
    );
}


