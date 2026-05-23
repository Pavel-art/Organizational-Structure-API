package models

import "time"

type Department struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(200);not null" json:"name"`
	ParentID  *int      `gorm:"index" json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`

	Parent    *Department  `gorm:"foreignKey:ParentID" json:"-"`
	Children  []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID" json:"employees,omitempty"`
}
