/**
 * Dashboard functionality for managing projects
 */

const Dashboard = {
    // State
    projects: [],
    
    // DOM elements
    elements: {
        projectsContainer: null,
        modal: null,
        form: null,
        newProjectBtn: null,
        closeModalBtn: null,
        cancelBtn: null,
    },

    /**
     * Initialize dashboard
     */
    async init() {
        this.cacheDOMElements();
        this.attachEventListeners();
        await this.loadUserProfile();
        await this.loadProjects();
    },

    /**
     * Load and display user profile
     */
    async loadUserProfile() {
        try {
            const user = await API.auth.getProfile();
            this.renderUserAvatar(user);
        } catch (error) {
            console.error('Failed to load user profile:', error);
        }
    },

    /**
     * Render user avatar in header
     */
    renderUserAvatar(user) {
        const container = document.getElementById('user-avatar-container');
        if (!container) return;

        const avatarHTML = user.avatar_url 
            ? `<img src="${user.avatar_url}" alt="${user.username}" class="user-avatar" />`
            : `<div class="user-avatar-fallback">${user.username.charAt(0).toUpperCase()}</div>`;

        const tooltipHTML = `
            <div class="user-tooltip" id="user-tooltip">
                <div class="user-tooltip-name">${this.escapeHtml(user.username)}</div>
                <div class="user-tooltip-id">ID: ${user.id}</div>
            </div>
        `;

        container.innerHTML = avatarHTML + tooltipHTML;

        // Add hover event for tooltip
        const avatarElement = container.querySelector('.user-avatar, .user-avatar-fallback');
        const tooltip = document.getElementById('user-tooltip');
        
        if (avatarElement && tooltip) {
            let hideTimeout;
            
            avatarElement.addEventListener('mouseenter', () => {
                clearTimeout(hideTimeout);
                tooltip.classList.add('show');
            });

            avatarElement.addEventListener('mouseleave', () => {
                hideTimeout = setTimeout(() => {
                    tooltip.classList.remove('show');
                }, 200);
            });

            tooltip.addEventListener('mouseenter', () => {
                clearTimeout(hideTimeout);
            });

            tooltip.addEventListener('mouseleave', () => {
                hideTimeout = setTimeout(() => {
                    tooltip.classList.remove('show');
                }, 200);
            });
        }
    },

    /**
     * Cache DOM elements
     */
    cacheDOMElements() {
        this.elements.projectsContainer = document.getElementById('projects-container');
        this.elements.modal = document.getElementById('project-modal');
        this.elements.form = document.getElementById('project-form');
        this.elements.newProjectBtn = document.getElementById('new-project-btn');
        this.elements.closeModalBtn = document.getElementById('close-modal');
        this.elements.cancelBtn = document.getElementById('cancel-btn');
    },

    /**
     * Attach event listeners
     */
    attachEventListeners() {
        // Open modal
        this.elements.newProjectBtn.addEventListener('click', () => this.openModal());
        
        // Close modal
        this.elements.closeModalBtn.addEventListener('click', () => this.closeModal());
        this.elements.cancelBtn.addEventListener('click', () => this.closeModal());
        
        // Close modal on background click
        this.elements.modal.addEventListener('click', (e) => {
            if (e.target === this.elements.modal) {
                this.closeModal();
            }
        });

        // Form submission
        this.elements.form.addEventListener('submit', (e) => this.handleFormSubmit(e));
    },

    /**
     * Load all projects
     */
    async loadProjects() {
        try {
            this.showLoading();
            this.projects = await API.projects.getAll();
            this.renderProjects();
        } catch (error) {
            this.showError('Failed to load projects: ' + error.message);
        }
    },

    /**
     * Render projects in the grid
     */
    renderProjects() {
        if (this.projects.length === 0) {
            this.elements.projectsContainer.innerHTML = `
                <div class="empty-state" style="grid-column: 1 / -1;">
                    <div class="empty-state-icon">📋</div>
                    <div class="empty-state-text">No projects yet</div>
                    <div class="empty-state-subtext">Create your first project to get started</div>
                </div>
            `;
            return;
        }

        this.elements.projectsContainer.innerHTML = this.projects.map(project => `
            <div class="project-card" data-project-id="${project.id}">
                <div class="project-card-header">
                    <h3 class="project-card-title">${this.escapeHtml(project.name)}</h3>
                </div>
                <p class="project-card-description">${this.escapeHtml(project.description || 'No description')}</p>
                <div class="project-card-actions">
                    <button class="btn btn-sm btn-delete" data-project-id="${project.id}" onclick="Dashboard.deleteProject(${project.id}, event)">Delete</button>
                </div>
            </div>
        `).join('');

        // Add click handlers to project cards
        document.querySelectorAll('.project-card').forEach(card => {
            card.addEventListener('click', (e) => {
                // Don't navigate if clicking on delete button
                if (!e.target.classList.contains('btn-delete')) {
                    const projectId = card.dataset.projectId;
                    this.openProject(projectId);
                }
            });
        });
    },

    /**
     * Open project board
     */
    openProject(projectId) {
        window.location.href = `/board.html?project=${projectId}`;
    },

    /**
     * Show loading state
     */
    showLoading() {
        this.elements.projectsContainer.innerHTML = '<div class="loading-spinner">Loading projects...</div>';
    },

    /**
     * Show error message
     */
    showError(message) {
        this.elements.projectsContainer.innerHTML = `
            <div class="empty-state" style="grid-column: 1 / -1;">
                <div class="empty-state-icon">⚠️</div>
                <div class="empty-state-text">Error</div>
                <div class="empty-state-subtext">${this.escapeHtml(message)}</div>
            </div>
        `;
    },

    /**
     * Open modal
     */
    openModal() {
        this.elements.modal.classList.add('active');
        this.elements.form.reset();
        document.getElementById('project-name').focus();
    },

    /**
     * Close modal
     */
    closeModal() {
        this.elements.modal.classList.remove('active');
        this.elements.form.reset();
    },

    /**
     * Handle form submission
     */
    async handleFormSubmit(e) {
        e.preventDefault();
        
        const name = document.getElementById('project-name').value.trim();
        const description = document.getElementById('project-description').value.trim();

        if (!name) {
            alert('Please enter a project name');
            return;
        }

        try {
            // Disable submit button
            const submitBtn = this.elements.form.querySelector('button[type="submit"]');
            submitBtn.disabled = true;
            submitBtn.textContent = 'Creating...';

            await API.projects.create({ name, description });
            
            this.closeModal();
            await this.loadProjects();
        } catch (error) {
            alert('Failed to create project: ' + error.message);
        } finally {
            // Re-enable submit button
            const submitBtn = this.elements.form.querySelector('button[type="submit"]');
            submitBtn.disabled = false;
            submitBtn.textContent = 'Create Project';
        }
    },

    /**
     * Delete project
     */
    async deleteProject(projectId, event) {
        event.stopPropagation();
        
        if (!confirm('Are you sure you want to delete this project? This action cannot be undone.')) {
            return;
        }

        try {
            await API.projects.delete(projectId);
            await this.loadProjects();
        } catch (error) {
            alert('Failed to delete project: ' + error.message);
        }
    },

    /**
     * Escape HTML to prevent XSS
     */
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },
};

// Initialize dashboard when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => Dashboard.init());
} else {
    Dashboard.init();
}
