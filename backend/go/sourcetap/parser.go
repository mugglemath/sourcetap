package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"sourcetap/utils"
)

type parsedResult struct {
	ParsedDescription  string   `json:"ParsedDescription"`
	DeadlineDate       string   `json:"DeadlineDate"`
	MinDegree          string   `json:"MinDegree"`
	MinYearsExperience int      `json:"MinYearsExperience"`
	Modality           string   `json:"Modality"`
	Domain             string   `json:"Domain"`
	Languages          []string `json:"Languages"`
	Frameworks         []string `json:"Frameworks"`
}

func Parser(jobs []Job) []Job {
	apiKey := utils.GetEnv("OPENAI_API_KEY")
	prompt := getLLMPrompt()

	client := utils.CreateOpenAIClient(apiKey)

	for i, job := range jobs {
		message := fmt.Sprintf("%s\n\nJob description: %s", prompt, job.Description)

		chatResp, err := utils.SendMessage(&client, message, prompt)
		if err != nil {
			log.Printf("Error sending job %s to API: %v", job.JobId, err)
			continue
		}

		responseText := chatResp.Choices[0].Message.Content
		var res parsedResult
		err = json.Unmarshal([]byte(responseText), &res)
		if err != nil {
			log.Printf("Error decoding LLM response for job %s: %v. Response was: %s", job.JobId, err, responseText)
			continue
		}

		fmt.Println("Parsed result:")
		fmt.Printf("Parsed Description: %s\n", res.ParsedDescription)
		fmt.Printf("Deadline Date: %s\n", res.DeadlineDate)
		fmt.Printf("Minimum Degree: %s\n", res.MinDegree)
		fmt.Printf("Minimum Years Experience: %d\n", res.MinYearsExperience)
		fmt.Printf("Modality: %s\n", res.Modality)
		fmt.Printf("Domain: %s\n", res.Domain)

		fmt.Println("Languages:")
		for _, lang := range res.Languages {
			fmt.Printf("  - %s\n", lang)
		}

		fmt.Println("Frameworks:")
		for _, fw := range res.Frameworks {
			fmt.Printf("  - %s\n", fw)
		}

		jobs[i].ParsedDescription = res.ParsedDescription
		jobs[i].ExpiresDate = res.DeadlineDate
		jobs[i].MinDegree = res.MinDegree
		jobs[i].MinYearsExperience = res.MinYearsExperience

		if res.Modality != "" {
			jobs[i].Modality = Modality{
				Name: res.Modality,
			}
		}

		if res.Domain != "" {
			jobs[i].Domain = Domain{
				Name: res.Domain,
			}
		}

		for _, langName := range res.Languages {
			lang := Language{
				Name: langName,
			}
			jobs[i].Languages = append(jobs[i].Languages, lang)
		}

		for _, fwName := range res.Frameworks {
			fw := Framework{
				Name: fwName,
			}
			jobs[i].Frameworks = append(jobs[i].Frameworks, fw)
		}

		log.Printf("Job %s updated. MinDegree: %s, MinYearsExperience: %d, Domain: %s, Modality: %s",
			job.JobId, res.MinDegree, res.MinYearsExperience, res.Domain, res.Modality)
	}

	return jobs
}

func getLLMPrompt() string {
	path := os.Getenv("LLM_PROMPT_PATH")
	if path == "" {
		log.Fatal("LLM_PROMPT_PATH not defined in environment")
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read prompt file: %v", err)
	}
	return string(bytes)
}
