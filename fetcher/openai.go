package fetcher

import (
	"context"
	"fmt"

	"ewintr.nl/yogai/model"
	"github.com/sashabaranov/go-openai"
)

const summarizePrompt = `You are an helpful assistant. Your task is to extract all text that refers to the content of a yoga workout video from the description a user gives you.
You will not add introductory sentences like "This text is about", or "Summary of...". Just give the words verbatim. Trim any white space back to a simple space
`

type OpenAI struct {
	client *openai.Client
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		client: openai.NewClient(apiKey),
	}
}

func (o *OpenAI) FetchSummary(video *model.Video) error {
	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: summarizePrompt,
				},

				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("%s\n\n%s", video.YoutubeTitle, video.YoutubeDescription),
				},
			},
		})

	if err != nil {
		return fmt.Errorf("failed to fetch summary: %w", err)
	}

	video.Summary = resp.Choices[len(resp.Choices)-1].Message.Content

	return nil
}
