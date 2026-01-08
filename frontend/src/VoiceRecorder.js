import { useState, useRef, useEffect } from 'react';
import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:9090';

// Convert blob to base64
const blobToBase64 = (blob) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onloadend = () => {
      const base64String = reader.result.split(',')[1]; // Remove data:audio/wav;base64, prefix
      resolve(base64String);
    };
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
};

export const useVoiceRecorder = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [audioBlob, setAudioBlob] = useState(null);
  const [recordingTime, setRecordingTime] = useState(0);
  const mediaRecorderRef = useRef(null);
  const streamRef = useRef(null);
  const chunksRef = useRef([]);
  const timerRef = useRef(null);

  const startRecording = async () => {
    try {
      // Check if we're on a secure context (HTTPS or localhost)
      // Browsers allow microphone access on localhost even over HTTP
      const hostname = window.location.hostname;
      const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1' || hostname === '0.0.0.0' || hostname === '';
      const isSecure = window.isSecureContext || window.location.protocol === 'https:';
      
      if (!isSecure && !isLocalhost) {
        throw new Error('SECURE_CONTEXT_REQUIRED');
      }

      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;
      
      const mediaRecorder = new MediaRecorder(stream, {
        mimeType: 'audio/webm;codecs=opus'
      });
      
      mediaRecorderRef.current = mediaRecorder;
      chunksRef.current = [];
      
      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          chunksRef.current.push(event.data);
        }
      };
      
      mediaRecorder.onstop = () => {
        const blob = new Blob(chunksRef.current, { type: 'audio/webm' });
        setAudioBlob(blob);
        // Stop all tracks
        if (streamRef.current) {
          streamRef.current.getTracks().forEach(track => track.stop());
        }
      };
      
      mediaRecorder.start();
      setIsRecording(true);
      setRecordingTime(0);
      
      // Start timer
      timerRef.current = setInterval(() => {
        setRecordingTime(prev => prev + 1);
      }, 1000);
      
    } catch (error) {
      console.error('Error starting recording:', error);
      throw error;
    }
  };

  const stopRecording = () => {
    if (mediaRecorderRef.current) {
      // Check if recorder is actually recording (state might be async)
      if (mediaRecorderRef.current.state === 'recording') {
        mediaRecorderRef.current.stop();
      }
      setIsRecording(false);
      
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
      
      // Stop all tracks
      if (streamRef.current) {
        streamRef.current.getTracks().forEach(track => track.stop());
        streamRef.current = null;
      }
    }
  };

  const resetRecording = () => {
    setAudioBlob(null);
    setRecordingTime(0);
    chunksRef.current = [];
  };

  useEffect(() => {
    return () => {
      // Cleanup on unmount
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
      if (streamRef.current) {
        streamRef.current.getTracks().forEach(track => track.stop());
      }
    };
  }, []);

  return {
    isRecording,
    audioBlob,
    recordingTime,
    startRecording,
    stopRecording,
    resetRecording
  };
};

export const registerVoice = async (name, audioBlob) => {
  try {
    const base64Audio = await blobToBase64(audioBlob);
    
    const response = await axios.post(`${API_BASE_URL}/api/voice/register`, {
      name: name,
      audio_data: base64Audio,
      audio_format: 'webm'
    });
    
    return response.data;
  } catch (error) {
    console.error('Error registering voice:', error);
    throw error;
  }
};

export const recognizeVoice = async (audioBlob) => {
  try {
    const base64Audio = await blobToBase64(audioBlob);
    
    const response = await axios.post(`${API_BASE_URL}/api/voice/recognize`, {
      audio_data: base64Audio,
      audio_format: 'webm'
    });
    
    return response.data;
  } catch (error) {
    console.error('Error recognizing voice:', error);
    throw error;
  }
};

export const sendVoiceToChat = async (audioBlob) => {
  try {
    const base64Audio = await blobToBase64(audioBlob);
    
    const response = await axios.post(`${API_BASE_URL}/api/chat`, {
      audio_data: base64Audio,
      audio_format: 'webm'
    });
    
    return response.data;
  } catch (error) {
    console.error('Error sending voice to chat:', error);
    throw error;
  }
};

// Format time in MM:SS
export const formatTime = (seconds) => {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
};

