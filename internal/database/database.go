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
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage for PgBouncer
	}), &gorm.Config{
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
		&models.Feedback{},
		&models.Item{},
		&models.InventoryTransaction{},
	)
}

func Seed(db *gorm.DB, adminUsername, adminPassword string) {
	// Seed center
	var center models.Center
	if db.Unscoped().First(&center).Error != nil {
		center = models.Center{Name: "Main Center", Location: "Headquarters"}
		db.Create(&center)
		log.Println("Seeded: Main Center")
	}

	// Seed departments
	deptNames := []string{"Seva Department", "Langar Department", "Medical Department"}
	depts := make([]models.Department, 0, len(deptNames))
	for _, name := range deptNames {
		var dept models.Department
		if db.Unscoped().Where("name = ?", name).First(&dept).Error != nil {
			dept = models.Department{CenterID: center.ID, Name: name}
			db.Create(&dept)
			log.Printf("Seeded department: %s\n", name)
		}
		depts = append(depts, dept)
	}

	// Seed super admin user
	if adminUsername != "" && adminPassword != "" {
		var superAdminUser models.User
		hash, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		
		if db.Unscoped().Where("username = ?", adminUsername).First(&superAdminUser).Error != nil {
			superAdminUser = models.User{
				Username:     adminUsername,
				PasswordHash: string(hash),
				Role:         models.RoleSuperAdmin,
			}
			db.Create(&superAdminUser)
			log.Printf("Seeded: Created superadmin user (%s)\n", adminUsername)
		} else {
			// Update existing admin password to match current environment config
			db.Model(&superAdminUser).Update("password_hash", string(hash))
			log.Printf("Seeded: Updated superadmin user password (%s)\n", adminUsername)
		}
	}
}
