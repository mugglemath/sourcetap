package main

import (
	"fmt"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InsertJobs handles inserting or updating jobs in the database
func InsertJobs(db *gorm.DB, jobs []Job) error {
	if len(jobs) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for i := range jobs {
			job := jobs[i]

			var existingJob Job
			result := tx.Where("job_id = ?", job.JobId).First(&existingJob)

			if result.Error != nil {
				if result.Error == gorm.ErrRecordNotFound {
					log.Printf("Creating new job: %s - %s", job.JobId, job.Title)

					if err := processLanguages(tx, &job); err != nil {
						return fmt.Errorf("failed to process languages for job %s: %w", job.JobId, err)
					}

					if err := processFrameworks(tx, &job); err != nil {
						return fmt.Errorf("failed to process frameworks for job %s: %w", job.JobId, err)
					}

					if err := tx.Create(&job).Error; err != nil {
						return fmt.Errorf("failed to create job %s: %w", job.JobId, err)
					}
				} else {
					return fmt.Errorf("error checking for existing job %s: %w", job.JobId, result.Error)
				}
			} else {
				log.Printf("Updating existing job: %s - %s", job.JobId, job.Title)

				if err := tx.Model(&Job{}).Where("job_id = ?", job.JobId).
					Omit("Languages", "Frameworks").Updates(&job).Error; err != nil {
					return fmt.Errorf("failed to update job %s: %w", job.JobId, err)
				}

				if err := tx.Exec("DELETE FROM job_languages WHERE job_id IN (SELECT id FROM jobs WHERE job_id = ?)", job.JobId).Error; err != nil {
					return fmt.Errorf("failed to clear languages for job %s: %w", job.JobId, err)
				}

				if err := processLanguages(tx, &job); err != nil {
					return fmt.Errorf("failed to process languages for job %s: %w", job.JobId, err)
				}

				for _, lang := range job.Languages {
					if err := tx.Exec("INSERT INTO job_languages (job_id, language_id) VALUES ((SELECT id FROM jobs WHERE job_id = ?), ?)",
						job.JobId, lang.ID).Error; err != nil {
						return fmt.Errorf("failed to add language association for job %s: %w", job.JobId, err)
					}
				}

				if err := tx.Exec("DELETE FROM job_frameworks WHERE job_id IN (SELECT id FROM jobs WHERE job_id = ?)", job.JobId).Error; err != nil {
					return fmt.Errorf("failed to clear frameworks for job %s: %w", job.JobId, err)
				}

				if err := processFrameworks(tx, &job); err != nil {
					return fmt.Errorf("failed to process frameworks for job %s: %w", job.JobId, err)
				}

				for _, fw := range job.Frameworks {
					if err := tx.Exec("INSERT INTO job_frameworks (job_id, framework_id) VALUES ((SELECT id FROM jobs WHERE job_id = ?), ?)",
						job.JobId, fw.ID).Error; err != nil {
						return fmt.Errorf("failed to add framework association for job %s: %w", job.JobId, err)
					}
				}
			}
		}
		return nil
	})
}

// processLanguages ensures all languages exist in the DB and associates them with the job
func processLanguages(tx *gorm.DB, job *Job) error {
	var languages []Language
	for i := range job.Languages {
		lang := Language{Name: job.Languages[i].Name}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&lang).Error; err != nil {
			return err
		}

		if err := tx.Where("name = ?", lang.Name).First(&lang).Error; err != nil {
			return err
		}

		languages = append(languages, lang)
	}

	job.Languages = languages
	return nil
}

// processFrameworks ensures all frameworks exist in the DB and associates them with the job
func processFrameworks(tx *gorm.DB, job *Job) error {
	var frameworks []Framework
	for i := range job.Frameworks {
		fw := Framework{Name: job.Frameworks[i].Name}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&fw).Error; err != nil {
			return err
		}

		if err := tx.Where("name = ?", fw.Name).First(&fw).Error; err != nil {
			return err
		}

		frameworks = append(frameworks, fw)
	}

	job.Frameworks = frameworks
	return nil
}

// ToJobMetadata converts a Job to a JobMetadata for API responses
func ToJobMetadata(job Job) JobMetadata {
	languages := make([]string, len(job.Languages))
	for i, lang := range job.Languages {
		languages[i] = lang.Name
	}

	frameworks := make([]string, len(job.Frameworks))
	for i, fw := range job.Frameworks {
		frameworks[i] = fw.Name
	}

	return JobMetadata{
		JobId:              job.JobId,
		Title:              job.Title,
		Company:            job.Company,
		Location:           job.Location,
		PostedDate:         job.PostedDate,
		ExpiresDate:        job.ExpiresDate,
		Salary:             job.Salary,
		Url:                job.Url,
		Modality:           job.Modality,
		MinYearsExperience: job.MinYearsExperience,
		MinDegree:          job.MinDegree,
		Domain:             job.Domain,
		Languages:          languages,
		Frameworks:         frameworks,
		S3Pointer:          job.S3Pointer,
	}
}
