/**
 * Chat Interface System
 * Handles WebSocket communication, message management, and user interactions
 */

// Chat constants
const CHAT_CONFIG = {
    RECONNECT_BASE_DELAY: 10000, // Base delay for reconnection
    RECONNECT_MAX_DELAY: 60000,  // Maximum reconnection delay
    RECONNECT_JITTER_MAX: 5000,  // Maximum jitter to add
    STATUS_CHECK_INTERVAL: 30000,
    DEFAULT_INPUT_BEHAVIOR: 'enter_to_send'
};

const MESSAGE_TYPES = {
    AI_PROMPT: 'ai_prompt',
    AI_RESPONSE: 'ai_response',
    AI_RESPONSE_END: 'ai_response_end',
    SESSION_STATUS: 'session_status',
    ERROR: 'error'
};

/**
 * WebSocket connection manager
 */
class WebSocketManager {
    constructor(chatId, provider) {
        // Prevent multiple instances for the same chat using a global registry
        const instanceKey = `${chatId}_${provider}`;
        
        // Use window object to store instances globally to survive page reloads within same session
        if (!window._wsManagerInstances) {
            window._wsManagerInstances = {};
        }
        
        console.log('WebSocketManager constructor called for:', instanceKey);
        console.log('Existing WebSocket instances:', Object.keys(window._wsManagerInstances));
        
        if (window._wsManagerInstances[instanceKey]) {
            const existingInstance = window._wsManagerInstances[instanceKey];
            console.log('Reusing existing WebSocket instance for chat:', chatId, 'key:', instanceKey);
            console.log('Existing instance connection state:', existingInstance.ws ? existingInstance.ws.readyState : 'no socket');
            
            // Clean up old connection if it's not in good state
            if (existingInstance.ws && existingInstance.ws.readyState === WebSocket.CLOSED) {
                console.log('Cleaning up closed WebSocket connection');
                delete window._wsManagerInstances[instanceKey];
            } else {
                return existingInstance;
            }
        }
        
        this.chatId = chatId;
        this.provider = provider;
        this.ws = null;
        this.connected = false;
        this.reconnectTimeout = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.eventHandlers = new Map();
        this.instanceKey = instanceKey;
        
        // Store instance globally
        window._wsManagerInstances[instanceKey] = this;
        
        console.log('Created new WebSocket instance for chat:', chatId, 'key:', instanceKey);
        console.log('Total instances now:', Object.keys(window._wsManagerInstances).length);
    }

    /**
     * Establish WebSocket connection
     */
    connect() {
        // Don't create new connection if already connecting or connected
        if (this.ws && (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)) {
            console.log('WebSocket already connecting/connected, skipping new connection');
            return;
        }
        
        // Clean up existing connection if it's in an error state
        if (this.ws && this.ws.readyState === WebSocket.CLOSING) {
            console.log('Waiting for existing WebSocket to close...');
            // Wait for the connection to fully close before creating a new one
            setTimeout(() => this.connect(), 100);
            return;
        }
        
        console.log('Creating new WebSocket connection');
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        this.ws = new WebSocket(wsUrl);
        this.setupEventHandlers();
    }

    /**
     * Setup WebSocket event handlers
     */
    setupEventHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.connected = true;
            this.reconnectAttempts = 0; // Reset counter on successful connection
            this.emit('connected');
            
            // Send session status immediately since onopen guarantees connection is ready
            this.sendSessionStatus();
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.emit('message', message);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.connected = false;
            this.emit('disconnected');
            
