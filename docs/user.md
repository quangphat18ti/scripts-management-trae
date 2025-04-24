# User Management

## Overview
This implementation provides:

- Root account creation from environment variables
- Role-based access control (RBAC) with Root, Admin, and Member roles
- JWT-based authentication middleware
- Role-based authorization middleware
- User management APIs with proper permission checks
- Password management with proper authorization
The API endpoints will be:

- POST /api/users - Create new user (Root and Admin only)
- DELETE /api/users/:id - Delete user (Root and Admin only)
- PUT /api/users/:id/password - Change user password (Root and Admin only)

## User Repository

This repository implementation includes all the necessary functions for the UserService:

- Create : Creates a new user with timestamps
- FindByUsername : Finds a user by their username
- FindByID : Finds a user by their ObjectID
- Delete : Deletes a user by their ID
- UpdatePassword : Updates a user's password and updates the timestamp
- List : Lists all users (useful for admin functionality)
Each function includes proper error handling and timestamp management where appropriate. The repository uses MongoDB's native operations and returns appropriate errors when documents are not found or operations fail.
