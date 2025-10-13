# Liteboard Frontend

A complete Trello-like frontend interface for the Liteboard project management system.

## Structure

```
frontend/
├── index.html              # Login page
├── dashboard_new.html      # Projects dashboard
├── board.html             # Trello-like board view
├── css/
│   └── styles.css         # Main stylesheet
└── js/
    ├── api.js             # API client
    ├── auth.js            # Authentication handling
    ├── dashboard.js       # Dashboard logic
    └── board.js           # Board logic with drag-and-drop
```

## Features

### Login Page (`/`)
- Beautiful gradient design
- GitHub OAuth integration
- Secure authentication flow

### Dashboard (`/dashboard`)
- Grid layout of all user projects
- Create new project with modal dialog
- Click project card to open board
- Delete project functionality
- Empty state when no projects exist

### Board View (`/board.html?project=ID`)
- Horizontal scrolling Trello-like layout
- Create/Delete lists (columns)
- Editable list titles (click to edit)
- Add/Edit/Delete cards
- Drag and drop cards between lists
- Modal dialogs for card editing
- Smooth animations

## API Integration

The frontend integrates with the following backend endpoints:

### Authentication
- `GET /auth/github/login` - Initiate GitHub OAuth
- `POST /auth/logout` - Logout user

### Projects
- `GET /api/projects` - Get all projects
- `POST /api/projects` - Create project
- `GET /api/projects/:id` - Get project details
- `PUT /api/projects/:id` - Update project
- `DELETE /api/projects/:id` - Delete project

### Lists (Columns)
- `GET /api/content_lists?projectid=X` - Get lists for project
- `POST /api/content_lists` - Create list
- `PUT /api/content_lists/:id` - Update list (including items)
- `DELETE /api/content_lists/:id` - Delete list

### Cards (Entries)
- `GET /api/content_entries` - Get all entries
- `POST /api/content_entries` - Create entry
- `PUT /api/content_entries/:id` - Update entry
- `DELETE /api/content_entries/:id` - Delete entry

## Development

### Running Locally

1. Start the Go backend:
   ```bash
   go run main.go -p 8080
   ```

2. Open browser to `http://localhost:8080`

3. You'll need to configure GitHub OAuth:
   - Create a GitHub OAuth App
   - Set `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` in `.env`
   - Set callback URL to `http://localhost:8080/auth/github/callback`

### Testing Without OAuth

To test the dashboard and board views without OAuth, you can:

1. Temporarily disable the `LoginRequired()` middleware in `main.go`
2. Or create a test user in the database and manually set a session cookie

## Technical Details

### Drag and Drop

The board uses HTML5 Drag and Drop API:
- Cards are draggable between lists
- Visual feedback during drag (opacity, borders)
- Updates backend via PUT requests to lists

### State Management

- Simple client-side state in JavaScript objects
- Optimistic UI updates
- Reload from server after mutations

### Error Handling

- 401 errors redirect to login page
- User-friendly error messages via alerts
- Console logging for debugging

### Security

- XSS prevention via HTML escaping
- Session-based authentication
- CSRF protection via SameSite cookies

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- ES6+ JavaScript features used
- No IE11 support

## Future Enhancements

Potential improvements:
- Drag and drop reordering within lists
- Card details modal with more fields
- Real-time updates via WebSocket
- Keyboard shortcuts
- Card search and filtering
- Due dates and labels
- User avatars and assignments
- Activity log
- Markdown support in card descriptions
