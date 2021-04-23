import React, { useState, useEffect } from "react";
import "../style/Header.css";
import { CSSTransition } from "react-transition-group";
import {Link} from "react-router-dom";

export default function Header() {
    const [isNavVisible, setNavVisibility] = useState(false);
    const [isSmallScreen, setIsSmallScreen] = useState(false);

    useEffect(() => {
        const mediaQuery = window.matchMedia("(max-width: 700px)");
        mediaQuery.addListener(handleMediaQueryChange);
        handleMediaQueryChange(mediaQuery);

        return () => {
            mediaQuery.removeListener(handleMediaQueryChange);
        };
    }, []);

    const handleMediaQueryChange = (mediaQuery) => {
        if (mediaQuery.matches) {
            setIsSmallScreen(true);
        } else {
            setIsSmallScreen(false);
        }
    };

    const toggleNav = () => {
        setNavVisibility(!isNavVisible);
    };

    return (
        <header className="Header">
            <CSSTransition
                in={!isSmallScreen || isNavVisible}
                timeout={350}
                classNames="NavAnimation"
                unmountOnExit
            >
                <nav className="Nav">
                    <a href="/">Player</a>
                    <a href="/upload">Upload</a>
                    <a href="/progress">Progress</a>
                    <a href="/worker">Worker</a>
                </nav>
            </CSSTransition>
            <button onClick={toggleNav} className="Burger">
        <span role="img" aria-label="">
          {" "}
            🍔{" "}
        </span>
            </button>
        </header>
    );
}
