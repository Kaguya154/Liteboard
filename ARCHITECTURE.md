# Architecture Diagram

## System Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Frontend (Browser)                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐    │
│  │   Dashboard     │  │   Board View    │  │  Share Join     │    │
│  │                 │  │                 │  │     Page        │    │
│  │  - Avatar       │  │  - Permissions  │  │  - Token Val.   │    │
│  │  - Projects     │  │  - Share Links  │  │  - Auto Join    │    │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘    │
│           │                    │                     │              │
│           └────────────────────┼─────────────────────┘              │
│                                │                                     │
│                         API Client (api.js)                         │
│                    - auth, permissions, share                       │
└────────────────────────────────┼───────────────────────────────────┘
                                 │ HTTPS
┌────────────────────────────────┼───────────────────────────────────┐
│                      Backend (Go/Hertz)                             │
├────────────────────────────────┼───────────────────────────────────┤
│                                │                                     │
│  ┌─────────────────────────────▼──────────────────────────────┐    │
│  │                    API Routes                               │    │
│  │  ┌───────────────┐  ┌───────────────┐  ┌──────────────┐   │    │
│  │  │  Permission   │  │  Share Link   │  │  User        │   │    │
│  │  │  Management   │  │  Management   │  │  Profile     │   │    │
│  │  └───────┬───────┘  └───────┬───────┘  └──────┬───────┘   │    │
│  └──────────┼──────────────────┼──────────────────┼───────────┘    │
│             │                  │                  │                 │
│  ┌──────────▼──────────────────▼──────────────────▼───────────┐    │
│  │              Authentication Middleware                      │    │
│  │         - Session validation                                │    │
│  │         - Permission checks                                 │    │
│  └──────────┬──────────────────┬──────────────────┬───────────┘    │
│             │                  │                  │                 │
│  ┌──────────▼──────┐  ┌────────▼────────┐  ┌─────▼──────────┐     │
│  │   Permission    │  │  Share Token    │  │   User         │     │
│  │   Logic         │  │  Logic          │  │   Logic        │     │
│  │  (internal)     │  │  (api/share.go) │  │  (auth)        │     │
│  └──────────┬──────┘  └────────┬────────┘  └─────┬──────────┘     │
│             │                  │                  │                 │
│  ┌──────────▼──────────────────▼──────────────────▼───────────┐    │
│  │                  Database Layer (SQLite)                    │    │
│  │  ┌────────────┐  ┌──────────────┐  ┌──────────────────┐   │    │
│  │  │   user     │  │share_token   │  │detail_permission │   │    │
│  │  │            │  │              │  │                  │   │    │
│  │  │ +avatar_url│  │ +token       │  │ user_id          │   │    │
│  │  │            │  │ +project_id  │  │ content_ids[]    │   │    │
│  │  │            │  │ +perm_level  │  │ action           │   │    │
│  │  │            │  │ +expires_at  │  │                  │   │    │
│  │  └────────────┘  └──────────────┘  └──────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘

                              ┌─────────────┐
                              │   GitHub    │
                              │   OAuth     │
                              │  (Avatar)   │
                              └─────────────┘
```

## Data Flow Examples

### 1. Generating a Share Link

```
User (Browser)                Backend                    Database
     │                           │                           │
     │  Click "Share"            │                           │
     ├──────────────────────────>│                           │
     │                           │                           │
     │  POST /api/projects/5/    │                           │
     │  share                    │                           │
     │  {perm:"read", hours:24}  │                           │
     ├──────────────────────────>│                           │
     │                           │                           │
     │                           │ Check: User has admin     │
     │                           │ permission on project 5   │
     │                           ├──────────────────────────>│
     │                           │<──────────────────────────┤
     │                           │ ✓ Authorized              │
     │                           │                           │
     │                           │ Generate random token     │
     │                           │ (32 bytes, base64)        │
     │                           │                           │
     │                           │ INSERT share_token        │
     │                           ├──────────────────────────>│
     │                           │<──────────────────────────┤
     │                           │ Token ID: 42              │
     │                           │                           │
     │  200 OK                   │                           │
     │  {token:"abc...",         │                           │
     │   project_id:5,           │                           │
     │   expires_at:...}         │                           │
     │<──────────────────────────┤                           │
     │                           │                           │
     │  Display link with        │                           │
     │  copy button              │                           │
     │                           │                           │
