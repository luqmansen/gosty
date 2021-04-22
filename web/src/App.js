import React from "react";
import {
    BrowserRouter as Router,
    Switch,
    Route,
    Link
} from "react-router-dom";
import PlayerPage from "./PlayerPage";
import WorkerPage from "./WorkerPage";
import ProgressPage from "./ProgressPage";

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
                </Switch>
            </div>
        </Router>
    );
}