            // Only attempt reconnection if under the limit
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
                this.scheduleReconnect();
            } else {
                console.warn('Max reconnection attempts reached, stopping reconnection');
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.emit('error', error);
        };
    }

    /**
     * Send session status to server
     */
    sendSessionStatus() {
        // Only send if WebSocket is in OPEN state
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.warn('WebSocket not open, cannot send session status. State:', this.ws ? this.ws.readyState : 'no socket');
            return false;
        }
        
        const success = this.send({
            type: MESSAGE_TYPES.SESSION_STATUS,
            data: {
                chat_id: this.chatId,
                provider: this.provider
            }
        });
        
        if (!success) {
            console.warn('Failed to send session status, will retry on next connection');
            // Retry after a short delay if the send failed
            setTimeout(() => {
                this.sendSessionStatus();
            }, 1000);
        }
        
        return success;
    }

    /**
     * Send message through WebSocket
     */
    send(message) {
        if (this.connected && this.ws && this.ws.readyState === WebSocket.OPEN) {
            try {
                this.ws.send(JSON.stringify(message));
                return true;
            } catch (error) {
                console.error('Failed to send WebSocket message:', error);
                return false;
            }
        }
        console.warn('WebSocket not ready for sending. State:', this.ws ? this.ws.readyState : 'no socket');
        return false;
    }

    /**
     * Schedule reconnection attempt with exponential backoff and jitter
     */
    scheduleReconnect() {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
        }
        
        this.reconnectAttempts++;
        
        // Calculate exponential backoff delay
        const backoffDelay = Math.min(
            CHAT_CONFIG.RECONNECT_BASE_DELAY * Math.pow(2, this.reconnectAttempts - 1),
            CHAT_CONFIG.RECONNECT_MAX_DELAY
        );
        
        // Add jitter to prevent thundering herd
        const jitter = Math.random() * CHAT_CONFIG.RECONNECT_JITTER_MAX;
        const totalDelay = backoffDelay + jitter;
        
        console.log(`Scheduling reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${Math.round(totalDelay)}ms (base: ${backoffDelay}ms, jitter: ${Math.round(jitter)}ms)`);
        
        this.reconnectTimeout = setTimeout(() => {
            this.connect();
        }, totalDelay);
    }

    /**
     * Reset reconnection attempts and retry
     */
    resetReconnection() {
        this.reconnectAttempts = 0;
        console.log('Reconnection attempts reset');
        if (!this.connected) {
            this.connect();
        }
    }

    /**
     * Event system for communication with chat interface
     */
    on(event, handler) {
        if (!this.eventHandlers.has(event)) {
            this.eventHandlers.set(event, []);
        }
        this.eventHandlers.get(event).push(handler);
    }

    emit(event, data) {
        const handlers = this.eventHandlers.get(event) || [];
        handlers.forEach(handler => handler(data));
    }

    /**
     * Cleanup WebSocket connection
     */
    destroy() {
        if (this.reconnectTimeout) {
            clearTimeout(this.reconnectTimeout);
        }
        if (this.ws) {
            this.ws.close();
        }
        
        // Remove from global instance cache
        if (this.instanceKey && window._wsManagerInstances) {
            delete window._wsManagerInstances[this.instanceKey];
            console.log('WebSocket manager removed from global cache:', this.instanceKey);
        }
    }
}

/**
 * Chat input behavior manager
 */
class ChatInputManager {
    constructor() {
        this.behavior = CHAT_CONFIG.DEFAULT_INPUT_BEHAVIOR;
        this.loadSettings();
    }

    /**
     * Load chat input settings from API
     */
    async loadSettings() {
        try {
            const response = await fetch('/api/settings');
            if (response.ok) {
                const result = await response.json();
                // Handle new standardized response structure
                const settings = result.data || result;
                this.behavior = settings.chatInputBehavior || CHAT_CONFIG.DEFAULT_INPUT_BEHAVIOR;
                console.log('Chat input behavior loaded:', this.behavior);
            }
        } catch (error) {
            console.error('Failed to load chat settings:', error);
            // Fallback to cookie if API fails
            this.loadFromCookie();
        }
    }

    /**
     * Load settings from cookie as fallback
     */
    loadFromCookie() {
        const cookies = document.cookie.split(';');
        for (let cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'chatInputBehavior') {
                this.behavior = decodeURIComponent(value);
                console.log('Chat input behavior loaded from cookie:', this.behavior);
                break;
            }
        }
    }

    /**
     * Handle keyboard input based on current behavior
     */
    handleKeyDown(event, sendCallback) {
        if (event.key !== 'Enter') return;

        if (this.behavior === 'enter_to_send') {
            if (event.shiftKey) {
                // Allow Shift+Enter for new line
                return true;
            } else {
                // Enter sends message
                event.preventDefault();
                sendCallback();
            }
        } else {
            if (event.ctrlKey || event.metaKey) {
                // Ctrl+Enter sends message
                event.preventDefault();
                sendCallback();
            }
            // Enter creates new line (default behavior)
        }
    }

    /**
     * Get placeholder text based on current behavior
     */
    getPlaceholderText(baseText) {
        if (this.behavior === 'enter_to_send') {
            return `${baseText} (Shift+Enter for new line)`;
        } else {
            return `${baseText} (Ctrl+Enter to send)`;
        }
    }

    /**
     * Get send hint text
     */
    getSendHint() {
        if (this.behavior === 'enter_to_send') {
            return 'Enter to send, Shift+Enter for new line';
        } else {
            return 'Ctrl+Enter to send, Enter for new line';
        }
    }
}

