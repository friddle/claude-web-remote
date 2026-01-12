import { CapacitorHttp } from '@capacitor/core';
import { LocalNotifications } from '@capacitor/local-notifications';

const APP_KEY_SESSIONS = 'clauded_sessions';

class App {
    constructor() {
        this.sessions = JSON.parse(localStorage.getItem(APP_KEY_SESSIONS) || '[]');
        this.currentSession = null;
    }

    init() {
        this.renderSessionList();
        this.setupEventListeners();

        // Request notification permission on load
        this.requestNotificationPermission();
    }

    setupEventListeners() {
        // Modal toggles
        const btnAdd = document.getElementById('btnAddSession');
        if (btnAdd) {
            btnAdd.addEventListener('click', () => this.showModal('modalAdd'));
        }

        const btnCancel = document.getElementById('btnCancelAdd');
        if (btnCancel) {
            btnCancel.addEventListener('click', () => this.hideModal('modalAdd'));
        }

        // Add Session Form
        const formAdd = document.getElementById('formAddSession');
        if (formAdd) {
            formAdd.addEventListener('submit', (e) => {
                e.preventDefault();
                this.addSession();
            });
        }

        // Back button in session view
        const btnBack = document.getElementById('btnBack');
        if (btnBack) {
            btnBack.addEventListener('click', () => this.closeSession());
        }

        // Tab Bar
        document.querySelectorAll('.tab-item').forEach(item => {
            item.addEventListener('click', () => {
                const tab = item.dataset.tab;
                this.switchTab(tab);
            });
        });
    }

    async requestNotificationPermission() {
        try {
            await LocalNotifications.requestPermissions();
        } catch (e) {
            console.error('Notification permission error:', e);
        }
    }

    // --- Session Management ---

    addSession() {
        const name = document.getElementById('inputName').value;
        const host = document.getElementById('inputHost').value;
        const sessionId = document.getElementById('inputSession').value;
        const password = document.getElementById('inputPassword').value;

        const newSession = {
            id: Date.now().toString(),
            name,
            host,
            sessionId,
            password,
            createdAt: new Date().toISOString()
        };

        this.sessions.push(newSession);
        this.saveSessions();
        this.renderSessionList();
        this.hideModal('modalAdd');

        // Clear form
        document.getElementById('formAddSession').reset();
        document.getElementById('inputHost').value = 'clauded.friddle.me'; // Restore default
    }

    saveSessions() {
        localStorage.setItem(APP_KEY_SESSIONS, JSON.stringify(this.sessions));
    }

    deleteSession(id, event) {
        if (event) event.stopPropagation();
        if (!confirm('Are you sure you want to delete this session?')) return;

        this.sessions = this.sessions.filter(s => s.id !== id);
        this.saveSessions();
        this.renderSessionList();
    }

    renderSessionList() {
        const list = document.getElementById('sessionList');
        if (!list) return;

        list.innerHTML = '';

        if (this.sessions.length === 0) {
            list.innerHTML = `
                <div style="text-align: center; padding: 40px; color: #999;">
                    <div style="font-size: 40px; margin-bottom: 10px;">💬</div>
                    <p>No sessions yet.</p>
                    <p>Click + to add one.</p>
                </div>
            `;
            return;
        }

        this.sessions.forEach(session => {
            const item = document.createElement('div');
            item.className = 'session-item';
            item.innerHTML = `
                <div class="session-icon">
                    <span>>_</span>
                </div>
                <div class="session-info">
                    <div class="session-name">${this.escapeHtml(session.name)}</div>
                    <div class="session-detail">${this.escapeHtml(session.host)} / ${this.escapeHtml(session.sessionId)}</div>
                </div>
            `;

            // Long press to delete could be better, but let's just click to open
            item.addEventListener('click', () => this.openSession(session));

            list.appendChild(item);
        });
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // --- View Management ---

    openSession(session) {
        this.currentSession = session;

        const title = document.getElementById('sessionTitle');
        if (title) title.textContent = session.name;

        const frame = document.getElementById('terminalFrame');
        if (frame) {
            // Construct URL: https://session:password@host/session
            // Or if host is just domain: https://domain/session
            // The logic from previous app.js: `https://${session}:${password}@${host}/${session}`

            // Handle host having protocol or not
            let host = session.host;
            let protocol = 'https://';
            if (host.startsWith('http://') || host.startsWith('https://')) {
                // simple parsing
                if (host.startsWith('http://')) {
                    protocol = 'http://';
                    host = host.substring(7);
                } else {
                    host = host.substring(8);
                }
            }

            // Construct auth URL
            const url = `${protocol}${session.sessionId}:${session.password}@${host}/${session.sessionId}`;
            console.log('Opening session URL:', url);
            frame.src = url;
        }

        this.showView('view-session');
    }

    closeSession() {
        this.currentSession = null;
        const frame = document.getElementById('terminalFrame');
        if (frame) {
            frame.src = 'about:blank';
        }
        this.showView('view-home');
    }

    showView(viewId) {
        document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
        const view = document.getElementById(viewId);
        if (view) view.classList.add('active');
    }

    showModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) modal.classList.add('active');
    }

    hideModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) modal.classList.remove('active');
    }

    switchTab(tab) {
        document.querySelectorAll('.tab-item').forEach(t => {
            if (t.dataset.tab === tab) t.classList.add('active');
            else t.classList.remove('active');
        });

        if (tab === 'home') {
            // Logic for home tab
        } else if (tab === 'settings') {
            // Logic for settings tab
            alert('Settings not implemented yet');
        }
    }

    // --- Tmux Controls ---

    sendTmux(action) {
        if (!this.currentSession) return;

        console.log('Sending Tmux Action:', action);

        // NOTE: This is a simulation/placeholder.
        // Sending actual keys to the iframe requires the iframe to be on the same origin
        // or the server to expose a specific API for input injection.
        // Assuming Gotty running on the server might capture these if we use a specific API.

        // If we were using a native terminal plugin, we would write to the PTY here.

        // For now, we'll try to use a hypothetical API endpoint if it existed,
        // or just show feedback to the user.

        const feedback = document.createElement('div');
        feedback.style.position = 'fixed';
        feedback.style.top = '50%';
        feedback.style.left = '50%';
        feedback.style.transform = 'translate(-50%, -50%)';
        feedback.style.background = 'rgba(0,0,0,0.7)';
        feedback.style.color = 'white';
        feedback.style.padding = '10px 20px';
        feedback.style.borderRadius = '8px';
        feedback.style.zIndex = '2000';
        feedback.textContent = `Tmux: ${action}`;
        document.body.appendChild(feedback);

        setTimeout(() => feedback.remove(), 1000);
    }
}

// Initialize
const app = new App();
window.app = app; // Expose for inline handlers
document.addEventListener('DOMContentLoaded', () => app.init());
