/**
 * API Client for Liteboard
 * Handles all HTTP requests to the backend API
 */

const API = {
    // Base configuration
    baseURL: '',
    
    /**
     * Generic fetch wrapper with error handling
     */
    async request(url, options = {}) {
        const config = {
            credentials: 'include', // Include cookies for session-based auth
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        };

        try {
            const response = await fetch(url, config);
            
            // Handle authentication errors
            if (response.status === 401) {
                window.location.href = '/';
                throw new Error('Unauthorized - Please login');
            }

            // Parse JSON response
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || `HTTP ${response.status}: ${response.statusText}`);
            }

            return data;
        } catch (error) {
            console.error('API Request Error:', error);
            throw error;
        }
    },

    /**
     * Projects API
     */
    projects: {
        async getAll() {
            return API.request('/api/projects');
        },

        async getById(id) {
            return API.request(`/api/projects/${id}`);
        },

        async create(projectData) {
            return API.request('/api/projects', {
                method: 'POST',
                body: JSON.stringify(projectData),
            });
        },

        async update(id, projectData) {
            return API.request(`/api/projects/${id}`, {
                method: 'PUT',
                body: JSON.stringify(projectData),
            });
        },

        async delete(id) {
            return API.request(`/api/projects/${id}`, {
                method: 'DELETE',
            });
        },
    },

    /**
     * Content Lists API (Trello columns)
     */
    lists: {
        async getByProject(projectId) {
            return API.request(`/api/content_lists?projectid=${projectId}`);
        },

        async getById(id) {
            return API.request(`/api/content_lists/${id}`);
        },

        async create(listData) {
            return API.request('/api/content_lists', {
                method: 'POST',
                body: JSON.stringify(listData),
            });
        },

        async update(id, listData) {
            return API.request(`/api/content_lists/${id}`, {
                method: 'PUT',
                body: JSON.stringify(listData),
            });
        },

        async delete(id) {
            return API.request(`/api/content_lists/${id}`, {
                method: 'DELETE',
            });
        },
    },

    /**
     * Content Entries API (Trello cards)
     */
    entries: {
        async getAll() {
            return API.request('/api/content_entries');
        },

        async getById(id) {
            return API.request(`/api/content_entries/${id}`);
        },

        async create(entryData) {
            return API.request('/api/content_entries', {
                method: 'POST',
                body: JSON.stringify(entryData),
            });
        },

        async update(id, entryData) {
            return API.request(`/api/content_entries/${id}`, {
                method: 'PUT',
                body: JSON.stringify(entryData),
            });
        },

        async delete(id) {
            return API.request(`/api/content_entries/${id}`, {
                method: 'DELETE',
            });
        },
    },

    /**
     * Authentication API
     */
    auth: {
        async logout() {
            try {
                await API.request('/auth/logout', {
                    method: 'POST',
                });
                window.location.href = '/';
            } catch (error) {
                // Still redirect on error
                window.location.href = '/';
            }
        },
    },
};
