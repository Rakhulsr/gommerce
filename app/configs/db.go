package configs

import (
	"fmt"

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

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Errorf("Failed To Connect To The Database")
		return nil, err
	}

	fmt.Println("DB is successfully connect")

	return db, nil

}
