package main

import "sourcetap/utils"

func main() {
	utils.LoadEnvironmentVariables()
	jobs := Scraper()
	Parser(jobs)
}
