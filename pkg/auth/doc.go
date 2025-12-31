// Package auth implements player authentication and identity management.
//
// Responsibilities:
//   - Player registration and login
//   - Password hashing and verification
//   - JWT issuance and validation
//
// Non-responsibilities:
//   - Authorization beyond identity (roles, permissions)
//   - Game or lobby logic
//   - Transport-specific concerns outside HTTP handlers
package auth
