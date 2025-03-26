package main

import (
	"fmt"
	"log"
	"os"
	"sourcetap/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// load env variables
	utils.LoadEnvironmentVariables()

	// establish db connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// migrate schema from structs in models.go
	db.AutoMigrate(&Job{}, &Language{}, &Framework{})

	// // run scraper
	// jobs := Scraper()

	// // parse job descriptions
	// Parser(jobs)
}
