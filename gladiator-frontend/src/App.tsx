import React from 'react';
import { BrowserRouter as Router, Route, Routes, Link } from 'react-router-dom';
import CharacterBuilder from './pages/CharacterBuilder';
import Arena from './pages/Arena';
import Profile from './pages/Profile';
import './App.css';

function App() {
  return (
    <Router>
      <div>
        <nav>
          <ul>
            <li>
              <Link to="/">Character Builder</Link>
            </li>
            <li>
              <Link to="/arena">Arena</Link>
            </li>
            <li>
              <Link to="/profile">Profile</Link>
            </li>
          </ul>
        </nav>

        <hr />

        <Routes>
          <Route path="/" element={<CharacterBuilder />} />
          <Route path="/arena" element={<Arena />} />
          <Route path="/profile" element={<Profile />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
