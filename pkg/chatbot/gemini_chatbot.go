package chatbot

import (
	"backend-ai-supernote/internal/dto"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ChatHistory struct {
	Chat string
	Role string
}

func GetGeminiResponse(
	ctx context.Context,
	apiKey string,
	ChatHistories []*ChatHistory,
) (string, error) {
	chatContents := make([]*dto.GeminiChatContent, 0)
	for _, chatHistory := range ChatHistories {
		chatContents = append(chatContents, &dto.GeminiChatContent{
			Parts: []*dto.GeminiChatParts{
				&dto.GeminiChatParts{
					Text: chatHistory.Chat,
				},
			},
			Role: chatHistory.Role,
		})
	}
	payload := dto.GeminiChatRequest{
		Contents: chatContents,
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent",
		bytes.NewBuffer(payloadJson),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "",
			fmt.Errorf("Status error, got status %d.  With response body %s",
				res.StatusCode, string(resBody))
	}

	var geminiRes dto.GeminiChatResponse
	err = json.Unmarshal(resBody, &geminiRes)
	if err != nil {
		return "", err
	}

	return geminiRes.Candidates[0].Content.Parts[0].Text, nil
}

func DecideToUseRAGE(
	ctx context.Context,
	apiKey string,
	ChatHistories []*ChatHistory,
) (bool, error) {

	chatContents := make([]*dto.GeminiChatContent, 0)
	for _, chatHistory := range ChatHistories {
		chatContents = append(chatContents, &dto.GeminiChatContent{
			Parts: []*dto.GeminiChatParts{
				&dto.GeminiChatParts{
					Text: chatHistory.Chat,
				},
			},
			Role: chatHistory.Role,
		})
	}
	payload := dto.GeminiChatRequest{
		Contents: chatContents,
		GenerationConfig: &dto.GeminiChatGenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema: &dto.GeminiChatResponseSchema{
				Type: "OBJECT",
				Properties: &dto.GeminiChatAppSchema{
					AnswerDirectly: &dto.GeminiChatPropertySchema{
						Type: "BOOLEAN",
					},
				},
				Required: []string{
					"answer_directly",
				},
			},
		},
	}
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(
		"POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent",
		bytes.NewBuffer(payloadJson),
	)
	if err != nil {
		return false, err
	}

	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	if res.StatusCode != http.StatusOK {
		return false,
			fmt.Errorf("Status error, got status %d.  With response body %s",
				res.StatusCode, string(resBody))
	}

	var geminiRes dto.GeminiChatResponse
	err = json.Unmarshal(resBody, &geminiRes)
	if err != nil {
		return false, err
	}
	var appSchema dto.GeminiResponseAppSchema
	err = json.Unmarshal([]byte(geminiRes.Candidates[0].Content.Parts[0].Text), &appSchema)
	if err != nil {
		return false, nil
	}

	return !appSchema.AnswerDirectly, nil
}
