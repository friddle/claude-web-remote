import { CapacitorHttp } from '@capacitor/core';
import { LocalNotifications } from '@capacitor/local-notifications';
import { BackgroundTask } from '@capacitor/background-task';

class ClaudeDApp {
    constructor() {
        this.connected = false;
        this.sessionId = null;
        this.webhookUrl = null;
        this.notifications = [];
    }

    async init() {
        // Request notification permissions
        await this.requestNotificationPermission();

        // Setup form handler
        document.getElementById('connectForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.connect();
        });

        // Load saved credentials
        this.loadCredentials();
    }

    async requestNotificationPermission() {
        try {
            const result = await LocalNotifications.requestPermissions();
            console.log('Notification permissions:', result);
        } catch (error) {
            console.error('Failed to request notification permission:', error);
        }
    }

    loadCredentials() {
        const host = localStorage.getItem('clauded_host');
        const session = localStorage.getItem('clauded_session');
        const password = localStorage.getItem('clauded_password');

        if (host) document.getElementById('host').value = host;
        if (session) document.getElementById('session').value = session;
        if (password) document.getElementById('password').value = password;
    }

    saveCredentials(host, session, password) {
        localStorage.setItem('clauded_host', host);
        localStorage.setItem('clauded_session', session);
        localStorage.setItem('clauded_password', password);
    }

    async connect() {
        const host = document.getElementById('host').value;
        const session = document.getElementById('session').value;
        const password = document.getElementById('password').value;

        // Save credentials
        this.saveCredentials(host, session, password);

        try {
            // Subscribe to notifications via SSE
            this.connectToNotificationStream(host, session);

            this.sessionId = session;
            this.connected = true;

            this.showStatus('Connected successfully!', 'connected');

            // Open WebView to terminal
            this.openTerminal(host, session, password);

        } catch (error) {
            console.error('Connection failed:', error);
            this.showStatus('Connection failed: ' + error.message, 'error');
        }
    }

    connectToNotificationStream(host, session) {
        // Use SSE for notifications
        const url = `https://${host}/api/v1/notifications/stream?session_id=${session}`;
        console.log('Connecting to SSE:', url);

        if (this.eventSource) {
            this.eventSource.close();
        }

        this.eventSource = new EventSource(url);

        this.eventSource.onopen = () => {
            console.log('SSE connected');
        };

        this.eventSource.onerror = (err) => {
            console.error('SSE error:', err);
            // Optional: Try to reconnect after delay
        };

        // Listen for specific events
        // Assuming the server sends events with type 'message' or custom types
        // The server code sends: c.SSEvent(string(notif.Type), string(data))
        
        // Listen for all standard notification types
        const eventTypes = ['task_completed', 'error', 'progress', 'system_status'];
        
        eventTypes.forEach(type => {
            this.eventSource.addEventListener(type, (e) => {
                try {
                    const data = JSON.parse(e.data);
                    this.handleNotification(data);
                } catch (err) {
                    console.error('Failed to parse notification:', err);
                }
            });
        });
        
        // Also listen for generic messages just in case
        this.eventSource.onmessage = (e) => {
             try {
                const data = JSON.parse(e.data);
                this.handleNotification(data);
            } catch (err) {
                // If it's not JSON, might be a keep-alive or raw string
            }
        };
    }
    
    handleNotification(notif) {
        console.log('Received notification:', notif);
        // Map server notification structure to app structure if needed
        // Server sends: { Type, Title, Message, Data, Timestamp }
        
        this.showNotification(notif.Title || 'Notification', notif.Message || '', notif);
    }

    openTerminal(host, session, password) {
        // Open browser with terminal
        const url = `https://${session}:${password}@${host}/${session}`;
        window.open(url, '_blank');
    }

    showNotification(title, body, data = {}) {
        // Add to notifications list
        this.notifications.unshift({
            title,
            body,
            data,
            time: new Date().toLocaleString()
        });

        this.renderNotifications();

        // Show system notification
        LocalNotifications.schedule({
            notifications: [
                {
                    id: Date.now(),
                    title,
                    body,
                    sound: 'default',
                    schedule: { at: new Date() }
                }
            ]
        });
    }

    renderNotifications() {
        const container = document.getElementById('notifications');

        if (this.notifications.length === 0) {
            container.innerHTML = '<p style="text-align: center; color: #6c757d;">No notifications yet</p>';
            return;
        }

        container.innerHTML = this.notifications.map(notif => `
            <div class="notification-item ${notif.data.type === 'error' ? 'error' : ''}">
                <div class="notification-time">${notif.time}</div>
                <div><strong>${notif.title}</strong></div>
                <div>${notif.body}</div>
            </div>
        `).join('');
    }

    showStatus(message, type) {
        const statusEl = document.getElementById('status');
        statusEl.textContent = message;
        statusEl.className = `status ${type}`;
    }
}

// Initialize app
const app = new ClaudeDApp();
document.addEventListener('DOMContentLoaded', () => app.init());
