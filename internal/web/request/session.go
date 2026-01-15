package request

type PermissionGroup uint8

const (
	GUEST  PermissionGroup = 0
	VIEWER PermissionGroup = 1
	EDITOR PermissionGroup = 2
	ADMIN  PermissionGroup = 3
)

type SessionInfo struct {
	Username        string
	PermissionGroup PermissionGroup
}