```

### 2. Joining via Share Link

```
User (Browser)              Backend                    Database
     │                         │                           │
     │  Visit /share?token=xyz │                           │
     ├────────────────────────>│                           │
     │                         │                           │
     │  Check if logged in     │                           │
     ├────────────────────────>│                           │
     │<────────────────────────┤                           │
     │  ✓ Session valid        │                           │
     │                         │                           │
     │  POST /api/share/xyz/   │                           │
     │  join                   │                           │
     ├────────────────────────>│                           │
     │                         │                           │
     │                         │ SELECT share_token        │
     │                         │ WHERE token='xyz'         │
     │                         ├─────────────────────────>│
     │                         │<─────────────────────────┤
     │                         │ Found: project_id=5,      │
     │                         │        perm="read"        │
     │                         │                           │
     │                         │ Check expires_at          │
     │                         │ ✓ Not expired             │
     │                         │                           │
     │                         │ Add user to project       │
     │                         │ INSERT/UPDATE             │
     │                         │ detail_permission         │
     │                         ├─────────────────────────>│
     │                         │<─────────────────────────┤
     │                         │ ✓ Permission added        │
     │                         │                           │
     │  200 OK                 │                           │
     │  {project_id:5}         │                           │
     │<────────────────────────┤                           │
     │                         │                           │
     │  Show success,          │                           │
     │  link to project        │                           │
     │                         │                           │
```

### 3. Adding User Permission

```
Owner (Browser)             Backend                    Database
     │                         │                           │
     │  Click "Manage Access"  │                           │
     │  Enter user ID: 42      │                           │
     │  Select: "write"        │                           │
     │  Click "Add"            │                           │
     ├────────────────────────>│                           │
     │                         │                           │
     │  POST /api/projects/5/  │                           │
     │  permissions            │                           │
     │  {user_id:42,           │                           │
     │   perm_level:"write"}   │                           │
     ├────────────────────────>│                           │
     │                         │                           │
     │                         │ Check: User has admin     │
     │                         │ on project 5              │
     │                         ├─────────────────────────>│
     │                         │<─────────────────────────┤
     │                         │ ✓ Authorized              │
     │                         │                           │
     │                         │ Find existing permission  │
     │                         │ for user 42 on project    │
     │                         ├─────────────────────────>│
     │                         │<─────────────────────────┤
     │                         │ No existing write perm    │
     │                         │                           │
     │                         │ INSERT detail_permission  │
     │                         │ OR UPDATE content_ids[]   │
     │                         ├─────────────────────────>│
     │                         │<─────────────────────────┤
     │                         │ ✓ Permission granted      │
     │                         │                           │
     │  200 OK                 │                           │
     │  {message:"added"}      │                           │
     │<────────────────────────┤                           │
     │                         │                           │
     │  Refresh permission     │                           │
     │  list, show user 42     │                           │
     │                         │                           │
```

## Component Interactions

### Permission Check Flow

```
┌────────────────────────────────────────────────────────────────┐
│                    HTTP Request                                │
│                  (with session cookie)                         │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            ▼
                 ┌──────────────────────┐
                 │  Session Middleware   │
                 │  - Verify session     │
                 │  - Load user object   │
                 └──────────┬────────────┘
                            │
                            ▼
            ┌───────────────────────────────┐
            │  Permission Check Middleware  │
            │  - Extract content ID         │
            │  - Check required action      │
            │  - Call HasPermission()       │
            └───────────┬───────────────────┘
                        │
                        ▼
        ┌───────────────────────────────────────┐
        │  HasPermission() (internal package)   │
        │  1. Query detail_permission table     │
        │  2. Check if content_id in list       │
        │  3. Compare permission levels         │
        │  4. Return boolean                    │
        └───────────┬───────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        ▼                       ▼
   ✓ Allowed              ✗ Denied
        │                       │
        ▼                       ▼
  Execute Handler        403 Forbidden
```

### Avatar Loading Flow

```
┌─────────────────────────────────────────────────────────┐
│           Page Load (Dashboard)                         │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
              ┌─────────────────────┐
              │  Dashboard.init()   │
              │  - Load profile     │
              │  - Render avatar    │
              └──────────┬──────────┘
                         │
                         ▼
              ┌─────────────────────────┐
              │  GET /api/user/profile  │
              └──────────┬──────────────┘
                         │
                         ▼
              ┌─────────────────────────┐
              │  Backend: GetUserProfile│
              │  - Get from session     │
              │  - Return user object   │
              └──────────┬──────────────┘
                         │
                         ▼
              ┌────────────────────────────┐
              │  Response:                 │
              │  {id, username, email,     │
              │   avatar_url, groups}      │
              └──────────┬─────────────────┘
                         │
                         ▼
              ┌─────────────────────────┐
              │  renderUserAvatar()     │
              │  - Create img element   │
              │  - Or fallback div      │
              │  - Add tooltip          │
              │  - Attach to header     │
              └─────────────────────────┘
