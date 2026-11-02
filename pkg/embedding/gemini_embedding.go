package embedding

import (
	"backend-ai-supernote/internal/dto"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetGeminiEmbedding(
	apiKey string,
	text string,
) (*dto.EmbbedingResponse, error) {
	gemeniReq := dto.EmbeddingRequest{
		Model: "models/gemini-embedding-2",
		Content: dto.EmbeddingRequestContent{
			Parts: []dto.EmbeddingRequestContentPart{
				{
					Text: text,
				},
			},
		},
		TaskType: "RETRIEVAL_DOCUMENT",
	}

	gemeniReqJson, err := json.Marshal(gemeniReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-2:embedContent",
		bytes.NewBuffer(gemeniReqJson),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	resByte, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil,
			fmt.Errorf("error from response, code %d, body %s",
				res.StatusCode,
				string(resByte))
	}

	var resEmbedding dto.EmbbedingResponse
	err = json.Unmarshal(resByte, &resEmbedding)
	if err != nil {
		return nil, err
	}

	return &resEmbedding, nil
}
