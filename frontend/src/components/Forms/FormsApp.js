import React, { useState, useEffect } from 'react';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import Forms from './Forms';
import FormAnswers from './FormAnswers';
import './FormsApp.css';

const THEME_KEY = 'forms-ui-theme';
const THEME_DEFAULT = 'default';
const THEME_NIGHTLY = 'nightly';

const FormsApp = () => {
  const location = useLocation();
  const [theme, setTheme] = useState(() => localStorage.getItem(THEME_KEY) || THEME_DEFAULT);

  useEffect(() => {
    localStorage.setItem(THEME_KEY, theme);
  }, [theme]);

  const themeClass = theme === THEME_NIGHTLY ? 'theme-nightly' : 'theme-default';

  return (
    <div className={`forms-app ${themeClass}`}>
      <nav className="forms-nav">
        <div className="nav-container">
          <Link to="/forms" className="logo">
            📝 Form Management
          </Link>
          <ul className="nav-links">
            <li>
              <Link 
                to="/forms" 
                className={location.pathname === '/forms' || location.pathname === '/forms/' ? 'active' : ''}
              >
                Forms
              </Link>
            </li>
            <li>
              <Link 
                to="/forms/answers" 
                className={location.pathname === '/forms/answers' ? 'active' : ''}
              >
                Answers
              </Link>
            </li>
            <li className="theme-switcher">
              <span className="theme-label">Theme:</span>
              <button
                type="button"
                className={`theme-btn ${theme === THEME_DEFAULT ? 'active' : ''}`}
                onClick={() => setTheme(THEME_DEFAULT)}
                title="Default (orange)"
              >
                Default
              </button>
              <button
                type="button"
                className={`theme-btn ${theme === THEME_NIGHTLY ? 'active' : ''}`}
                onClick={() => setTheme(THEME_NIGHTLY)}
                title="Nightly (deep purple & blue)"
              >
                Nightly
              </button>
            </li>
          </ul>
        </div>
      </nav>
      <div className="forms-scroll-area">
        <Routes>
          <Route path="/" element={<Forms />} />
          <Route path="/answers" element={<FormAnswers />} />
        </Routes>
      </div>
    </div>
  );
};

export default FormsApp;
