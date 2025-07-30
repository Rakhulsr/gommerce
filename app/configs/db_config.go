package configs

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenConnection() (*gorm.DB, error) {

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		LoadENV.DBUser,
		LoadENV.DBPassword,
		LoadENV.DBHost,
		LoadENV.DBPort,
		LoadENV.DBName,
	)

	maxRetries := 10
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to database (Attempt %d/%d) using DSN: %s", i+1, maxRetries, dsn)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {

			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				pingErr = sqlDB.Ping()
				if pingErr == nil {
					log.Println("✅ Database connection successful!")
					return db, nil
				}
			}

			log.Printf("❌ Failed to ping database: %v. Retrying in %v...", pingErr, retryDelay)
		} else {
			log.Printf("❌ Failed to open GORM connection: %v. Retrying in %v...", err, retryDelay)
		}

		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("Failed to connect to the database after %d retries. Last DSN used: %s", maxRetries, dsn)
}
