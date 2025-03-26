package utils

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// CreateOpenAIClient creates an OpenAI client using the given API key.
func CreateOpenAIClient(apiKey string) openai.Client {
	return openai.NewClient(option.WithAPIKey(apiKey))
}

// SendMessage sends a message to the OpenAI API and returns the chat completion response.
func SendMessage(client *openai.Client, content string, prompt string) (*openai.ChatCompletion, error) {
	fullMessage := prompt + " " + content

	// Use openai.UserMessage to create a message parameter
	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(fullMessage),
		},
		Model: openai.ChatModelGPT4oMini,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return chatCompletion, nil
}
