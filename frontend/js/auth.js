/**
 * Authentication handling for Liteboard
 * Manages login state and redirects
 */

const Auth = {
    /**
     * Check if user is logged in by attempting to fetch projects
     * This is a simple check since we're using session-based auth
     */
    async checkAuth() {
        try {
            const response = await fetch('/api/projects', {
                credentials: 'include',
            });
            
            if (response.status === 401) {
                return false;
            }
            
            return response.ok;
        } catch (error) {
            return false;
        }
    },

    /**
     * Redirect to login if not authenticated
     */
    async requireAuth() {
        const isAuthenticated = await this.checkAuth();
        if (!isAuthenticated) {
            window.location.href = '/';
        }
        return isAuthenticated;
    },

    /**
     * Logout user
     */
    async logout() {
        try {
            await fetch('/auth/logout', {
                method: 'POST',
                credentials: 'include',
            });
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            window.location.href = '/';
        }
    },

    /**
     * Initialize auth on protected pages
     */
    init() {
        // Check if on a protected page (dashboard or board)
        const protectedPages = ['/dashboard', '/board.html'];
        const currentPath = window.location.pathname;
        
        if (protectedPages.some(page => currentPath.includes(page))) {
            this.requireAuth();
        }
    },
};

// Auto-initialize auth checking on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => Auth.init());
} else {
    Auth.init();
}
