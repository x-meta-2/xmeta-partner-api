package database

// Admin models — READ-ONLY references to tables managed by xmeta-admin-api.
// These structs are used by AdminAuth middleware to query the shared DB.
// They are NOT included in AutoMigrate — admin-api owns these tables.

type (
	AdminPermission struct {
		Base
		Name        string `gorm:"column:name;not null" json:"name"`
		Description string `gorm:"column:description" json:"description"`
	}

	AdminGroup struct {
		Base
		Name        string            `gorm:"column:name;not null" json:"name"`
		Permissions []AdminPermission `gorm:"many2many:admin_group_permissions;joinForeignKey:admin_group_id;joinReferences:permission_id" json:"permissions"`
	}

	AdminGroupPermission struct {
		AdminGroupID string `gorm:"column:admin_group_id;primaryKey"`
		PermissionID string `gorm:"column:permission_id;primaryKey"`
	}

	AdminUser struct {
		Base
		Email        string      `gorm:"column:email;not null;uniqueIndex" json:"email"`
		AdminGroupID string      `gorm:"column:admin_group_id;type:text" json:"adminGroupId"`
		AdminGroup   *AdminGroup `gorm:"foreignKey:AdminGroupID" json:"adminGroup"`
		Department   string      `gorm:"column:department" json:"department"`
		Status       string      `gorm:"column:status;not null" json:"status"`
		IsEnabled    bool        `gorm:"column:is_enabled;default:true" json:"isEnabled"`
	}
)
