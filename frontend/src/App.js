import React, { useState, useRef, useEffect } from 'react';
import axios from 'axios';
import SpeechRecognition, { useSpeechRecognition } from 'react-speech-recognition';
import VoiceRegistrationModal from './VoiceRegistrationModal';
import { useVoiceRecorder, sendVoiceToChat, formatTime } from './VoiceRecorder';
import './App.css';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:9090';

// Polyfill for browsers that don't support speech recognition
if (!window.SpeechRecognition && !window.webkitSpeechRecognition) {
  // Create a mock implementation
  window.SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition || (() => {
    console.warn('Speech recognition not supported in this browser');
    return null;
  });
}

function App() {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [micPermission, setMicPermission] = useState(null); // null = checking/prompt, true = granted, false = denied
  const [isMicAvailable, setIsMicAvailable] = useState(true); // Optimistically assume available, will be checked
  const [isCheckingPermission, setIsCheckingPermission] = useState(true); // Track if we're still checking
  const [browserSupportChecked, setBrowserSupportChecked] = useState(false); // Track if we've checked browser support
  const [showVoiceRegistration, setShowVoiceRegistration] = useState(false);
  const [voiceMode, setVoiceMode] = useState('text'); // 'text' or 'voice'
  const messagesEndRef = useRef(null);
  const inputRef = useRef(null);

  // Voice recorder for attendance
  const {
    isRecording: isVoiceRecording,
    audioBlob: voiceAudioBlob,
    recordingTime: voiceRecordingTime,
    startRecording: startVoiceRecording,
    stopRecording: stopVoiceRecording,
    resetRecording: resetVoiceRecording
  } = useVoiceRecorder();

  const {
    transcript,
    listening,
    resetTranscript,
    browserSupportsSpeechRecognition
  } = useSpeechRecognition();

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Request microphone permission on page load
  useEffect(() => {
    const requestMicPermission = async () => {
      setIsCheckingPermission(true);
      setBrowserSupportChecked(true);
      
      // First check if browser supports speech recognition
      if (!browserSupportsSpeechRecognition) {
        console.log('Browser does not support speech recognition');
        setIsMicAvailable(false);
        setMicPermission(false);
        setIsCheckingPermission(false);
        return;
      }

      // Check if navigator.mediaDevices is available
      if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        console.log('MediaDevices API not available');
        setIsMicAvailable(false);
        setMicPermission(false);
        setIsCheckingPermission(false);
        return;
      }

      setIsMicAvailable(true);

      // Try to check existing permission status first
      if (navigator.permissions && navigator.permissions.query) {
        try {
          const permissionStatus = await navigator.permissions.query({ name: 'microphone' });
          console.log('Permission status:', permissionStatus.state);
          
          if (permissionStatus.state === 'granted') {
            setMicPermission(true);
            setIsCheckingPermission(false);
            return; // Already granted, no need to request
          } else if (permissionStatus.state === 'denied') {
            setMicPermission(false);
            setIsCheckingPermission(false);
            return; // Already denied, don't request again
          }
          
          // Listen for permission changes
          permissionStatus.onchange = () => {
            console.log('Permission changed to:', permissionStatus.state);
            if (permissionStatus.state === 'granted') {
              setMicPermission(true);
            } else if (permissionStatus.state === 'denied') {
              setMicPermission(false);
            }
          };
        } catch (permError) {
          // Permissions API might not support 'microphone' in some browsers
          // Continue to request permission directly
          console.log('Permissions API not available, will request permission directly');
        }
      }

      // Request permission directly (this will show browser prompt)
      try {
        console.log('Requesting microphone permission...');
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        console.log('Microphone permission granted');
        setMicPermission(true);
        // Stop the stream immediately - we just needed permission
        stream.getTracks().forEach(track => track.stop());
      } catch (err) {
        console.log('Microphone permission error:', err.name, err.message);
        if (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError') {
          setMicPermission(false); // User denied permission - disable button
        } else if (err.name === 'NotFoundError' || err.name === 'DevicesNotFoundError') {
          // No microphone device found - don't disable button, user might plug one in
          setIsMicAvailable(false);
          setMicPermission(null); // Keep as null so button stays enabled
          console.log('No microphone device found, but keeping button enabled');
        } else {
          // Other error (e.g., NotReadableError) - keep as null so user can try again
          setMicPermission(null);
        }
      } finally {
        setIsCheckingPermission(false);
      }
    };

    // Small delay to ensure browserSupportsSpeechRecognition is ready
    const timer = setTimeout(() => {
      requestMicPermission();
    }, 100);

    return () => clearTimeout(timer);
  }, [browserSupportsSpeechRecognition]);

  // Update input when transcript changes
  useEffect(() => {
    if (transcript) {
      setInput(transcript);
    }
  }, [transcript]);

  const handleSend = async (e) => {
    e.preventDefault();
    if (!input.trim() || loading) return;

    // Stop listening if currently recording
    if (listening) {
      SpeechRecognition.stopListening();
    }

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

  const handleVoiceAttendance = async () => {
    if (isVoiceRecording) {
      stopVoiceRecording();
      return;
    }

    // Check if we're on HTTPS or localhost (browsers allow mic access on localhost even over HTTP)
    const hostname = window.location.hostname;
    const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '0.0.0.0' || hostname === '';
    const isSecure = window.isSecureContext || window.location.protocol === 'https:';
    
    // Only show warning if not localhost and not HTTPS
    if (!isSecure && !isLocalhost) {
      setMessages(prev => [...prev, {
        type: 'error',
        content: '⚠️ Voice features require HTTPS or localhost. For development, access via http://localhost:9090 or http://127.0.0.1:9090. For production, enable HTTPS.'
      }]);
      return;
    }

    try {
      await startVoiceRecording();
    } catch (error) {
      console.error('Error starting voice recording:', error);
      
      let errorMessage = 'Failed to start recording. ';
      if (error.message === 'SECURE_CONTEXT_REQUIRED' || error.name === 'NotAllowedError') {
        if (!isSecure && !isLocalhost) {
          errorMessage = '⚠️ Voice recording requires HTTPS or localhost. For development, access via http://localhost:9090 or http://127.0.0.1:9090. For production, enable HTTPS.';
        } else {
          errorMessage = 'Microphone permission denied. Please allow microphone access in your browser settings.';
        }
      } else if (error.name === 'NotFoundError' || error.name === 'DevicesNotFoundError') {
        errorMessage = 'No microphone found. Please connect a microphone device.';
      } else {
        errorMessage += error.message || 'Please check microphone permissions.';
      }
      
      setMessages(prev => [...prev, {
        type: 'error',
        content: errorMessage
      }]);
    }
  };

  // Handle voice recording completion
  useEffect(() => {
    if (voiceAudioBlob && !isVoiceRecording) {
      const sendVoice = async () => {
        setLoading(true);
        try {
          const response = await sendVoiceToChat(voiceAudioBlob);
          
          setMessages(prev => [...prev, {
            type: 'user',
            content: '🎤 Voice attendance'
          }]);
          
          setMessages(prev => [...prev, {
            type: 'assistant',
            content: response.response || 'Voice processed'
          }]);
          
          resetVoiceRecording();
        } catch (error) {
          console.error('Error sending voice:', error);
          setMessages(prev => [...prev, {
            type: 'error',
            content: error.response?.data?.error || 'Failed to process voice. Please try again.'
          }]);
        } finally {
          setLoading(false);
        }
      };
      
      sendVoice();
    }
  }, [voiceAudioBlob, isVoiceRecording]);

  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    // You could add a toast notification here
  };

  const handleMicClick = async () => {
    console.log('Mic button clicked!', {
      browserSupportsSpeechRecognition,
      isMicAvailable,
      micPermission,
      listening
    });

    // Check browser support first
    if (!browserSupportsSpeechRecognition) {
      console.log('Browser does not support speech recognition');
      alert('Speech recognition is not supported in this browser. Please use Chrome, Edge, or Safari.');
      return;
    }
    
    // Check mic availability - but don't block if it's just not detected yet
    if (!isMicAvailable) {
      console.log('Microphone not available, but trying anyway...');
      // Try to get devices again in case user plugged one in
      try {
        const devices = await navigator.mediaDevices.enumerateDevices();
        const audioInputs = devices.filter(device => device.kind === 'audioinput');
        if (audioInputs.length === 0) {
          alert('No microphone found. Please connect a microphone device and try again.');
          return;
        } else {
          // Found a mic, update state
          setIsMicAvailable(true);
        }
      } catch (err) {
        console.error('Error enumerating devices:', err);
        alert('Unable to check for microphone devices. Please ensure a microphone is connected.');
        return;
      }
    }

    // Request permission if not already granted
    if (micPermission !== true) {
      console.log('Requesting microphone permission...');
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        console.log('Permission granted, stopping stream');
        setMicPermission(true);
        setIsMicAvailable(true); // Update availability status
        // Stop the stream immediately, we just needed permission
        stream.getTracks().forEach(track => track.stop());
      } catch (error) {
        console.error('Microphone access error:', error);
        if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
          setMicPermission(false); // User explicitly denied - disable button
          alert('Microphone permission is required for voice input. Please enable it in your browser settings and refresh the page.');
        } else if (error.name === 'NotFoundError' || error.name === 'DevicesNotFoundError') {
          setIsMicAvailable(false);
          setMicPermission(null); // Keep button enabled, user might plug in a mic
          alert('No microphone device found. Please connect a microphone and try again.');
        } else {
          setMicPermission(null); // Other error - keep button enabled
          alert('Unable to access microphone: ' + (error.message || 'Unknown error') + '. Please check your microphone connection and browser settings.');
        }
        return;
      }
    }

    if (listening) {
      console.log('Stopping listening...');
      SpeechRecognition.stopListening();
      resetTranscript();
    } else {
      console.log('Starting listening...');
      resetTranscript();
      try {
        SpeechRecognition.startListening({ 
          continuous: true, 
          language: 'en-US',
          interimResults: true
        });
        console.log('Listening started successfully');
      } catch (error) {
        console.error('Error starting speech recognition:', error);
        alert('Failed to start voice input: ' + (error.message || 'Unknown error'));
      }
    }
  };

  // Disable mic button ONLY when permission is explicitly denied (false)
  // Otherwise, keep it enabled - let the user try and handle errors gracefully
  // This ensures button is enabled by default and only disabled when user explicitly denies permission
  const isMicDisabled = micPermission === false;
  
  // Debug logging (can be removed in production)
  useEffect(() => {
    console.log('Mic button state:', {
      browserSupportsSpeechRecognition,
      isMicAvailable,
      micPermission,
      isCheckingPermission,
      browserSupportChecked,
      isMicDisabled,
      buttonWillBeDisabled: isMicDisabled || loading
    });
  }, [browserSupportsSpeechRecognition, isMicAvailable, micPermission, isCheckingPermission, browserSupportChecked, isMicDisabled, loading]);

  return (
    <div className="app">
      <div className="chat-container">
        <div className="chat-header">
          <h1>Transfinder Form/Report Assistant</h1>
          <p>Transform your data into beautiful forms and comprehensive reports with AI-powered precision</p>
          <div className="header-actions">
            <button
              className="voice-register-button"
              onClick={() => setShowVoiceRegistration(true)}
              title="Register your voice for attendance"
            >
              🎙️ Register Voice
            </button>
          </div>
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
              <div className="voice-feature-notice">
                <p><strong>🎙️ Voice Attendance Feature:</strong></p>
                <p>Click "Register Voice" to register your voice, then use the 📢 button to log attendance by saying your name followed by "I'm here" or "Punch in".</p>
                {(() => {
                  const hostname = window.location.hostname;
                  const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '0.0.0.0' || hostname === '';
                  const isSecure = window.isSecureContext || window.location.protocol === 'https:';
                  return !isSecure && !isLocalhost;
                })() && (
                  <p className="voice-https-warning">
                    ⚠️ <strong>Note:</strong> Voice features require HTTPS or localhost. For development, access via <code>http://localhost:9090</code> or <code>http://127.0.0.1:9090</code>. For production, enable HTTPS.
                  </p>
                )}
              </div>
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

        <form className="input-container" onSubmit={handleSend} onClick={(e) => e.stopPropagation()}>
          <div className="input-wrapper">
            <input
              ref={inputRef}
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Type your SQL request here..."
              className="message-input"
              disabled={loading || listening || isVoiceRecording}
            />
            <button
              type="button"
              className={`mic-button ${listening ? 'listening' : ''} ${isMicDisabled ? 'disabled' : ''} ${(micPermission === null || isCheckingPermission) && !isMicDisabled ? 'pending' : ''}`}
              onClick={async (e) => {
                e.preventDefault();
                e.stopPropagation();
                console.log('Button onClick fired');
                await handleMicClick();
              }}
              disabled={isMicDisabled || loading || isVoiceRecording}
              title={
                isMicDisabled
                  ? 'Microphone not available or permission denied'
                  : isCheckingPermission
                  ? 'Checking microphone permission...'
                  : micPermission === null
                  ? 'Click to request microphone permission'
                  : listening
                  ? 'Stop recording'
                  : 'Start voice input'
              }
            >
              {listening ? '🛑' : (micPermission === null || isCheckingPermission) && !isMicDisabled ? '🎤?' : '🎤'}
            </button>
          </div>
          <button
            type="button"
            className={`voice-attendance-button ${isVoiceRecording ? 'recording' : ''}`}
            onClick={handleVoiceAttendance}
            disabled={loading || listening}
            title={isVoiceRecording ? 'Stop voice recording' : 'Voice attendance (say your name + "I\'m here" or "Punch in")'}
          >
            {isVoiceRecording ? `⏹ ${formatTime(voiceRecordingTime)}` : '📢'}
          </button>
          <button
            type="submit"
            className="send-button"
            disabled={loading || !input.trim() || listening || isVoiceRecording}
          >
            {loading ? '⏳' : '➤'}
          </button>
        </form>

        <VoiceRegistrationModal
          isOpen={showVoiceRegistration}
          onClose={() => setShowVoiceRegistration(false)}
          onSuccess={(name) => {
            setShowVoiceRegistration(false);
            setMessages(prev => [...prev, {
              type: 'assistant',
              content: `Voice registered successfully for ${name}! You can now use voice commands for attendance.`
            }]);
          }}
        />
      </div>
    </div>
  );
}

export default App;

