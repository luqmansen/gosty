import React from "react";
import {
    BrowserRouter as Router,
    Switch,
    Route,
    Link
} from "react-router-dom";
import PlayerPage from "./Pages/PlayerPage";
import WorkerPage from "./Pages/WorkerPage";
import ProgressPage from "./Pages/ProgressPage";
import {VideoUpload} from "./Pages/VideoUpload/VideoUpload";

export default function BasicExample() {
    return (
        <Router>
            <div>
                <ul>
                    <li>
                        <Link to="/">Home</Link>
                    </li>
                    <li>
                        <Link to="/worker">Worker Info</Link>
                    </li>
                    <li>
                        <Link to="/progress">Task Progress</Link>
                    </li>
                    <li>
                        <Link to="/upload">Upload Video</Link>
                    </li>
                </ul>

                <hr />

                {/*
          A <Switch> looks through all its children <Route>
          elements and renders the first one whose path
          matches the current URL. Use a <Switch> any time
          you have multiple routes, but you want only one
          of them to render at a time
        */}
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
    );
}


