package prompts

import (
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

const AUTOCODER_CODING_RESULTS_FUNCTION_NAME = "coding_results"

func GetComicExamplePrompt() (openai.ChatCompletionRequest, error) {
	systemMessage := openai.ChatCompletionMessage{
		Role:    "system",
		Content: `This is the system message.`,
	}

	userMessage := openai.ChatCompletionMessage{
		Role: "user",
		Content: fmt.Sprintf(`Hello: %s`,
			"world",
		),
	}

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatCompletionMessage{systemMessage, userMessage},
		Tools: []openai.Tool{
			{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        "generate_comics",
					Description: "Generate comics for the given ids",
					Parameters: jsonschema.Definition{
						Type: jsonschema.Object,
						Properties: map[string]jsonschema.Definition{
							"ids": {
								Type: jsonschema.Array,
								Items: &jsonschema.Definition{
									Type:        jsonschema.String,
									Description: "The id of the comic",
								},
								Description: "The ids of the comics to generate",
							},
						},
						Required: []string{"ids"},
					},
				},
			},
		},
	}

	return req, nil
}
