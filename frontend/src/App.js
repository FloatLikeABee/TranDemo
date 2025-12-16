import React, { useState, useRef, useEffect } from 'react';
import axios from 'axios';
import './App.css';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

function App() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef(null);
  const inputRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async (e) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    const userMessage = input.trim();
    setInput('');
    setMessages(prev => [...prev, { type: 'user', content: userMessage }]);
    setLoading(true);

    try {
      const response = await axios.post(`${API_BASE_URL}/api/chat`, {
        message: userMessage
      });

      const aiResponse = response.data.response;
      const sql = response.data.sql;

      setMessages(prev => [...prev, {
        type: 'assistant',
        content: aiResponse,
        sql: sql
      }]);
    } catch (error) {
      console.error('Error:', error);
      setMessages(prev => [...prev, {
        type: 'error',
        content: error.response?.data?.error || 'Failed to get response. Please try again.'
      }]);
    } finally {
      setLoading(false);
      inputRef.current?.focus();
    }
  };

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    // You could add a toast notification here
  };

  return (
    <div className="app">
      <div className="chat-container">
        <div className="chat-header">
          <h1>Transfinder Form/Report Assistant</h1>
          <p>Transform your data into beautiful forms and comprehensive reports with AI-powered precision</p>
        </div>

        <div className="messages-container">
          {messages.length === 0 && (
            <div className="welcome-message">
              <div className="welcome-icon">🤖</div>
              <h2>Welcome to Transfinder Form/Report Assistant</h2>
              <p>Start by describing the form or report you need. For example:</p>
              <ul>
                <li>"Create a student enrollment form with all required fields"</li>
                <li>"Generate a monthly transportation report showing route statistics"</li>
                <li>"Build a form to track student attendance by date and route"</li>
              </ul>
            </div>
          )}

          {messages.map((msg, idx) => (
            <div key={idx} className={`message ${msg.type}`}>
              <div className="message-content">
                {msg.type === 'user' && (
                  <div className="message-bubble user-bubble">
                    {msg.content}
                  </div>
                )}
                {msg.type === 'assistant' && (
                  <div className="message-bubble assistant-bubble">
                    <div className="response-text">{msg.content.replace(/Here's the SQL query based on your request:\n\n/g, '')}</div>
                    {msg.sql && (
                      <div className="sql-block">
                        <div className="sql-header">
                          <span>SQL Query</span>
                          <button
                            className="copy-button"
                            onClick={() => copyToClipboard(msg.sql)}
                            title="Copy SQL"
                          >
                            📋 Copy
                          </button>
                        </div>
                        <pre><code>{msg.sql}</code></pre>
                      </div>
                    )}
                  </div>
                )}
                {msg.type === 'error' && (
                  <div className="message-bubble error-bubble">
                    ⚠️ {msg.content}
                  </div>
                )}
              </div>
            </div>
          ))}

          {loading && (
            <div className="message assistant">
              <div className="message-bubble assistant-bubble">
                <div className="loading-dots">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        <form className="input-container" onSubmit={handleSend}>
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Type your SQL request here..."
            className="message-input"
            disabled={loading}
          />
          <button
            type="submit"
            className="send-button"
            disabled={loading || !input.trim()}
          >
            {loading ? '⏳' : '➤'}
          </button>
        </form>
      </div>
    </div>
  );
}

export default App;

