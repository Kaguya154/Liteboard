# Permission Management and Sharing Feature Implementation

This document describes the implementation of permission management, sharing functionality, and user avatar features for Liteboard.

## Features Implemented

### 1. Database Schema Updates

#### User Table
- Added `avatar_url` field to store user avatars from OAuth providers

#### Share Token Table (New)
- `id`: Primary key
- `token`: Unique token string for share links
- `project_id`: Project being shared
- `permission_level`: Permission level (read/write)
- `created_at`: Creation timestamp
- `expires_at`: Expiration timestamp

### 2. Backend API Endpoints

#### Permission Management
- `POST /api/projects/:id/permissions` - Add user permission to project
  - Request body: `{ "user_id": int, "permission_level": "read|write|admin" }`
  - Requires: admin permission on project
  
- `GET /api/projects/:id/permissions` - Get all permissions for a project
  - Response: List of users with their permission levels
  - Requires: read permission on project

- `DELETE /api/projects/:id/permissions/:userId` - Remove user permission
  - Requires: admin permission on project

#### Share Token Management
- `POST /api/projects/:id/share` - Generate share link
  - Request body: `{ "permission_level": "read|write", "expires_in_hours": int }`
  - Response: Share token object with URL-safe token
  - Requires: admin permission on project

- `GET /api/projects/:id/shares` - List active share tokens for project
  - Filters out expired tokens automatically
  - Requires: admin permission on project

- `DELETE /api/shares/:id` - Delete a share token
  - Requires: logged in user

- `POST /api/share/:token/join` - Join project via share token
  - Validates token is not expired
  - Automatically adds user to project with specified permission level
  - Requires: logged in user
  - Public endpoint (no existing project permission needed)

#### User Profile
- `GET /api/user/profile` - Get current user profile
  - Returns user info including avatar URL
  - Requires: logged in user

### 3. Frontend Components

#### User Avatar Component
Location: Dashboard header (`dashboard_new.html`)

Features:
- Displays user avatar from OAuth provider (GitHub)
- Falls back to initial letter if no avatar available
- Hover tooltip showing username and user ID
- Auto-loads on dashboard initialization

Implementation in `dashboard.js`:
```javascript
loadUserProfile()  // Fetches user data
renderUserAvatar(user)  // Renders avatar with tooltip
```

#### Permission Management Modal
Location: Board view (`board.html`)

Features:
- Add users by ID with permission level selection (read/write/admin)
- List all users with permissions
- Remove user permissions
- Shows user email and permission level with color coding

Access: Click "👥 Manage Access" button in board header

#### Share Link Modal
Location: Board view (`board.html`)

Features:
- Generate new share links with:
  - Permission level selection (read/write)
  - Expiration time in hours
- List active share links with:
  - Full shareable URL
  - Permission level and expiration time
  - Copy to clipboard button
  - Delete button
- Automatically filters expired tokens

Access: Click "🔗 Share" button in board header

#### Share Join Page
Location: `/share` route (`share.html`)

Features:
- Accepts share token via query parameter (`?token=...`)
- Validates token and checks expiration
- Automatically adds user to project if logged in
- Redirects to login if not authenticated
- Shows success message with link to project
- Error handling for invalid/expired tokens

### 4. OAuth Integration

Updated GitHub OAuth to capture avatar URL:
- Modified `auth/openid.go` to extract `avatar_url` from GitHub API
- Updates user avatar on every login to keep it fresh
- Stores avatar URL in user database record

### 5. Permission System

The implementation follows the existing permission model:
- Uses `detail_permission` table with JSON array of content IDs
- Three permission levels: read < write < admin
- Permission inheritance: admin includes write, write includes read
- Project permissions automatically apply to sub-items

### 6. Security Features

- Share tokens use cryptographically secure random generation (32 bytes, base64 encoded)
- Token expiration validation on every use
- Permission checks on all management endpoints
- XSS protection via HTML escaping in frontend

## UI Design

### Color-Coded Permission Levels
- Read: Blue (`#e3f2fd` background)
- Write: Orange (`#fff3e0` background)  
- Admin: Pink (`#fce4ec` background)

### Responsive Design
- Modals work on all screen sizes
- Touch-friendly buttons
- Clear visual hierarchy

## Usage Examples

### Sharing a Project (Admin)
1. Open a project board
2. Click "🔗 Share" button
3. Select permission level (read/write)
4. Set expiration hours (default: 24)
5. Click "Generate Link"
6. Copy the generated link and share it

### Joining via Share Link (User)
1. Receive share link: `https://liteboard.example.com/share?token=abc123...`
2. Click the link (redirects to login if needed)
3. After login, automatically joins project
4. Click "Go to Project" to access the board

### Managing Permissions (Admin)
1. Open a project board
2. Click "👥 Manage Access"
3. Enter user ID and select permission level
4. Click "Add" to grant access
5. Use "Remove" button to revoke access

## API Response Examples

### Get Project Permissions
```json
[
  {
    "user_id": 1,
    "username": "alice",
    "email": "alice@example.com",
    "permission_level": "admin"
  },
  {
    "user_id": 2,
    "username": "bob",
    "email": "bob@example.com",
    "permission_level": "write"
  }
]
```

### Generate Share Token
```json
{
  "id": 1,
  "token": "Kq8xY2p3N8...(base64)",
  "project_id": 5,
  "permission_level": "read",
  "created_at": 1697385600,
  "expires_at": 1697472000
}
```

### User Profile
```json
{
  "id": 1,
  "username": "alice",
  "email": "alice@example.com",
  "groups": ["user", "admin"],
  "avatar_url": "https://avatars.githubusercontent.com/u/12345?v=4"
}
```

## Files Modified

### Backend
- `main.go` - Added share_token table, share routes, user profile endpoint
- `internal/model.go` - Added ShareToken and ProjectPermission models, avatar_url to User
- `internal/crud.go` - Added ShareToken CRUD functions
- `internal/permission.go` - Exported GetPermissionLevel function
- `auth/auth.go` - Added avatar_url support to User models
- `auth/openid.go` - Updated OAuth to capture avatar URL
- `api/share.go` - New file with all permission and sharing endpoints
- `api/auth.go` - Added GetUserProfile endpoint

### Frontend
- `frontend/dashboard_new.html` - Added user avatar container
- `frontend/board.html` - Added permission and share buttons and modals
- `frontend/share.html` - New page for joining via share links
- `frontend/css/styles.css` - Added styles for avatar, permissions, and share components
- `frontend/js/api.js` - Added API client methods for permissions and sharing
- `frontend/js/dashboard.js` - Added avatar loading and rendering
- `frontend/js/board.js` - Added permission and share management functionality

## Testing

The application was successfully built and the server starts without errors. All routes are properly registered:
- Permission management endpoints: ✓
- Share token endpoints: ✓
- User profile endpoint: ✓
- Share join page: ✓

Manual testing approach:
1. Login with GitHub OAuth
2. Create a project
3. Access permission management modal
4. Generate share links
5. Test share link joining flow

## Future Enhancements

Potential improvements not in scope:
- Email notifications for new permissions
- Share link usage analytics
- Role-based permission templates
- Bulk user import
- Permission audit log
