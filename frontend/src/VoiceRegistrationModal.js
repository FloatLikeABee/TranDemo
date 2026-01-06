import React, { useState } from 'react';
import { useVoiceRecorder, registerVoice, formatTime } from './VoiceRecorder';
import './VoiceRegistrationModal.css';

const VoiceRegistrationModal = ({ isOpen, onClose, onSuccess }) => {
  const [name, setName] = useState('');
  const [step, setStep] = useState('input'); // 'input', 'recording', 'success', 'error'
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  
  const {
    isRecording,
    audioBlob,
    recordingTime,
    startRecording,
    stopRecording,
    resetRecording
  } = useVoiceRecorder();

  const handleStartRecording = async () => {
    if (!name.trim()) {
      setError('Please enter your name');
      return;
    }
    
    // Check if we're on HTTPS or localhost (browsers allow mic access on localhost even over HTTP)
    const hostname = window.location.hostname;
    const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '0.0.0.0' || hostname === '';
    const isSecure = window.isSecureContext || window.location.protocol === 'https:';
    
    // Only block if not localhost and not HTTPS
    if (!isSecure && !isLocalhost) {
      setError('⚠️ Voice recording requires HTTPS or localhost. For development, access via http://localhost:9090 or http://127.0.0.1:9090. For production, enable HTTPS.');
      setStep('input');
      return;
    }
    
    setError('');
    setStep('recording');
    try {
      await startRecording();
    } catch (err) {
      let errorMsg = 'Failed to start recording. ';
      if (err.message === 'SECURE_CONTEXT_REQUIRED' || err.name === 'NotAllowedError') {
        if (!isSecureContext) {
          errorMsg = '⚠️ Voice recording requires HTTPS or localhost. Please use HTTPS or access via localhost/127.0.0.1.';
        } else {
          errorMsg = 'Microphone permission denied. Please allow microphone access in your browser settings.';
        }
      } else if (err.name === 'NotFoundError' || err.name === 'DevicesNotFoundError') {
        errorMsg = 'No microphone found. Please connect a microphone device.';
      } else {
        errorMsg += err.message || 'Please check microphone permissions.';
      }
      setError(errorMsg);
      setStep('input');
    }
  };

  const handleStopRecording = () => {
    stopRecording();
  };

  const handleSubmit = async () => {
    if (!audioBlob) {
      setError('Please record your voice first');
      return;
    }

    setLoading(true);
    setError('');

    try {
      await registerVoice(name.trim(), audioBlob);
      setStep('success');
      if (onSuccess) {
        onSuccess(name);
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to register voice. Please try again.');
      setStep('input');
    } finally {
      setLoading(false);
    }
  };

  const handleRetry = () => {
    resetRecording();
    setStep('input');
    setError('');
  };

  const handleClose = () => {
    resetRecording();
    setName('');
    setStep('input');
    setError('');
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="voice-modal-overlay" onClick={handleClose}>
      <div className="voice-modal" onClick={(e) => e.stopPropagation()}>
        <div className="voice-modal-header">
          <h2>Register Your Voice</h2>
          <button className="voice-modal-close" onClick={handleClose}>×</button>
        </div>

        <div className="voice-modal-body">
          {step === 'input' && (
            <>
              <p className="voice-modal-instructions">
                Enter your name and record a voice sample. Say your name clearly, followed by "I'm here" or "Punch in".
              </p>
              <input
                type="text"
                className="voice-name-input"
                placeholder="Enter your name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={loading}
              />
              {error && <div className="voice-error">{error}</div>}
              <button
                className="voice-record-button"
                onClick={handleStartRecording}
                disabled={loading || !name.trim()}
              >
                🎤 Start Recording
              </button>
            </>
          )}

          {step === 'recording' && (
            <>
              <div className="voice-recording-indicator">
                <div className="voice-pulse"></div>
                <p className="voice-recording-text">Recording... {formatTime(recordingTime)}</p>
              </div>
              <p className="voice-recording-instructions">
                Say: "{name}, I'm here" or "{name}, Punch in"
              </p>
              <button
                className="voice-stop-button"
                onClick={handleStopRecording}
              >
                ⏹ Stop Recording
              </button>
            </>
          )}

          {audioBlob && step !== 'recording' && step !== 'success' && (
            <>
              <div className="voice-preview">
                <p>Voice sample recorded ({formatTime(recordingTime)})</p>
                <audio controls src={URL.createObjectURL(audioBlob)} />
              </div>
              {error && <div className="voice-error">{error}</div>}
              <div className="voice-modal-actions">
                <button
                  className="voice-submit-button"
                  onClick={handleSubmit}
                  disabled={loading}
                >
                  {loading ? '⏳ Registering...' : '✓ Register Voice'}
                </button>
                <button
                  className="voice-retry-button"
                  onClick={handleRetry}
                  disabled={loading}
                >
                  🔄 Retry
                </button>
              </div>
            </>
          )}

          {step === 'success' && (
            <div className="voice-success">
              <div className="voice-success-icon">✓</div>
              <h3>Voice Registered Successfully!</h3>
              <p>Your voice has been registered. You can now use voice commands for attendance.</p>
              <button className="voice-close-button" onClick={handleClose}>
                Close
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default VoiceRegistrationModal;

