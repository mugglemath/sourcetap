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
	Modality           string
	PostedDate         string
	ExpiresDate        string
	Salary             string
	Url                string
	MinYearsExperience int
	Degree             string
	MinDegree          string
	Domain             string
	Description        string `gorm:"type:text"`
	ParsedDescription  string `gorm:"type:text"`
	S3Pointer          string
	Languages          []Language  `gorm:"many2many:job_languages;"`
	Frameworks         []Framework `gorm:"many2many:job_frameworks;"`
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

// AllowedDomains are the valid domain strings.
var AllowedDomains = []string{
	"Backend",
	"Full-Stack",
	"AI/ML",
	"Data",
	"QA",
	"Front-End",
	"Security",
	"DevOps",
	"Mobile",
	"Site Reliability",
	"Networking",
	"Embedded Systems",
	"Gaming",
	"Financial",
	"Other",
}

// AllowedModalities are the valid modality strings.
var AllowedModalities = []string{
	"In-Office",
	"Hybrid",
	"Remote",
}

// AllowedDegrees are the valid degree strings.
var AllowedDegrees = []string{
	"Bachelor's",
	"Master's",
	"Ph.D",
	"Unspecified",
}