/**
 * Provider status manager
 */
class ProviderStatusManager {
    constructor(provider) {
        this.provider = provider;
        this.status = {};
        this.checkInterval = null;
    }

    /**
     * Start periodic status checking
     */
    startStatusChecking() {
        this.checkStatus();
        this.checkInterval = setInterval(() => {
            this.checkStatus();
        }, CHAT_CONFIG.STATUS_CHECK_INTERVAL);
    }

    /**
     * Check provider status
     */
    async checkStatus() {
        try {
            console.log(`Checking status for provider: ${this.provider}`);
            const response = await fetch(`/api/providers/${this.provider}/status`);
            if (response.ok) {
                const result = await response.json();
                // Handle new standardized response structure
                this.status = result.data || result;
                console.log('Provider status response:', this.status);
                this.emit('statusUpdated', this.status);
            } else {
                console.error('Provider status API error:', response.status, response.statusText);
            }
        } catch (error) {
            console.error('Failed to check provider status:', error);
        }
    }

    /**
     * Stop status checking
     */
    stopStatusChecking() {
        if (this.checkInterval) {
            clearInterval(this.checkInterval);
            this.checkInterval = null;
        }
    }

    /**
     * Simple event emitter for status updates
     */
    emit(event, data) {
        window.dispatchEvent(new CustomEvent(`providerStatus:${event}`, {
            detail: data
        }));
    }
}

/**
 * Main chat interface factory for Alpine.js
 */
