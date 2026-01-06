# ClaudeD Android Client

Android client for ClaudeD - Access your Claude Code sessions from anywhere via your Android device.

## Features

- 📱 WebView interface to Claude Code terminal
- 🔔 Push notifications for task completion
- 🔐 Secure connection with password authentication
- 🌐 Webhook receiver for real-time notifications
- 💾 Save credentials locally for quick access

## Prerequisites

- Node.js 18+ and npm
- Android Studio (for building APK/AAB)
- Android SDK

## Installation

```bash
npm install
```

## Development

```bash
# Sync web assets to Android
npm run sync

# Open in Android Studio
npm run open
```

## Building APK/AAB

### Build Debug APK

```bash
npm run build-apk
```

The APK will be located at:
```
android/app/build/outputs/apk/debug/app-debug.apk
```

### Build Release AAB (for Play Store)

```bash
npm run build-aab
```

The AAB will be located at:
```
android/app/build/outputs/bundle/release/app-release.aab
```

## Usage

1. **Install the APK** on your Android device

2. **Enter connection details**:
   - Host: e.g., `clauded.friddle.me` or your custom host
   - Session ID: Your ClaudeD session ID
   - Password: Your session password

3. **Connect**:
   - Tap "Connect to Claude Code"
   - The app will subscribe to notifications
   - WebView will open showing the terminal

4. **Notifications**:
   - The app runs a local webhook server on port 8080
   - When Claude Code completes tasks, you'll receive push notifications
   - Tap the notification to open the terminal

## Architecture

```
android_client/
├── www/                    # Web assets
│   ├── index.html         # Main UI
│   └── js/
│       └── app.js         # Application logic
├── android/               # Native Android project
│   └── app/src/main/java/com/friddle/clauded/plugin/
│       └── WebhookReceiverPlugin.java  # Webhook server plugin
├── capacitor.config.json  # Capacitor configuration
└── package.json          # Dependencies
```

## Webhook Receiver Plugin

The app includes a native Android plugin that:
- Starts an HTTP server on `localhost:8080`
- Listens for webhook POST requests
- Forwards notifications to the JavaScript layer
- Triggers system notifications

## Notification Flow

```
Claude Code (Client)
    ↓ (task completion)
ClaudeD Server
    ↓ (webhook)
Android App (WebhookReceiverPlugin)
    ↓ (local notification)
System Notification
    ↓ (user taps)
Opens WebView terminal
```

## Permissions

The app requires the following Android permissions:
- `INTERNET` - Connect to remote server
- `ACCESS_NETWORK_STATE` - Check network connectivity
- `POST_NOTIFICATIONS` - Show push notifications
- `FOREGROUND_SERVICE` - Run webhook receiver in background

## Security Notes

- Credentials are stored locally in `localStorage`
- HTTPS is used for all remote connections
- Session IDs and passwords are never transmitted outside the app
- Webhook server only listens on `localhost`

## Troubleshooting

### Can't receive notifications
- Ensure the app has notification permissions
- Check that the webhook server is running
- Verify the server URL is reachable

### Build errors
- Make sure Android SDK is installed
- Update Android Studio to the latest version
- Run `npm run sync` before building

### WebView won't load
- Check your network connection
- Verify the host and session ID are correct
- Ensure the remote server is running

## Development Workflow

1. Modify files in `www/` directory
2. Run `npm run sync` to copy changes to Android
3. Build and install APK
4. Test on device/emulator

## License

ISC
