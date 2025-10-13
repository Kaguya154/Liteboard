/**
 * Board functionality for Trello-like interface
 * Handles lists and cards with drag-and-drop
 */

const Board = {
    // State
    projectId: null,
    project: null,
    lists: [],
    draggedCard: null,
    draggedFrom: null,
    
    // DOM elements
    elements: {
        boardContainer: null,
        projectTitle: null,
        addListBtn: null,
        listModal: null,
        listForm: null,
        cardModal: null,
        cardForm: null,
    },

    /**
     * Initialize board
     */
    async init() {
        this.getProjectIdFromURL();
        this.cacheDOMElements();
        this.attachEventListeners();
        await this.loadBoard();
    },

    /**
     * Get project ID from URL parameter
     */
    getProjectIdFromURL() {
        const urlParams = new URLSearchParams(window.location.search);
        this.projectId = urlParams.get('project');
        
        if (!this.projectId) {
            alert('No project ID specified');
            window.location.href = '/dashboard';
        }
    },

    /**
     * Cache DOM elements
     */
    cacheDOMElements() {
        this.elements.boardContainer = document.getElementById('board-container');
        this.elements.projectTitle = document.getElementById('project-title');
        this.elements.addListBtn = document.getElementById('add-list-btn');
        this.elements.listModal = document.getElementById('list-modal');
        this.elements.listForm = document.getElementById('list-form');
        this.elements.cardModal = document.getElementById('card-modal');
        this.elements.cardForm = document.getElementById('card-form');
    },

    /**
     * Attach event listeners
     */
    attachEventListeners() {
        // Add list button
        this.elements.addListBtn.addEventListener('click', () => this.openListModal());
        
        // List modal controls
        document.getElementById('close-list-modal').addEventListener('click', () => this.closeListModal());
        document.getElementById('cancel-list-btn').addEventListener('click', () => this.closeListModal());
        this.elements.listModal.addEventListener('click', (e) => {
            if (e.target === this.elements.listModal) this.closeListModal();
        });
        
        // Card modal controls
        document.getElementById('close-card-modal').addEventListener('click', () => this.closeCardModal());
        document.getElementById('cancel-card-btn').addEventListener('click', () => this.closeCardModal());
        document.getElementById('delete-card-btn').addEventListener('click', () => this.deleteCard());
        this.elements.cardModal.addEventListener('click', (e) => {
            if (e.target === this.elements.cardModal) this.closeCardModal();
        });
        
        // Form submissions
        this.elements.listForm.addEventListener('submit', (e) => this.handleListFormSubmit(e));
        this.elements.cardForm.addEventListener('submit', (e) => this.handleCardFormSubmit(e));
    },

    /**
     * Load board data
     */
    async loadBoard() {
        try {
            this.showLoading();
            
            // Load project details
            this.project = await API.projects.getById(this.projectId);
            this.elements.projectTitle.textContent = this.project.name;
            
            // Load lists
            this.lists = await API.lists.getByProject(this.projectId);
            
            this.renderBoard();
        } catch (error) {
            this.showError('Failed to load board: ' + error.message);
        }
    },

    /**
     * Render the entire board
     */
    renderBoard() {
        if (this.lists.length === 0) {
            this.elements.boardContainer.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📝</div>
                    <div class="empty-state-text">No lists yet</div>
                    <div class="empty-state-subtext">Click "Add List" to create your first list</div>
                </div>
            `;
            return;
        }

        this.elements.boardContainer.innerHTML = this.lists.map(list => this.renderList(list)).join('');
        this.attachListEventListeners();
    },

    /**
     * Render a single list
     */
    renderList(list) {
        const cards = (list.items || [])
            .filter(item => item.Entry) // Only show entries, not nested lists
            .map(item => item.Entry);

        return `
            <div class="board-list" data-list-id="${list.id}">
                <div class="list-header">
                    <input type="text" 
                           class="list-title" 
                           value="${this.escapeHtml(list.title)}" 
                           data-list-id="${list.id}"
                           readonly>
                    <button class="btn-delete-list" data-list-id="${list.id}" title="Delete list">×</button>
                </div>
                <div class="list-cards" data-list-id="${list.id}">
                    ${cards.map(card => this.renderCard(card, list.id)).join('')}
                </div>
                <button class="add-card-btn" data-list-id="${list.id}">+ Add a card</button>
            </div>
        `;
    },

    /**
     * Render a single card
     */
    renderCard(card, listId) {
        return `
            <div class="card" 
                 draggable="true" 
                 data-card-id="${card.id}" 
                 data-list-id="${listId}">
                <div class="card-title">${this.escapeHtml(card.title || card.content)}</div>
                ${card.content && card.title !== card.content ? `<div class="card-content">${this.escapeHtml(card.content)}</div>` : ''}
            </div>
        `;
    },

    /**
     * Attach event listeners to lists and cards
     */
    attachListEventListeners() {
        // List title editing
        document.querySelectorAll('.list-title').forEach(input => {
            input.addEventListener('click', (e) => {
                e.target.removeAttribute('readonly');
                e.target.select();
            });
            
            input.addEventListener('blur', (e) => this.handleListTitleChange(e));
            input.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    e.target.blur();
                }
            });
        });

        // Delete list buttons
        document.querySelectorAll('.btn-delete-list').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const listId = e.target.dataset.listId;
                this.deleteList(listId);
            });
        });

        // Add card buttons
        document.querySelectorAll('.add-card-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const listId = e.target.dataset.listId;
                this.openCardModal(null, listId);
            });
        });

        // Card click to edit
        document.querySelectorAll('.card').forEach(card => {
            card.addEventListener('click', (e) => {
                const cardId = e.currentTarget.dataset.cardId;
                const listId = e.currentTarget.dataset.listId;
                this.openCardModal(cardId, listId);
            });

            // Drag and drop events
            card.addEventListener('dragstart', (e) => this.handleDragStart(e));
            card.addEventListener('dragend', (e) => this.handleDragEnd(e));
        });

        // Drop zones
        document.querySelectorAll('.list-cards').forEach(zone => {
            zone.addEventListener('dragover', (e) => this.handleDragOver(e));
            zone.addEventListener('drop', (e) => this.handleDrop(e));
            zone.addEventListener('dragleave', (e) => this.handleDragLeave(e));
        });
    },

    /**
     * Show loading state
     */
    showLoading() {
        this.elements.boardContainer.innerHTML = '<div class="loading-spinner">Loading board...</div>';
    },

    /**
     * Show error message
     */
    showError(message) {
        this.elements.boardContainer.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">⚠️</div>
                <div class="empty-state-text">Error</div>
                <div class="empty-state-subtext">${this.escapeHtml(message)}</div>
            </div>
        `;
    },

    /**
     * List modal controls
     */
    openListModal() {
        this.elements.listModal.classList.add('active');
        this.elements.listForm.reset();
        document.getElementById('list-title').focus();
    },

    closeListModal() {
        this.elements.listModal.classList.remove('active');
        this.elements.listForm.reset();
    },

    /**
     * Handle list form submission
     */
    async handleListFormSubmit(e) {
        e.preventDefault();
        
        const title = document.getElementById('list-title').value.trim();
        if (!title) return;

        try {
            const listData = {
                type: 'list',
                title: title,
                items: [],
                project_id: parseInt(this.projectId)
            };

            await API.lists.create(listData);
            this.closeListModal();
            await this.loadBoard();
        } catch (error) {
            alert('Failed to create list: ' + error.message);
        }
    },

    /**
     * Handle list title change
     */
    async handleListTitleChange(e) {
        const input = e.target;
        const listId = input.dataset.listId;
        const newTitle = input.value.trim();
        
        input.setAttribute('readonly', 'true');

        if (!newTitle) {
            // Revert to original value
            await this.loadBoard();
            return;
        }

        try {
            const list = this.lists.find(l => l.id == listId);
            if (!list) return;

            list.title = newTitle;
            await API.lists.update(listId, list);
        } catch (error) {
            alert('Failed to update list title: ' + error.message);
            await this.loadBoard();
        }
    },

    /**
     * Delete list
     */
    async deleteList(listId) {
        if (!confirm('Are you sure you want to delete this list? All cards in it will be lost.')) {
            return;
        }

        try {
            await API.lists.delete(listId);
            await this.loadBoard();
        } catch (error) {
            alert('Failed to delete list: ' + error.message);
        }
    },

    /**
     * Card modal controls
     */
    openCardModal(cardId, listId) {
        if (cardId) {
            // Edit existing card
            const list = this.lists.find(l => l.id == listId);
            const card = list.items.find(item => item.Entry && item.Entry.id == cardId)?.Entry;
            
            if (card) {
                document.getElementById('card-id').value = card.id;
                document.getElementById('card-list-id').value = listId;
                document.getElementById('card-title').value = card.title || card.content || '';
                document.getElementById('card-content').value = card.content || '';
                document.getElementById('card-modal-title').textContent = 'Edit Card';
                document.getElementById('delete-card-btn').style.display = 'block';
            }
        } else {
            // New card
            this.elements.cardForm.reset();
            document.getElementById('card-list-id').value = listId;
            document.getElementById('card-modal-title').textContent = 'Add Card';
            document.getElementById('delete-card-btn').style.display = 'none';
        }
        
        this.elements.cardModal.classList.add('active');
        document.getElementById('card-title').focus();
    },

    closeCardModal() {
        this.elements.cardModal.classList.remove('active');
        this.elements.cardForm.reset();
    },

    /**
     * Handle card form submission
     */
    async handleCardFormSubmit(e) {
        e.preventDefault();
        
        const cardId = document.getElementById('card-id').value;
        const listId = document.getElementById('card-list-id').value;
        const title = document.getElementById('card-title').value.trim();
        const content = document.getElementById('card-content').value.trim();

        if (!title) {
            alert('Please enter a card title');
            return;
        }

        try {
            const cardData = {
                type: 'task',
                title: title,
                content: content || title,
            };

            if (cardId) {
                // Update existing card
                cardData.id = parseInt(cardId);
                await API.entries.update(cardId, cardData);
            } else {
                // Create new card
                const newCard = await API.entries.create(cardData);
                
                // Add card to list
                const list = this.lists.find(l => l.id == listId);
                if (list) {
                    list.items.push({ Entry: newCard });
                    await API.lists.update(listId, list);
                }
            }

            this.closeCardModal();
            await this.loadBoard();
        } catch (error) {
            alert('Failed to save card: ' + error.message);
        }
    },

    /**
     * Delete card
     */
    async deleteCard() {
        const cardId = document.getElementById('card-id').value;
        const listId = document.getElementById('card-list-id').value;

        if (!confirm('Are you sure you want to delete this card?')) {
            return;
        }

        try {
            // Remove from list
            const list = this.lists.find(l => l.id == listId);
            if (list) {
                list.items = list.items.filter(item => !item.Entry || item.Entry.id != cardId);
                await API.lists.update(listId, list);
            }

            // Delete the entry
            await API.entries.delete(cardId);

            this.closeCardModal();
            await this.loadBoard();
        } catch (error) {
            alert('Failed to delete card: ' + error.message);
        }
    },

    /**
     * Drag and drop handlers
     */
    handleDragStart(e) {
        this.draggedCard = e.target;
        this.draggedFrom = e.target.dataset.listId;
        e.target.classList.add('dragging');
        e.dataTransfer.effectAllowed = 'move';
        e.dataTransfer.setData('text/html', e.target.innerHTML);
    },

    handleDragEnd(e) {
        e.target.classList.remove('dragging');
    },

    handleDragOver(e) {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'move';
        e.currentTarget.classList.add('drag-over');
    },

    handleDragLeave(e) {
        e.currentTarget.classList.remove('drag-over');
    },

    async handleDrop(e) {
        e.preventDefault();
        e.stopPropagation();
        
        const dropZone = e.currentTarget;
        dropZone.classList.remove('drag-over');
        
        const targetListId = dropZone.dataset.listId;
        const cardId = this.draggedCard.dataset.cardId;
        const sourceListId = this.draggedFrom;

        if (sourceListId === targetListId) {
            // Same list, no need to update
            return;
        }

        try {
            // Find the card
            const sourceList = this.lists.find(l => l.id == sourceListId);
            const cardItem = sourceList.items.find(item => item.Entry && item.Entry.id == cardId);
            
            if (!cardItem) return;

            // Remove from source list
            sourceList.items = sourceList.items.filter(item => !item.Entry || item.Entry.id != cardId);
            
            // Add to target list
            const targetList = this.lists.find(l => l.id == targetListId);
            targetList.items.push(cardItem);

            // Update both lists
            await Promise.all([
                API.lists.update(sourceListId, sourceList),
                API.lists.update(targetListId, targetList),
            ]);

            // Refresh board
            await this.loadBoard();
        } catch (error) {
            alert('Failed to move card: ' + error.message);
            await this.loadBoard();
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

// Initialize board when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => Board.init());
} else {
    Board.init();
}
