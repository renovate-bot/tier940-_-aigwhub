/**
 * Utility Functions
 * Common functionality used across the application
 */

/**
 * API utilities
 */
window.apiUtils = {
    /**
     * Perform fetch with error handling
     */
    async request(url, options = {}) {
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
            },
        };

        const mergedOptions = { ...defaultOptions, ...options };

        try {
            const response = await fetch(url, mergedOptions);
            
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    },

    /**
     * GET request
     */
    async get(url) {
        return this.request(url);
    },

    /**
     * POST request
     */
    async post(url, data) {
        return this.request(url, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    },

    /**
     * PUT request
     */
    async put(url, data) {
        return this.request(url, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    },

    /**
     * DELETE request
     */
    async delete(url) {
        return this.request(url, {
            method: 'DELETE',
        });
    }
};

/**
 * UI utilities
 */
window.uiUtils = {
    /**
     * Show unified banner notification
     */
    showNotification(message, type = 'info', duration = 5000) {
        // Create notification container if it doesn't exist
        let container = document.getElementById('notification-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'notification-container';
            container.className = 'fixed top-4 right-4 z-50 space-y-2';
            document.body.appendChild(container);
        }

        // Create notification element
        const notification = document.createElement('div');
        notification.className = `p-4 rounded-lg shadow-lg border transition-all duration-300 max-w-md ${this.getNotificationClasses(type)}`;
        
        // Add icon and message
        const icon = this.getNotificationIcon(type);
        notification.innerHTML = `
            <div class="flex items-start space-x-3">
                <div class="flex-shrink-0">
                    ${icon}
                </div>
                <div class="flex-1">
                    <p class="text-sm font-medium">${message}</p>
                </div>
                <button class="flex-shrink-0 ml-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300" onclick="this.parentElement.parentElement.remove()">
                    <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                    </svg>
                </button>
            </div>
        `;

        // Add to container with animation
        notification.style.transform = 'translateX(100%)';
        notification.style.opacity = '0';
        container.appendChild(notification);

        // Animate in
        setTimeout(() => {
            notification.style.transform = 'translateX(0)';
            notification.style.opacity = '1';
        }, 10);

        // Auto remove after duration
        if (duration > 0) {
            setTimeout(() => {
                this.removeNotification(notification);
            }, duration);
        }

        return notification;
    },

    /**
     * Get CSS classes for notification type
     */
    getNotificationClasses(type) {
        switch (type) {
            case 'success':
                return 'bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-200 border-green-200 dark:border-green-700';
            case 'error':
                return 'bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-200 border-red-200 dark:border-red-700';
            case 'warning':
                return 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-800 dark:text-yellow-200 border-yellow-200 dark:border-yellow-700';
            case 'info':
            default:
                return 'bg-blue-50 dark:bg-blue-900/20 text-blue-800 dark:text-blue-200 border-blue-200 dark:border-blue-700';
        }
    },

    /**
     * Get icon for notification type
     */
    getNotificationIcon(type) {
        switch (type) {
            case 'success':
                return '<svg class="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg>';
            case 'error':
                return '<svg class="w-5 h-5 text-red-400" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"></path></svg>';
            case 'warning':
                return '<svg class="w-5 h-5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path></svg>';
            case 'info':
            default:
                return '<svg class="w-5 h-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path></svg>';
        }
    },

    /**
     * Remove notification with animation
     */
    removeNotification(notification) {
        notification.style.transform = 'translateX(100%)';
        notification.style.opacity = '0';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    },

    /**
     * Show toast notification (legacy support)
     */
    showToast(message, type = 'info', duration = 5000) {
        return this.showNotification(message, type, duration);
    },

    /**
     * Format relative time
     */
    formatRelativeTime(date) {
        const now = new Date();
        const diff = now - new Date(date);
        const seconds = Math.floor(diff / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) {
            return days === 1 ? 'Yesterday' : `${days} days ago`;
        } else if (hours > 0) {
            return `${hours} hour${hours > 1 ? 's' : ''} ago`;
        } else if (minutes > 0) {
            return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
        } else {
            return 'Just now';
        }
    },

    /**
     * Debounce function calls
     */
    debounce(func, delay) {
        let timeoutId;
        return function (...args) {
            clearTimeout(timeoutId);
            timeoutId = setTimeout(() => func.apply(this, args), delay);
        };
    },

    /**
     * Throttle function calls
     */
    throttle(func, limit) {
        let inThrottle;
        return function (...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }
};

/**
 * Storage utilities
 */
window.storageUtils = {
    /**
     * Set item in localStorage with error handling
     */
    setItem(key, value) {
        try {
            localStorage.setItem(key, JSON.stringify(value));
            return true;
        } catch (error) {
            console.error('Failed to set localStorage item:', error);
            return false;
        }
    },

    /**
     * Get item from localStorage with error handling
     */
    getItem(key, defaultValue = null) {
        try {
            const item = localStorage.getItem(key);
            return item ? JSON.parse(item) : defaultValue;
        } catch (error) {
            console.error('Failed to get localStorage item:', error);
            return defaultValue;
        }
    },

    /**
     * Remove item from localStorage
     */
    removeItem(key) {
        try {
            localStorage.removeItem(key);
            return true;
        } catch (error) {
            console.error('Failed to remove localStorage item:', error);
            return false;
        }
    },

    /**
     * Clear all localStorage
     */
    clear() {
        try {
            localStorage.clear();
            return true;
        } catch (error) {
            console.error('Failed to clear localStorage:', error);
            return false;
        }
    }
};

/**
 * Form utilities
 */
window.formUtils = {
    /**
     * Serialize form data to object
     */
    serializeForm(form) {
        const formData = new FormData(form);
        const obj = {};
        
        for (const [key, value] of formData.entries()) {
            if (obj[key]) {
                // Handle multiple values
                if (!Array.isArray(obj[key])) {
                    obj[key] = [obj[key]];
                }
                obj[key].push(value);
            } else {
                obj[key] = value;
            }
        }
        
        return obj;
    },

    /**
     * Validate email format
     */
    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    },

    /**
     * Validate URL format
     */
    isValidUrl(url) {
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    }
};

/**
 * Error handling utilities
 */
window.errorUtils = {
    // Track recent errors to prevent duplicates
    recentErrors: new Set(),
    /**
     * Handle and display errors consistently
     */
    handleError(error, context = 'Unknown') {
        const errorMessage = error?.message || 'An unexpected error occurred';
        
        // Use original console.error to avoid infinite loop
        originalConsoleError(`[${context}] Error:`, error);
        
        // Send error to server logs
        this.logToServer(error, context, 'error');
        
        // Show user-friendly error message
        window.uiUtils.showNotification(
            `Error in ${context}: ${errorMessage}`,
            'error'
        );
    },

    /**
     * Create standardized error object
     */
    createError(message, code = null, details = null) {
        const error = new Error(message);
        if (code) error.code = code;
        if (details) error.details = details;
        return error;
    },

    /**
     * Log client-side errors to server
     */
    async logToServer(error, context, level = 'error') {
        try {
            const errorMessage = `${context}: ${error?.message || 'Unknown error'}`;
            const errorKey = `${errorMessage}:${window.location.href}`;
            
            // Prevent duplicate errors within 5 seconds
            if (this.recentErrors.has(errorKey)) {
                return;
            }
            
            this.recentErrors.add(errorKey);
            setTimeout(() => {
                this.recentErrors.delete(errorKey);
            }, 5000);

            const errorData = {
                message: errorMessage,
                stack: error?.stack || 'No stack trace',
                url: window.location.href,
                userAgent: navigator.userAgent,
                level: level
            };

            await fetch('/api/logs/client', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(errorData)
            });
        } catch (logError) {
            originalConsoleError('Failed to log error to server:', logError);
        }
    }
};

