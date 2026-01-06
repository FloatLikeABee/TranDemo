# Frontend Voice Recognition Module

## Overview

The frontend voice recognition module provides a complete user interface for voice registration and attendance logging through voice commands.

## Components

### 1. VoiceRecorder.js
A custom React hook and utility functions for voice recording and API communication.

**Exports:**
- `useVoiceRecorder()` - Hook for managing voice recording state
- `registerVoice(name, audioBlob)` - Register a voice profile
- `recognizeVoice(audioBlob)` - Recognize a speaker from voice
- `sendVoiceToChat(audioBlob)` - Send voice input to chat API
- `formatTime(seconds)` - Format time as MM:SS

**Features:**
- MediaRecorder API integration
- Automatic audio format conversion (WebM)
- Base64 encoding for API transmission
- Recording timer
- Cleanup on unmount

### 2. VoiceRegistrationModal.js
A modal component for voice profile registration.

**Features:**
- Step-by-step registration process
- Real-time recording indicator
- Audio preview before submission
- Error handling and retry functionality
- Success confirmation

**Steps:**
1. Input name
2. Record voice sample
3. Preview and submit
4. Success confirmation

### 3. App.js Integration
Enhanced main app component with voice features.

**New Features:**
- Voice registration button in header
- Voice attendance button (📢) in input area
- Automatic voice processing after recording
- Voice mode indicator
- Integration with existing chat system

## User Interface

### Voice Registration Flow

1. **Click "Register Voice" button** in header
2. **Enter name** in modal
3. **Click "Start Recording"**
4. **Speak**: "Your Name, I'm here" or "Your Name, Punch in"
5. **Click "Stop Recording"**
6. **Preview audio** and click "Register Voice"
7. **Success confirmation**

### Voice Attendance Flow

1. **Click 📢 button** (voice attendance button)
2. **Speak**: "Your Name, I'm here" or "Your Name, Punch in"
3. **Click 📢 again** to stop recording
4. **System automatically processes** and responds:
   - If recognized: "Punched in" or "Gotcha!"
   - If not recognized: "Sorry, you're not in our school."

## UI Components

### Header Actions
- **Register Voice Button**: Opens voice registration modal
- Styled with orange accent color
- Hover effects

### Input Area
- **Voice Attendance Button (📢)**: 
  - Green when idle
  - Red with timer when recording
  - Shows recording time
  - Disabled during text input or loading

### Voice Registration Modal
- **Modern dark theme** matching app design
- **Animated recording indicator** with pulse effect
- **Audio preview** with native controls
- **Responsive design** for mobile devices

## Styling

### CSS Classes

- `.voice-modal-overlay` - Modal backdrop
- `.voice-modal` - Modal container
- `.voice-recording-indicator` - Recording animation
- `.voice-pulse` - Pulsing microphone icon
- `.voice-attendance-button` - Attendance button
- `.voice-register-button` - Registration button

### Color Scheme
- **Orange (#FF8C00)**: Primary actions, accents
- **Green (#4CAF50)**: Voice attendance, success
- **Red (#ff4444)**: Recording state, errors
- **Dark (#1a1a1a, #2a2a2a)**: Backgrounds

## API Integration

### Endpoints Used

1. **POST /api/voice/register**
   - Registers new voice profile
   - Sends: name, audio_data (base64), audio_format

2. **POST /api/voice/recognize**
   - Recognizes speaker from voice
   - Sends: audio_data (base64), audio_format

3. **POST /api/chat**
   - Sends voice input through chat interface
   - Sends: audio_data (base64), audio_format
   - Returns: ChatResponse with automated message

## Audio Format

- **Format**: WebM (opus codec)
- **Encoding**: Base64
- **Capture**: MediaRecorder API
- **Browser Support**: Modern browsers (Chrome, Edge, Firefox, Safari)

## User Experience Features

1. **Visual Feedback**:
   - Recording indicator with pulse animation
   - Timer display during recording
   - Button state changes (idle → recording → processing)

2. **Error Handling**:
   - Permission denied messages
   - Network error handling
   - Retry functionality

3. **Accessibility**:
   - Clear button labels
   - Tooltips for actions
   - Keyboard navigation support

4. **Responsive Design**:
   - Mobile-friendly modal
   - Touch-optimized buttons
   - Adaptive layouts

## Usage Examples

### Register Voice
```javascript
import { registerVoice } from './VoiceRecorder';

const audioBlob = // ... recorded audio
const profile = await registerVoice('John Doe', audioBlob);
```

### Recognize Voice
```javascript
import { recognizeVoice } from './VoiceRecorder';

const audioBlob = // ... recorded audio
const result = await recognizeVoice(audioBlob);
if (result.recognized) {
  console.log(`Recognized: ${result.name}`);
  console.log(`Response: ${result.message}`);
}
```

### Use Voice Recorder Hook
```javascript
import { useVoiceRecorder } from './VoiceRecorder';

function MyComponent() {
  const {
    isRecording,
    audioBlob,
    recordingTime,
    startRecording,
    stopRecording,
    resetRecording
  } = useVoiceRecorder();

  return (
    <div>
      <button onClick={startRecording}>Start</button>
      <button onClick={stopRecording}>Stop</button>
      {isRecording && <p>Recording: {recordingTime}s</p>}
    </div>
  );
}
```

## Browser Compatibility

- **Chrome/Edge**: Full support
- **Firefox**: Full support
- **Safari**: Full support (iOS 14.5+)
- **Opera**: Full support

## Future Enhancements

- Real-time voice activity detection
- Multiple voice sample registration
- Voice quality indicators
- Offline voice recognition
- Voice command expansion
- Multi-language support
- Voice profile management UI

