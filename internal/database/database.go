package database

import (
	"log"

	"attendancemgmt/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Center{},
		&models.Department{},
		&models.User{},
		&models.Sewadar{},
		&models.Attendance{},
	)
}

func Seed(db *gorm.DB) {
	// Seed center
	var center models.Center
	if db.First(&center).Error != nil {
		center = models.Center{Name: "Main Center", Location: "Headquarters"}
		db.Create(&center)
		log.Println("Seeded: Main Center")
	}

	// Seed departments
	deptNames := []string{"Seva Department", "Langar Department", "Medical Department"}
	depts := make([]models.Department, 0, len(deptNames))
	for _, name := range deptNames {
		var dept models.Department
		if db.Where("name = ?", name).First(&dept).Error != nil {
			dept = models.Department{CenterID: center.ID, Name: name}
			db.Create(&dept)
			log.Printf("Seeded department: %s\n", name)
		}
		depts = append(depts, dept)
	}

	// Seed admin user
	var adminUser models.User
	if db.Where("username = ?", "admin").First(&adminUser).Error != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser = models.User{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         models.RoleCenterAdmin,
		}
		db.Create(&adminUser)
		log.Println("Seeded: admin user (admin/admin123)")
	}

	// Seed operator user
	var opUser models.User
	if db.Where("username = ?", "operator1").First(&opUser).Error != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("operator123"), bcrypt.DefaultCost)
		deptID := depts[0].ID
		opUser = models.User{
			Username:     "operator1",
			PasswordHash: string(hash),
			Role:         models.RoleOperator,
			DepartmentID: &deptID,
		}
		db.Create(&opUser)
		log.Println("Seeded: operator1 user (operator1/operator123)")
	}
}
