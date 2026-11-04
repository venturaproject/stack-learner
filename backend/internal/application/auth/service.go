// Package auth contains the application services for authentication and authorization.
// Services are split by responsibility:
//   - JWTService     — token generation and validation
//   - AuthService    — login, logout, refresh, profile
//   - UserService    — user CRUD and management
//   - RoleService    — role CRUD
//   - PermissionService — permission CRUD
package auth
