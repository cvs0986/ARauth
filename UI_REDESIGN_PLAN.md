# UI/UX Redesign Plan

## Overview
Complete redesign of the Admin Dashboard to provide a user-friendly, hierarchical navigation structure with drill-down capabilities.

## Navigation Structure

### For SYSTEM Users
```
System Level (Root)
├── Dashboard (Overview: tenants, system users, roles, permissions, capabilities, audit logs)
├── Tenants
│   └── [Drill down to Tenant] → Tenant Level View
│       ├── Users
│       │   └── [Drill down to User] → User Level View
│       ├── Roles
│       ├── Permissions
│       ├── Capabilities
│       ├── Features
│       ├── Settings
│       └── Audit Logs
├── System Users
│   └── [Drill down to User] → User Level View
├── System Roles
├── System Permissions
├── System Capabilities
├── Settings (System-level settings)
└── Audit Logs (System-level audit logs)
```

### For TENANT Users
```
Tenant Level (Root)
├── Dashboard (Overview: users, roles, permissions, features, audit logs)
├── Users
│   └── [Drill down to User] → User Level View
├── Roles
├── Permissions
├── Features
├── Capabilities
├── Settings (Tenant-level settings)
└── Audit Logs (Tenant-level audit logs)
```

## Key Features

### 1. Breadcrumb Navigation
- Show current location in hierarchy
- Allow quick navigation to parent levels
- Example: `System > Tenants > ACME Corp > Users > John Doe`

### 2. Context Switching
- SYSTEM users can switch between system view and tenant views
- Tenant selector in header for SYSTEM users
- Clear indication of current context

### 3. Dashboard Improvements

#### System Dashboard
- **Overview Cards:**
  - Total Tenants
  - Total System Users
  - Total System Roles
  - Active Capabilities
  - Recent Audit Events
- **Quick Actions:**
  - Create Tenant
  - Create System User
  - Create System Role
  - View System Settings
- **Recent Activity:**
  - Latest tenant creations
  - Latest user logins
  - System capability changes
  - Audit log entries

#### Tenant Dashboard
- **Overview Cards:**
  - Total Users
  - Total Roles
  - Enabled Features
  - Active Capabilities
  - Recent Audit Events
- **Quick Actions:**
  - Create User
  - Create Role
  - Enable Feature
  - View Settings
- **Recent Activity:**
  - Latest user creations
  - Feature enablements
  - Capability assignments
  - Audit log entries

### 4. List Views with Drill-Down

#### Tenant List (System Level)
- Table/Card view of all tenants
- Click on tenant → Navigate to tenant detail view
- Actions: View, Edit, Suspend, Resume, Delete
- Filters: Status, Created Date, etc.

#### User List (System/Tenant Level)
- Table/Card view of users
- Click on user → Navigate to user detail view
- Actions: View, Edit, Delete, Enable/Disable MFA
- Filters: Role, Status, Created Date, etc.

#### Role List
- Table/Card view of roles
- Click on role → Navigate to role detail view
- Actions: View, Edit, Delete, Assign Permissions
- Filters: Type, Created Date, etc.

### 5. Detail Views

#### Tenant Detail View
- **Overview Tab:**
  - Tenant information
  - Status
  - Created/Updated dates
  - Statistics (users, roles, etc.)
- **Users Tab:**
  - List of tenant users
  - Create user button
  - User management actions
- **Roles Tab:**
  - List of tenant roles
  - Create role button
  - Role management actions
- **Capabilities Tab:**
  - Assigned capabilities
  - Enable/disable capabilities
- **Features Tab:**
  - Enabled features
  - Enable/disable features
- **Settings Tab:**
  - Tenant settings
  - MFA requirements
  - Other configurations
- **Audit Logs Tab:**
  - Tenant-specific audit logs

#### User Detail View
- **Overview Tab:**
  - User information
  - Status
  - MFA status
  - Created/Updated dates
- **Roles Tab:**
  - Assigned roles
  - Assign/remove roles
- **Permissions Tab:**
  - Effective permissions
  - Permission breakdown
- **Capabilities Tab:**
  - User capability enrollment
  - Capability status
- **Activity Tab:**
  - User-specific audit logs
  - Login history

### 6. Sidebar Navigation

#### System Sidebar
```
🏠 Dashboard
🏢 Tenants
   └── [Tenant List]
👥 System Users
   └── [User List]
🔑 System Roles
   └── [Role List]
🛡️ System Permissions
   └── [Permission List]
🛠️ System Capabilities
   └── [Capability List]
⚙️ Settings
📋 Audit Logs
```

#### Tenant Sidebar
```
🏠 Dashboard
👥 Users
   └── [User List]
🔑 Roles
   └── [Role List]
🛡️ Permissions
   └── [Permission List]
✨ Features
   └── [Feature List]
🔧 Capabilities
   └── [Capability List]
⚙️ Settings
📋 Audit Logs
```

### 7. Implementation Steps

1. **Update Routing Structure**
   - Add nested routes for drill-down navigation
   - `/system/tenants/:tenantId/*` for tenant context
   - `/system/tenants/:tenantId/users/:userId/*` for user context
   - `/tenant/users/:userId/*` for tenant user context

2. **Create Context Providers**
   - `TenantContext` - Current tenant being viewed
   - `UserContext` - Current user being viewed
   - Manage state for breadcrumbs and navigation

3. **Update Sidebar Component**
   - Dynamic navigation based on current context
   - Show/hide items based on user type and context
   - Highlight current location

4. **Create Breadcrumb Component**
   - Display current navigation path
   - Clickable links to parent levels
   - Context-aware (System/Tenant/User)

5. **Redesign Dashboard Pages**
   - System Dashboard
   - Tenant Dashboard
   - User Dashboard (if needed)

6. **Update List Components**
   - Add drill-down functionality
   - Improve filtering and search
   - Better visual hierarchy

7. **Create Detail View Components**
   - Tenant Detail View
   - User Detail View
   - Role Detail View
   - Tab-based navigation within detail views

8. **Update Settings Pages**
   - System Settings
   - Tenant Settings
   - User Settings (if needed)

## Technical Considerations

### State Management
- Use React Context for navigation state
- Zustand store for current context (tenant, user)
- URL params for deep linking

### Routing
- Use React Router nested routes
- URL structure: `/system/tenants/:tenantId/users/:userId`
- Preserve query params for filters

### Performance
- Lazy load detail views
- Pagination for large lists
- Optimistic updates for actions

### Accessibility
- Keyboard navigation
- ARIA labels
- Screen reader support

## UI/UX Improvements

1. **Visual Hierarchy**
   - Clear distinction between levels
   - Consistent spacing and typography
   - Color coding for different contexts

2. **User Feedback**
   - Loading states
   - Success/error messages
   - Confirmation dialogs for destructive actions

3. **Responsive Design**
   - Mobile-friendly navigation
   - Collapsible sidebar
   - Responsive tables/cards

4. **Search and Filtering**
   - Global search
   - Context-aware filters
   - Saved filter presets


