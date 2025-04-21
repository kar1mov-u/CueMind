package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const cueCardPrompt = `
You are an AI assistant helping to create cue cards from study material.
The user has uploaded a file containing educational content (lecture notes, textbook sections, or reference material). Read and analyze the file content carefully.
Your task is to extract important and detailed concepts, definitions, and explanations, and format them as cue cards in JSON.
Each cue card should have:
- A front field: a question or prompt (e.g. "Define polymorphism", or "What is the purpose of TCP?")
- A back field: a clear and complete answer or explanation.
Ensure you:
- Do not skip technical details or nuance
- Break down long material into multiple cards if needed
- Cover all major concepts from the file
- Return valid JSON only, no code blocks or extra formatting

Format your response like this:
{
  "cards": [
    {
      "front": "What is ___?",
      "back": "..."
    },
    ...
  ]
}
`

type LLMService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func New(key string) *LLMService {
	ctx := context.TODO()
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Fatalf("Error on creating AI client :%v", err)
	}

	return &LLMService{model: CreateModel(client), client: client}

}
func CreateModel(client *genai.Client) *genai.GenerativeModel {
	model := client.GenerativeModel("gemini-1.5-flash")
	return model
}

func (s *LLMService) GenerateCardsFromFile(ctx context.Context, file io.Reader) (*FlashCardResponse, error) {

	//Upload to Server
	sFile, err := s.client.UploadFile(ctx, "", file, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer s.client.DeleteFile(ctx, sFile.Name)

	//LLM Request
	resp, err := s.model.GenerateContent(ctx, genai.FileData{URI: sFile.URI}, genai.Text(cueCardPrompt))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//polish response
	polishedResp, err := formatLLMResponse((resp.Candidates[0].Content.Parts[0]))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//format to Json objects
	flashCards, err := convertRespToStruct(polishedResp)
	if err != nil {
		log.Printf("error on converting to structs: %v \n", err)
	}
	for _, card := range flashCards.Cards[:5] {
		fmt.Println(card)
	}

	return flashCards, nil
	//save into DB

}

type Card struct {
	Front string
	Back  string
}

type FlashCardResponse struct {
	Cards []Card
}

func convertRespToStruct(resp string) (*FlashCardResponse, error) {
	var flashCards FlashCardResponse
	err := json.Unmarshal([]byte(resp), &flashCards)
	if err != nil {
		return nil, err
	}
	return &flashCards, nil
}

func formatLLMResponse(resp genai.Part) (string, error) {
	var llmText genai.Text
	var ok bool
	if llmText, ok = resp.(genai.Text); !ok {
		return "", fmt.Errorf("error on formating. response doesnt contain text")
	}
	text := string(llmText)
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	return text, nil
}
