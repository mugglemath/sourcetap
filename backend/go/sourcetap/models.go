package main

import (
	"time"

	"gorm.io/gorm"
)

// Job struct to be used by the backend and initial scraping
type Job struct {
	gorm.Model
	JobId              string `gorm:"uniqueIndex"`
	Title              string
	Company            string
	Location           string
	ModalityID         uint
	Modality           Modality `gorm:"foreignKey:ModalityID"`
	PostedDate         string
	ExpiresDate        string
	Salary             string
	Url                string
	MinYearsExperience int
	DegreeID           uint
	MinDegree          string
	DomainID           uint
	Domain             Domain `gorm:"foreignKey:DomainID"`
	Description        string `gorm:"type:text"`
	ParsedDescription  string `gorm:"type:text"`
	S3Pointer          string
	Languages          []Language  `gorm:"many2many:job_languages;"`
	Frameworks         []Framework `gorm:"many2many:job_frameworks;"`
}

type Domain struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;type:varchar(30)"`
	Jobs []Job  `gorm:"foreignKey:DomainID"`
}

type Modality struct {
	gorm.Model
	Name       string     `gorm:"uniqueIndex;type:varchar(30)"`
	Modalities []Modality `gorm:"foreignKey:ModalityID"`
}

type Degree struct {
	gorm.Model
	Name    string   `gorm:"uniqueIndex;type:varchar(30)"`
	Degrees []Degree `gorm:"foreignKey:DegreeID"`
}

type Language struct {
	gorm.Model
	Name string `gorm:"uniqueIndex"`
	Jobs []Job  `gorm:"many2many:job_languages;"`
}

type Framework struct {
	gorm.Model
	Name string `gorm:"uniqueIndex"`
	Jobs []Job  `gorm:"many2many:job_frameworks;"`
}

type JobLanguage struct {
	JobID      uint `gorm:"primaryKey"`
	LanguageID uint `gorm:"primaryKey"`
	CreatedAt  time.Time
}

type JobFramework struct {
	JobID       uint `gorm:"primaryKey"`
	FrameworkID uint `gorm:"primaryKey"`
	CreatedAt   time.Time
}

// JobMetadata for API responses without full text content
type JobMetadata struct {
	JobId              string
	Title              string
	Company            string
	Location           string
	PostedDate         string
	ExpiresDate        string
	Salary             string
	Url                string
	Modality           string
	MinYearsExperience int
	MinDegree          bool
	Domain             string
	Languages          []string
	Frameworks         []string
	S3Pointer          string
}

// ValidDomains represents the allowed domain values
func SeedDomains(db *gorm.DB) error {
	domains := []Domain{
		{Name: "Backend"},
		{Name: "Full-Stack"},
		{Name: "AI/ML"},
		{Name: "Data"},
		{Name: "QA"},
		{Name: "Front-End"},
		{Name: "Security"},
		{Name: "DevOps"},
		{Name: "Mobile"},
		{Name: "Site Reliability"},
		{Name: "Networking"},
		{Name: "Embedded Systems"},
		{Name: "Gaming"},
		{Name: "Financial"},
		{Name: "Other"},
	}

	return db.FirstOrCreate(&domains).Error
}

func SeedModalities(db *gorm.DB) error {
	modalities := []Modality{
		{Name: "In-Office"},
		{Name: "Hybrid"},
		{Name: "Remote"},
	}

	return db.FirstOrCreate(&modalities).Error
}

func SeedDegrees(db *gorm.DB) error {
	degrees := []Degree{
		{Name: "Bachelor's"},
		{Name: "Master's"},
		{Name: "Ph.D"},
		{Name: "Unspecified"},
	}

	return db.FirstOrCreate(&degrees).Error
}
