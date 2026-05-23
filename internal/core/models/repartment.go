package models

import "time"

type Department struct {
	ID        int    `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(200);not null;uniqueIndex:idx_department_parent_name"`
	ParentID  *int   `gorm:"uniqueIndex:idx_department_parent_name"`
	CreatedAt time.Time

	Parent    *Department  `gorm:"foreignKey:ParentID"`
	Children  []Department `gorm:"foreignKey:ParentID"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID"`
}
