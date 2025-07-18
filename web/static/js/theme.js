/**
 * Theme Management System
 * Provides centralized theme switching and persistence
 */

// Theme constants
const THEMES = {
    LIGHT: 'light',
    DARK: 'dark',
    AUTO: 'auto'
};

const STORAGE_KEYS = {
    THEME: 'theme',
    DARK_MODE: 'darkMode'
};

/**
 * Theme manager for handling light/dark/auto theme switching
 */
class ThemeManager {
    constructor() {
        this.currentTheme = this.getStoredTheme();
        this.darkMode = this.calculateDarkMode();
        this.initializeTheme();
    }

    /**
     * Get theme from localStorage with fallback
     */
    getStoredTheme() {
        return localStorage.getItem(STORAGE_KEYS.THEME) || THEMES.LIGHT;
    }

    /**
     * Calculate dark mode based on current theme
     */
    calculateDarkMode() {
        switch (this.currentTheme) {
            case THEMES.DARK:
                return true;
            case THEMES.LIGHT:
                return false;
            case THEMES.AUTO:
                return window.matchMedia('(prefers-color-scheme: dark)').matches;
            default:
                // Fallback to legacy darkMode setting
                return localStorage.getItem(STORAGE_KEYS.DARK_MODE) === 'true';
        }
    }

    /**
     * Initialize theme on page load
     */
    initializeTheme() {
        this.applyTheme();
        this.listenForSystemThemeChanges();
    }

    /**
     * Apply current theme to document
     */
    applyTheme() {
        const html = document.documentElement;
        if (this.darkMode) {
            html.classList.add('dark');
        } else {
            html.classList.remove('dark');
        }
    }

    /**
     * Get next theme in cycle: light -> dark -> auto -> light
     */
    getNextTheme() {
        switch (this.currentTheme) {
            case THEMES.LIGHT:
                return THEMES.DARK;
            case THEMES.DARK:
                return THEMES.AUTO;
            case THEMES.AUTO:
                return THEMES.LIGHT;
            default:
                return THEMES.LIGHT;
        }
    }

    /**
     * Toggle to next theme
     */
    toggleTheme() {
        const newTheme = this.getNextTheme();
        this.setTheme(newTheme);
    }

    /**
     * Set specific theme
     */
    setTheme(theme) {
        if (!Object.values(THEMES).includes(theme)) {
            console.warn(`Invalid theme: ${theme}`);
            return;
        }

        this.currentTheme = theme;
        this.darkMode = this.calculateDarkMode();
        
        this.persistTheme();
        this.applyTheme();
        this.notifyThemeChange();
    }

    /**
     * Persist theme to storage and cookies
     */
    persistTheme() {
        // Local storage
        localStorage.setItem(STORAGE_KEYS.THEME, this.currentTheme);
        localStorage.setItem(STORAGE_KEYS.DARK_MODE, this.darkMode.toString());
        
        // Cookie for server-side persistence
        const expires = new Date();
        expires.setDate(expires.getDate() + 30); // 30 days
        document.cookie = `theme=${this.currentTheme}; path=/; expires=${expires.toUTCString()}`;
    }

    /**
     * Listen for system theme changes when in auto mode
     */
    listenForSystemThemeChanges() {
        if (window.matchMedia) {
            const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            mediaQuery.addEventListener('change', () => {
                if (this.currentTheme === THEMES.AUTO) {
                    this.darkMode = mediaQuery.matches;
                    this.applyTheme();
                    this.notifyThemeChange();
                }
            });
        }
    }

    /**
     * Notify other components of theme change
     */
    notifyThemeChange() {
        window.dispatchEvent(new CustomEvent('themeChanged', {
            detail: {
                theme: this.currentTheme,
                darkMode: this.darkMode
            }
        }));
    }

    /**
     * Get human-readable theme description
     */
    getThemeDescription() {
        switch (this.currentTheme) {
            case THEMES.LIGHT:
                return 'Light Mode (click for Dark)';
            case THEMES.DARK:
                return 'Dark Mode (click for Auto)';
            case THEMES.AUTO:
                return 'Auto Mode (click for Light)';
            default:
                return 'Light Mode';
        }
    }

    /**
     * Get current theme state for Alpine.js integration
     */
    getAlpineState() {
        return {
            darkMode: this.darkMode,
            currentTheme: this.currentTheme,
            
            // Methods for Alpine.js
            toggleTheme: () => this.toggleTheme(),
            getThemeButtonTitle: () => this.getThemeDescription(),
            getCurrentTheme: () => this.currentTheme,
            
            // Legacy method for backward compatibility
            toggleDarkMode: () => this.toggleTheme()
        };
    }
}

// Global theme manager instance
window.themeManager = new ThemeManager();

// Alpine.js theme data factory
window.createThemeData = function() {
    return {
        darkMode: window.themeManager.darkMode,
        currentTheme: window.themeManager.currentTheme,
        
        // Initialize and listen for theme changes
        init() {
            // Listen for global theme changes
            window.addEventListener('themeChanged', (event) => {
                this.currentTheme = event.detail.theme;
                this.darkMode = event.detail.darkMode;
            });
        },
        
        // Methods for Alpine.js
        toggleTheme() {
            window.themeManager.toggleTheme();
        },
        
        getThemeButtonTitle() {
            return window.themeManager.getThemeDescription();
        },
        
        getCurrentTheme() {
            return this.currentTheme;
        },
        
        // Legacy method for backward compatibility
        toggleDarkMode() {
            this.toggleTheme();
        }
    };
};