/**
 * Performance utilities
 */
window.performanceUtils = {
    /**
     * Measure execution time
     */
    measure(name, fn) {
        const start = performance.now();
        const result = fn();
        const end = performance.now();
        
        console.log(`${name} took ${end - start} milliseconds`);
        return result;
    },

    /**
     * Measure async execution time
     */
    async measureAsync(name, fn) {
        const start = performance.now();
        const result = await fn();
        const end = performance.now();
        
        console.log(`${name} took ${end - start} milliseconds`);
        return result;
    }
};

// Global error handler
window.addEventListener('error', (event) => {
    // Provide more detailed error information for debugging
    const errorDetails = {
        message: event.message || 'Unknown error',
        filename: event.filename || 'Unknown file',
        lineno: event.lineno || 'Unknown line',
        colno: event.colno || 'Unknown column',
        error: event.error
    };
    
    let errorInfo;
    if (event.error === null || event.error === undefined) {
        errorInfo = {
            message: `Script error: ${errorDetails.message} at ${errorDetails.filename}:${errorDetails.lineno}:${errorDetails.colno}`,
            stack: `Error: ${errorDetails.message}\n    at ${errorDetails.filename}:${errorDetails.lineno}:${errorDetails.colno}`
        };
        if (window.errorUtils && typeof errorUtils.handleError === 'function') {
            window.errorUtils.handleError(errorInfo, 'Global Script Error');
        }
    } else {
        errorInfo = {
            message: `${errorDetails.message} at ${errorDetails.filename}:${errorDetails.lineno}:${errorDetails.colno}`,
            stack: event.error.stack || `Error: ${errorDetails.message}\n    at ${errorDetails.filename}:${errorDetails.lineno}:${errorDetails.colno}`
        };
        if (window.errorUtils && typeof errorUtils.handleError === 'function') {
            window.errorUtils.handleError(errorInfo, 'Global');
        }
    }
});

window.addEventListener('unhandledrejection', (event) => {
    if (window.errorUtils && typeof errorUtils.handleError === 'function') {
        const reason = event.reason || 'Unknown promise rejection';
        window.errorUtils.handleError(reason, 'Promise Rejection');
    }
});

// Capture original console.error for later use
const originalConsoleError = console.error;