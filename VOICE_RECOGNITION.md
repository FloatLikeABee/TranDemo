# Voice Recognition Module for Attendance Logging

## Overview

The voice recognition module enables hands-free attendance logging through natural voice commands. Users can register their voice and then use voice commands to log attendance, which is automatically processed and confirmed through the chat interface.

## Features

### 1. Voice Registration
- Users can prerecord their voice samples
- System associates voice features with user profiles
- Multiple voice samples per user for better recognition accuracy
- Voice samples stored securely in the database and filesystem

### 2. Speaker Recognition
- Matches incoming voice input against registered voice profiles
- Identifies the speaker from voice characteristics
- Returns user information if recognized

### 3. Command Interpretation
- Detects attendance-related phrases:
  - "I'm here"
  - "Punch in"
  - "Attendance register"
  - "Here"
  - "Present"
  - And similar variations

### 4. Automated Chat Responses
- **Recognized + Attendance Intent**: Responds with "Punched in" or "Gotcha!"
- **Not Recognized**: Responds with "Sorry, you're not in our school."
- Attendance is automatically logged in chat history

## API Endpoints

### Register Voice Profile
```http
POST /api/voice/register
Content-Type: application/json

{
  "name": "John Doe",
  "audio_data": "<base64_encoded_audio>",
  "audio_format": "wav"
}
```

**Response:**
```json
{
  "user_id": "abc123...",
  "name": "John Doe",
  "voice_samples": ["user_abc123_johndoe_20240101_120000.wav"],
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

### Recognize Voice
```http
POST /api/voice/recognize
Content-Type: application/json

{
  "audio_data": "<base64_encoded_audio>",
  "audio_format": "wav"
}
```

**Response (Recognized):**
```json
{
  "recognized": true,
  "user_id": "abc123...",
  "name": "John Doe",
  "transcript": "[Speech-to-text transcript]",
  "intent": "punch_in",
  "message": "Punched in"
}
```

**Response (Not Recognized):**
```json
{
  "recognized": false,
  "message": "Sorry, you're not in our school."
}
```

### List Voice Profiles
```http
GET /api/voice/profiles
```

### Delete Voice Profile
```http
DELETE /api/voice/profile/:user_id
```

## Chat Integration

Voice recognition is integrated into the chat interface. You can send voice input through the chat endpoint:

```http
POST /api/chat
Content-Type: application/json

{
  "audio_data": "<base64_encoded_audio>",
  "audio_format": "wav"
}
```

The system will:
1. Recognize the speaker
2. Detect attendance intent
3. Return appropriate response message
4. Log attendance automatically

## Configuration

Set the voice samples directory via environment variable:
```bash
VOICE_SAMPLES_DIR=./voice_samples
```

Default: `./voice_samples`

## Implementation Details

### Voice Storage
- Voice samples are stored as files in the `voice_samples` directory
- File naming: `{user_id}_{name}_{timestamp}.{format}`
- Profile metadata stored in BadgerDB

### Speaker Recognition
**Current Implementation:**
- Simplified matching using audio hash comparison
- Suitable for development and testing

**Production Recommendations:**
- Integrate with speaker verification services:
  - Azure Speaker Recognition API
  - AWS Voice ID
  - Google Cloud Speaker Recognition
- Or use open-source libraries:
  - Vosk (offline speech recognition)
  - Speaker verification models (e.g., ECAPA-TDNN)

### Speech-to-Text
**Current Implementation:**
- Placeholder for speech-to-text conversion

**Production Recommendations:**
- Google Cloud Speech-to-Text
- Azure Speech Services
- AWS Transcribe
- Vosk (offline)

## Usage Flow

1. **Registration:**
   - User records voice sample saying their name
   - System stores voice profile
   - User can add multiple samples for better accuracy

2. **Attendance Logging:**
   - User speaks: "I'm here" or "Punch in"
   - System recognizes voice
   - System detects attendance intent
   - System responds: "Punched in" or "Gotcha!"
   - Attendance logged automatically

3. **Unrecognized User:**
   - User speaks but not recognized
   - System responds: "Sorry, you're not in our school."

## Frontend Integration

The frontend should:
1. Use Web Speech API or MediaRecorder API to capture audio
2. Convert audio to base64
3. Send to `/api/chat` or `/api/voice/recognize` endpoint
4. Display the response message

Example JavaScript:
```javascript
// Capture audio using MediaRecorder
const mediaRecorder = new MediaRecorder(stream);
const chunks = [];

mediaRecorder.ondataavailable = (e) => chunks.push(e.data);
mediaRecorder.onstop = async () => {
  const blob = new Blob(chunks, { type: 'audio/wav' });
  const base64 = await blobToBase64(blob);
  
  // Send to API
  fetch('/api/voice/recognize', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      audio_data: base64,
      audio_format: 'wav'
    })
  });
};
```

## Security Considerations

- Voice samples contain biometric data - ensure secure storage
- Consider encryption for stored voice samples
- Implement access controls for voice profile management
- Comply with privacy regulations (GDPR, etc.)

## Future Enhancements

- Real-time voice recognition
- Multi-language support
- Voice command expansion (beyond attendance)
- Integration with attendance management systems
- Voice-based authentication
- Batch voice registration
- Voice quality validation