```

## Database Schema Relationships

```
┌──────────────┐         ┌──────────────────┐         ┌──────────────┐
│    user      │         │detail_permission │         │   project    │
├──────────────┤         ├──────────────────┤         ├──────────────┤
│ id (PK)      │◄────────┤ user_id (FK)     │         │ id (PK)      │
│ username     │         │ content_type     │────────►│ name         │
│ email        │         │ content_ids[]    │         │ description  │
│ avatar_url   │         │ action           │         │ creator_id   │
│ ...          │         └──────────────────┘         └──────────────┘
└──────────────┘                  │
                                  │
                                  │ References content_ids
                                  │ as JSON array
                                  │
                         ┌────────▼─────────┐
                         │  share_token     │
                         ├──────────────────┤
                         │ id (PK)          │
                         │ token (UNIQUE)   │
                         │ project_id (FK)  ├─────► project.id
                         │ permission_level │
                         │ created_at       │
                         │ expires_at       │
                         └──────────────────┘

Permission Storage Format:
┌──────────────────────────────────────────────────────────┐
│ detail_permission table                                  │
├──────────────────────────────────────────────────────────┤
│ user_id | content_type | content_ids    | action        │
├──────────────────────────────────────────────────────────┤
│ 1       | project      | [1, 3, 5]      | admin         │
│ 1       | project      | [2, 4]         | write         │
│ 2       | project      | [1]            | read          │
│ 3       | project      | [1, 2, 3]      | write         │
└──────────────────────────────────────────────────────────┘

Interpretation:
- User 1 has admin on projects 1,3,5 and write on 2,4
- User 2 has read on project 1
- User 3 has write on projects 1,2,3
```

## Security Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                     Security Checkpoints                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. HTTPS Transport (Production)                                │
│     └─> Encrypt all data in transit                            │
│                                                                 │
│  2. Session Authentication                                      │
│     └─> Verify user is logged in                               │
│                                                                 │
│  3. Permission Authorization                                    │
│     └─> Check user has required permission level               │
│                                                                 │
│  4. Token Validation (Share Links)                              │
│     └─> Verify token exists and not expired                    │
│                                                                 │
│  5. Input Validation                                            │
│     └─> Sanitize and validate all inputs                       │
│                                                                 │
│  6. XSS Prevention                                              │
│     └─> HTML escape all user-generated content                 │
│                                                                 │
│  7. SQL Injection Prevention                                    │
│     └─> Use parameterized queries via dbhelper                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## File Organization

```
liteboard/
├── main.go                      # Entry point, routes, DB setup
│
├── internal/                    # Core business logic
│   ├── model.go                 # Data models (User, ShareToken, etc.)
│   ├── crud.go                  # Database operations
│   ├── permission.go            # Permission checking logic
│   └── database.go              # DB connection
│
├── auth/                        # Authentication & session
│   ├── auth.go                  # Session middleware, User models
│   └── openid.go                # GitHub OAuth, avatar capture
│
├── api/                         # HTTP handlers
│   ├── project.go               # Project CRUD
│   ├── content.go               # Content CRUD
│   ├── permission.go            # Permission CRUD
│   ├── share.go                 # ★ Permission & share management
│   └── auth.go                  # ★ User profile endpoint
│
└── frontend/                    # Client-side code
    ├── dashboard_new.html       # ★ With avatar
    ├── board.html               # ★ With permission/share modals
    ├── share.html               # ★ Share link join page
    │
    ├── css/
    │   └── styles.css           # ★ Avatar, permission, share styles
    │
    └── js/
        ├── api.js               # ★ Permission/share API methods
        ├── dashboard.js         # ★ Avatar loading/rendering
        └── board.js             # ★ Permission/share management

★ = Modified or new for this feature
```

This architecture provides a clear separation of concerns, with frontend handling UI/UX, backend managing business logic and security, and the database storing persistent data.
