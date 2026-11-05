package dto

type GeminiChatParts struct {
	Text string `json:"text"`
}

type GeminiChatContent struct {
	Parts []*GeminiChatParts `json:"parts"`
	Role  string             `json:"role"`
}

type GeminiChatRequest struct {
	Contents         []*GeminiChatContent        `json:"contents"`
	GenerationConfig *GeminiChatGenerationConfig `json:"generationConfig"`
}

type GeminiChatCandidate struct {
	Content *GeminiChatContent `json:"content"`
}

type GeminiChatResponse struct {
	Candidates []*GeminiChatCandidate `json:"candidates"`
}

type GeminiChatPropertySchema struct {
	Type string `json:"type"`
}

type GeminiChatAppSchema struct {
	AnswerDirectly *GeminiChatPropertySchema `json:"answer_directly"`
}

type GeminiChatResponseSchema struct {
	Type       string               `json:"type"`
	Properties *GeminiChatAppSchema `json:"properties"`
	Required   []string             `json:"required"`
}
type GeminiChatGenerationConfig struct {
	ResponseMimeType string                    `json:"responseMimeType"`
	ResponseSchema   *GeminiChatResponseSchema `json:"responseSchema"`
}

type GeminiResponseAppSchema struct {
	AnswerDirectly bool `json:"answer_directly"`
}
