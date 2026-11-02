package dto

type EmbeddingRequestContentPart struct {
	Text string `json:"text"`
}

type EmbeddingRequestContent struct {
	Parts []EmbeddingRequestContentPart `json:"parts"`
}
type EmbeddingRequest struct {
	Model    string                  `json:"model"`
	Content  EmbeddingRequestContent `json:"content"`
	TaskType string                  `json:"task_type"`
}

type EmbbedingResponseEmbedding struct {
	Values []float32 `json:"values"`
}

type EmbbedingResponse struct {
	Embedding EmbbedingResponseEmbedding `json:"embedding"`
}
