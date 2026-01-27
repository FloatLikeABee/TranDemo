import React from 'react';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import Forms from './Forms';
import FormAnswers from './FormAnswers';
import './FormsApp.css';

const FormsApp = () => {
  const location = useLocation();

  return (
    <div className="forms-app">
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
          </ul>
        </div>
      </nav>
      <Routes>
        <Route path="/" element={<Forms />} />
        <Route path="/answers" element={<FormAnswers />} />
      </Routes>
    </div>
  );
};

export default FormsApp;
