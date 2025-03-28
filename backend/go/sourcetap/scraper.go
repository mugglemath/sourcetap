package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

var processedJobs = make(map[string]struct{})
var jobCount = 0

func Scraper() []Job {
	// setup two collectors - one for search results pages and one for job detail pages
	resultsCollector := colly.NewCollector(
		colly.AllowedDomains("seeker.worksourcewa.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36"),
		colly.Async(false),
	)

	detailsCollector := resultsCollector.Clone()

	detailsCollector.Async = true
	detailsCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 3,
	})

	currentPage := 1
	maxPages := 100

	// storage for all scraped jobs
	jobs := make([]Job, 0)

	var currentPageLinks []string
	var processedLinksCount int

	// search results page handler - finds all job links on a page
	resultsCollector.OnHTML("h2.with-badge", func(e *colly.HTMLElement) {
		jobLink := e.ChildAttr("a", "href")
		if jobLink == "" {
			return
		}

		jobIDRegex := regexp.MustCompile(`JobID=(\d+)`)
		matches := jobIDRegex.FindStringSubmatch(jobLink)
		if len(matches) < 2 {
			return
		}
		jobID := matches[1]

		if !strings.HasPrefix(jobLink, "http") {
			jobLink = "https://seeker.worksourcewa.com" + jobLink
		}

		if _, exists := processedJobs[jobID]; exists {
			fmt.Printf("Skipping already processed job: %s\n", jobID)
			return
		}

		currentPageLinks = append(currentPageLinks, jobLink)

		fmt.Printf("Found job link: %s (ID: %s)\n", jobLink, jobID)
	})

	// when a page is fully scraped, process all found job links
	resultsCollector.OnScraped(func(r *colly.Response) {
		fmt.Printf("Collected %d job links on page %d\n", len(currentPageLinks), currentPage)

		processedLinksCount = 0

		if len(currentPageLinks) == 0 {
			fmt.Printf("No job links found on page %d\n", currentPage)
			if currentPage < maxPages {
				currentPage++
				visitNextPage(resultsCollector, currentPage)
			}
			return
		}

		// visit each job detail page that was found
		for i, link := range currentPageLinks {
			jobCount++
			fmt.Printf("Job #%d: Visiting %s\n", jobCount, link)

			err := detailsCollector.Visit(link)
			if err != nil {
				fmt.Printf("Error visiting %s: %v\n", link, err)
				processedLinksCount++
			}

			if i < len(currentPageLinks)-1 {
				time.Sleep(100 * time.Millisecond)
			}
		}

		detailsCollector.Wait()

		fmt.Printf("All detail collectors have finished for page %d\n", currentPage)

		currentPageLinks = []string{}

		if currentPage < maxPages {
			currentPage++
			visitNextPage(resultsCollector, currentPage)
		} else {
			fmt.Printf("Reached maximum page limit (%d)\n", maxPages)
		}
	})

	// after each job detail page is scraped
	detailsCollector.OnScraped(func(r *colly.Response) {
		processedLinksCount++
		fmt.Printf("Processed job details %d/%d on page %d (URL: %s)\n",
			processedLinksCount, len(currentPageLinks), currentPage, r.Request.URL)
	})

	// job detail page handler - extracts all job information
	detailsCollector.OnHTML("body", func(e *colly.HTMLElement) {
		job := Job{
			Url: e.Request.URL.String(),
		}

		jobIDRegex := regexp.MustCompile(`JobID=(\d+)`)
		if matches := jobIDRegex.FindStringSubmatch(job.Url); len(matches) > 1 {
			job.JobId = matches[1]
			processedJobs[job.JobId] = struct{}{}
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
			log.Printf("failed to retrieve HTML: %v", err)
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
		fmt.Printf("Collected job: %s - %s\n", job.JobId, job.Title)
	})

	// error handlers for both collectors
	resultsCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("Results collector error: URL: %s failed with status code: %d\nError: %v",
			r.Request.URL, r.StatusCode, err)
	})

	detailsCollector.OnError(func(r *colly.Response, err error) {
		log.Printf("Details collector error: URL: %s failed with status code: %d\nError: %v",
			r.Request.URL, r.StatusCode, err)

		processedLinksCount++
	})

	// start the scraping process from the first page
	query := formatQuery(os.Getenv("QUERY"))
	startURL := fmt.Sprintf("https://seeker.worksourcewa.com/jobsearch/powersearch.aspx?q=%s&rad_units=miles&pp=25&nosal=true&vw=b&setype=2", query)

	fmt.Printf("Starting to scrape from %s\n", startURL)
	resultsCollector.Visit(startURL)

	// wait for scraping to complete by checking if job count stabilizes
	lastJobCount := 0
	stableCount := 0
	for {
		if len(jobs) > lastJobCount {
			fmt.Printf("Now have %d jobs...\n", len(jobs))
			lastJobCount = len(jobs)
			stableCount = 0
		} else {
			stableCount++
			if stableCount >= 5 {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("Scraping completed. Collected %d jobs.\n", len(jobs))
	return jobs
}

// construct and visit the next page in the job search results
func visitNextPage(collector *colly.Collector, page int) {
	fmt.Printf("Moving to page %d\n", page)

	query := formatQuery(os.Getenv("QUERY"))

	nextPageURL := fmt.Sprintf(
		"https://seeker.worksourcewa.com/jobsearch/powersearch.aspx?q=%s&rad_units=miles&pp=25&nosal=true&vw=b&setype=2&pg=%d&re=3",
		query, page)

	fmt.Printf("Next page URL: %s\n", nextPageURL)

	time.Sleep(1 * time.Second)

	collector.Visit(nextPageURL)
}

func formatQuery(query string) string {
	query = strings.TrimSpace(strings.ToLower(query))
	return strings.ReplaceAll(query, " ", "+")
}