window.createChatInterface = function(chatId, provider, initialMessages = []) {
    // Prevent multiple chat interfaces for the same chat using global registry
    const interfaceKey = `chat_${chatId}_${provider}`;
    
    // Use window object to store instances globally
    if (!window._chatInterfaceInstances) {
        window._chatInterfaceInstances = {};
    }
    
    console.log('createChatInterface called for:', interfaceKey);
    console.log('Existing chat interfaces:', Object.keys(window._chatInterfaceInstances));
    
    if (window._chatInterfaceInstances[interfaceKey]) {
        console.log('Reusing existing chat interface for:', interfaceKey);
        return window._chatInterfaceInstances[interfaceKey];
    }
    
    console.log('Creating new chat interface for:', interfaceKey, 'with', initialMessages.length, 'initial messages');
    const wsManager = new WebSocketManager(chatId, provider);
    const inputManager = new ChatInputManager();
    const statusManager = new ProviderStatusManager(provider);

    const interfaceInstance = {
        // State
        chatId: chatId,
        provider: provider,
        messages: [...initialMessages], // Initialize with server-provided messages
        newMessage: '',
        connected: false,
        isTyping: false,
        currentResponse: '',
        providerStatus: {},
        streamTimeout: null,

        // Initialization
        init() {
            this.setupWebSocket();
            this.setupStatusManager();
            this.setupMessageScrolling();
        },

        setupWebSocket() {
            wsManager.on('connected', () => {
                this.connected = true;
            });

            wsManager.on('disconnected', () => {
                this.connected = false;
                this.isTyping = false;
            });

            wsManager.on('message', (message) => {
                this.handleMessage(message);
            });

            wsManager.connect();
        },

        setupStatusManager() {
            window.addEventListener('providerStatus:statusUpdated', (event) => {
                console.log('Status updated event received:', event.detail);
                this.providerStatus = event.detail;
            });
            statusManager.startStatusChecking();
        },

        setupMessageScrolling() {
            this.$watch('messages', () => {
                this.$nextTick(() => {
                    this.$refs.messagesContainer.scrollTop = this.$refs.messagesContainer.scrollHeight;
                });
            });
        },

        // Message handling
        handleMessage(message) {
            switch (message.type) {
                case MESSAGE_TYPES.AI_RESPONSE:
                    this.handleAIResponse(message);
                    break;
                case MESSAGE_TYPES.AI_RESPONSE_END:
                    this.handleCompleteResponse();
                    break;
                case MESSAGE_TYPES.ERROR:
                    this.handleError(message);
                    break;
            }
        },

        handleAIResponse(message) {
            if (message.data.stream) {
                this.handleStreamingResponse(message);
            } else {
                this.handleCompleteResponse();
            }
        },

        handleStreamingResponse(message) {
            console.log('handleStreamingResponse called by interface:', `chat_${this.chatId}_${this.provider}`);
            console.log('Message content length:', message.data.content.length);
            console.log('Current messages count:', this.messages.length);
            
            // Initialize streaming if this is the first chunk
            if (!this.isTyping) {
                this.isTyping = true;
                this.currentResponse = '';
                console.log('Starting new streaming response');
            }
            
            this.currentResponse += message.data.content;
            
            const lastMessage = this.messages[this.messages.length - 1];
            console.log('Last message role:', lastMessage?.role, 'isStreaming:', lastMessage?.isStreaming);
            
            if (lastMessage && lastMessage.role === 'assistant' && lastMessage.isStreaming) {
                lastMessage.content = this.currentResponse;
                console.log('Updated existing streaming message, total length:', this.currentResponse.length);
            } else {
                const newMessage = {
                    id: `stream_${this.chatId}_${Date.now()}`,
                    role: 'assistant',
                    content: this.currentResponse,
                    isStreaming: true
                };
                this.messages.push(newMessage);
                console.log('Created new streaming message:', newMessage.id);
            }

            // Note: Streaming end is now handled by explicit AI_RESPONSE_END message
            // No more timeout-based detection
        },

        handleCompleteResponse() {
            this.isTyping = false;
            if (this.currentResponse) {
                const lastMessage = this.messages[this.messages.length - 1];
                if (lastMessage && lastMessage.isStreaming) {
                    lastMessage.isStreaming = false;
                }
                this.currentResponse = '';
            }
        },

        handleError(message) {
            this.isTyping = false;
            // Show error using unified notification system
            uiUtils.showNotification(`WebSocket Error: ${message.data.content}`, 'error', 8000);
        },

        // User interactions
        handleKeyDown(event) {
            inputManager.handleKeyDown(event, () => this.sendMessage());
        },

        sendMessage() {
            if (!this.connected || !this.newMessage.trim() || this.isTyping) return;
            
            const content = this.newMessage.trim();
            
            console.log('sendMessage called by interface:', `chat_${this.chatId}_${this.provider}`);
            console.log('Current messages count before send:', this.messages.length);
            
            // Add user message to UI
            const userMessage = {
                id: Date.now(),
                role: 'user',
                content: content
            };
            this.messages.push(userMessage);
            console.log('User message added to UI:', userMessage.id);
            
            // Send via WebSocket
            const success = wsManager.send({
                type: MESSAGE_TYPES.AI_PROMPT,
                data: {
                    chat_id: this.chatId,
                    provider: this.provider,
                    content: content,
                    timestamp: new Date().toISOString()
                }
            });
            
            if (success) {
                // Clear input and show typing indicator only if send was successful
                this.newMessage = '';
                this.isTyping = true;
                this.currentResponse = '';
                console.log('Message sent successfully via WebSocket');
            } else {
                // Remove the message from UI if send failed
                this.messages.pop();
                console.log('Message send failed, removed from UI');
                uiUtils.showNotification('Failed to send message. Please check your connection.', 'error');
            }
        },

        // UI helpers
        getPlaceholderText() {
            return inputManager.getPlaceholderText('Type your message...');
        },

        getSendHint() {
            return inputManager.getSendHint();
        },

        checkProviderStatus() {
            console.log('Manual provider status check triggered');
            statusManager.checkStatus();
        },

        // Cleanup
        destroy() {
            wsManager.destroy();
            statusManager.stopStatusChecking();
            if (this.streamTimeout) {
                clearTimeout(this.streamTimeout);
            }
            
            // Remove from global instance cache to prevent memory leak
            const interfaceKey = `chat_${this.chatId}_${this.provider}`;
            if (window._chatInterfaceInstances) {
                delete window._chatInterfaceInstances[interfaceKey];
                console.log('Chat interface cleaned up from global cache:', interfaceKey);
            }
        }
    };
    
    // Cache the interface instance globally
    window._chatInterfaceInstances[interfaceKey] = interfaceInstance;
    console.log('Chat interface cached globally:', interfaceKey);
    
    return interfaceInstance;
};