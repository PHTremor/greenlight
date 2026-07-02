package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

// a slice to hold permission codes for a specific user
type Permissions []string

// Helper method to check if the Permission slice contains a specic permission code
func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}

// the permission model
type PermissionModel struct {
	DB *sql.DB
}

// GetAllForUser returns a slice of all permission codes for a specific user
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
	SELECT permissions.code
	FROM permissions
	INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
	INNER JOIN users ON users_permissions.user_id = users.id
	WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddForUser() adds the provided permissions for a given user
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
	INSERT INTO users_permissions
	SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}

// NOTE: for Scaling Permissions or When your system entities grow,
// - you should pivot away from hardcoding strings in the application layer.
// - 1. Database-Driven Default Roles (Recommended)
// 	- Instead of assigning individual permissions to a user directly,
// 		assign the user a single default "Role" (e.g., Standard User) in the database.
// 		How it works: Create a roles table and a roles_permissions join table.
// 		The Query: Modify your insertion logic to copy permissions associated with a specific role ID.
// 		The Code: Your Go handler only needs to pass a single role identifier, like "user" or "default"
// CODE example:

// func (m PermissionModel) AddForUser(userID int64, role string) error {
// 	query := `
// 	INSERT INTO users_permissions (user_id, permission_id)
// 	SELECT $1, permission_id
// 	FROM roles_permissions
// 	JOIN roles ON roles.id = roles_permissions.role_id
// 	WHERE roles.name = $2`

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	_, err := m.DB.ExecContext(ctx, query, userID, role)
// 	return err
// }
