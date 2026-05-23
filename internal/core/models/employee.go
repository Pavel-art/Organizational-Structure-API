package models

import "time"

type Employee struct {
	ID           int        `gorm:"primaryKey"`
	DepartmentID int        `gorm:"not null;index"`
	FullName     string     `gorm:"type:varchar(200);not null"`
	Position     string     `gorm:"type:varchar(200);not null"`
	HiredAt      *time.Time `gorm:"type:date"`
	CreatedAt    time.Time

	Department Department `gorm:"foreignKey:DepartmentID"`
}
