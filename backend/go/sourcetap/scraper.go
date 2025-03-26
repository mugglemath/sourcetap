package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

func Scraper() []Job {
	// Create base collector for search results pages
	resultsCollector := colly.NewCollector(
		colly.AllowedDomains("seeker.worksourcewa.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36"),
	)

	// Create another collector for individual job details pages
	detailsCollector := resultsCollector.Clone()

	// Track jobs to avoid duplicates
	processedJobs := make(map[string]struct{})
	// currentPage := 1
	jobCount := 0
	maxJobs := 3

	// Create a slice to store all jobs
	jobs := make([]Job, 0)

	// Handle job listings on search results page
	resultsCollector.OnHTML("h2.with-badge", func(e *colly.HTMLElement) {

		if jobCount >= maxJobs {
			return
		}
		// Extract job link
		jobLink := e.ChildAttr("a", "href")
		if jobLink == "" {
			return
		}

		// Extract JobId from the job link
		jobIDRegex := regexp.MustCompile(`JobID=(\d+)`)
		matches := jobIDRegex.FindStringSubmatch(jobLink)
		if len(matches) < 2 {
			return
		}
		jobID := matches[1]

		// check if job already processed
		if _, exists := processedJobs[jobID]; exists {
			return
		}

		// mark job as processed
		processedJobs[jobID] = struct{}{}
		jobCount++

		if !strings.HasPrefix(jobLink, "http") {
			jobLink = "https://seeker.worksourcewa.com" + jobLink
		}

		fmt.Printf("Job #%d: Visiting %s\n", jobCount, jobLink)

		// Visit the job details page
		detailsCollector.Visit(jobLink)
		// Small delay to avoid overwhelming the server
		// time.Sleep(1 * time.Second)
	})

	// Handle pagination - find and follow "Next" link
	// resultsCollector.OnHTML("span.btn-toolbar.btn-pagination", func(e *colly.HTMLElement) {
	// 	// Look for the "Next" button
	// 	nextLink := e.ChildAttr("a[title='Next']", "href")
	// 	if nextLink != "" {
	// 		// Extract the JavaScript event and convert to URL
	// 		// Format is typically: javascript:_jsevt(['re',4],['page',2]);
	// 		if strings.Contains(nextLink, "javascript:_jsevt") {
	// 			// Extract the page number
	// 			currentPage++
	// 			fmt.Printf("Moving to page %d\n", currentPage)

	// 			// Construct the next page URL - this is a simplified approach
	// 			nextPageURL := fmt.Sprintf("https://seeker.worksourcewa.com/jobsearch/powersearch.aspx?q=software+engineer&rad_units=miles&pp=25&nosal=true&vw=b&setype=2&page=%d", currentPage)

	// 			// Visit the next page after a delay
	// 			// time.Sleep(2 * time.Second)
	// 			resultsCollector.Visit(nextPageURL)
	// 		}
	// 	} else {
	// 		fmt.Println("No more pages to process.")
	// 	}
	// })

	// Extract job details from the job page
	detailsCollector.OnHTML("body", func(e *colly.HTMLElement) {
		job := Job{
			Url: e.Request.URL.String(),
		}

		jobIDRegex := regexp.MustCompile(`JobID=(\d+)`)
		if matches := jobIDRegex.FindStringSubmatch(job.Url); len(matches) > 1 {
			job.JobId = matches[1]
		}

		job.Title = e.ChildText("h1.margin-bottom")
		if job.Title == "" {
			job.Title = e.ChildText("h1.job-view-header")
		}

		job.Company = e.ChildText("h4 .capital-letter")
		if job.Company == "" {
			job.Company = e.ChildText("span.job-view-employer")
		}

		job.Location = e.ChildText("h4 small.wrappable")
		if job.Location == "" {
			job.Location = e.ChildText("span.job-view-location")
		}

		dateText := e.ChildText("p:contains('Posted:')")
		if dateText != "" {
			re := regexp.MustCompile(`Posted:\s*(.+?)\s*-`)
			matches := re.FindStringSubmatch(dateText)
			if len(matches) > 1 {
				job.PostedDate = strings.TrimSpace(matches[1])
			}
		}
		if job.PostedDate == "" {
			job.PostedDate = e.ChildText("span.job-view-posting-date")
		}

		reExpires := regexp.MustCompile(`Expires:\s*<strong>(.*?)</strong>`)
		html, err := e.DOM.Html()
		if err != nil {
			log.Printf("failed to retreive HTML: %v", err)
		} else {
			if matches := reExpires.FindStringSubmatch(html); len(matches) > 1 {
				job.ExpiresDate = strings.TrimSpace(matches[1])
			}
		}

		job.Salary = e.ChildText("p.job-view-salary")
		if job.Salary == "" {
			e.ForEach("dl span", func(_ int, el *colly.HTMLElement) {
				if strings.Contains(el.ChildText("dt"), "Salary") {
					job.Salary = el.ChildText("dd")
				}
			})
		}

		for _, selector := range []string{
			"span#TrackingJobBody",
			"div.JobViewJobBody",
			"div.job-view-description",
			"div.directJobBody",
			"#jobViewFrame",
		} {
			job.Description = e.ChildText(selector)
			if job.Description != "" {
				break
			}
		}

		jobs = append(jobs, job)
	})

	// Handle errors
	resultsCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with status code: %d\nError: %v",
			r.Request.URL, r.StatusCode, err)
	})

	detailsCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed with status code: %d\nError: %v",
			r.Request.URL, r.StatusCode, err)
	})

	// Start the scraping process with the first page
	query := formatQuery(os.Getenv("QUERY"))
	startURL := fmt.Sprintf("https://seeker.worksourcewa.com/jobsearch/powersearch.aspx?q=%s&rad_units=miles&pp=25&nosal=true&vw=b&setype=2", query)

	fmt.Printf("Starting to scrape from %s\n", startURL)
	resultsCollector.Visit(startURL)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(jobs); err != nil {
		log.Fatal("Failed to encode jobs to JSON:", err)
	}

	return jobs
}

func formatQuery(query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	return strings.ReplaceAll(query, " ", "+")
}
