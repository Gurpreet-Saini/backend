package models

import (
	"time"

	"gorm.io/gorm"
)

type Center struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;uniqueIndex"`
	Location    string         `json:"location"`
	Departments []Department   `json:"departments,omitempty" gorm:"foreignKey:CenterID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Department struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CenterID    uint           `json:"center_id" gorm:"not null"`
	Center      *Center        `json:"center,omitempty" gorm:"foreignKey:CenterID"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Sewadars    []Sewadar      `json:"sewadars,omitempty" gorm:"foreignKey:DepartmentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Sewadar struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	SewadarCode      string         `json:"sewadar_id" gorm:"column:employee_id;uniqueIndex;not null"`
	Name             string         `json:"name" gorm:"not null"`
	ParentSpouseName string         `json:"parent_spouse_name"`
	Gender           string         `json:"gender"`
	BadgeStatus      string         `json:"badge_status"`
	DepartmentID     *uint          `json:"department_id"`
	Department       *Department    `json:"department,omitempty" gorm:"foreignKey:DepartmentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Phone            string         `json:"phone"`
	Email            string         `json:"email"`
	Attendance       []Attendance   `json:"attendance,omitempty" gorm:"foreignKey:SewadarID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

type Attendance struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	SewadarID    uint       `json:"sewadar_id" gorm:"not null"`
	Sewadar      *Sewadar   `json:"sewadar,omitempty" gorm:"foreignKey:SewadarID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DepartmentID *uint      `json:"department_id"`
	Department   *Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Date         time.Time  `json:"date" gorm:"type:date;not null"`
	CheckIn      time.Time  `json:"check_in" gorm:"not null"`
	CheckOut     *time.Time `json:"check_out"`
	MarkedBy     uint       `json:"marked_by"`
	MarkedByUser *User      `json:"marked_by_user,omitempty" gorm:"foreignKey:MarkedBy"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UserRole string

const (
	RoleCenterAdmin UserRole = "center_admin"
	RoleOperator    UserRole = "operator"
	RoleDeptViewer  UserRole = "dept_viewer"
)

type User struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Username     string     `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"not null"`
	Role         UserRole   `json:"role" gorm:"not null;default:'operator'"`
	DepartmentID *uint      `json:"department_id"` // scoped for operator/viewer
	Department   *Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
