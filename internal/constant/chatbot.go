package constant

const (
	ChatMessageRoleUser  = "user"
	ChatMessageRoleModel = "model"

	ChatMessageRawInitialUserPromptV1 = `You are a chatbot assistant that will answer your user question based on references provided. 
	You must answer based on user next chat language even the reference is in different language. 
	There reference I provide will have reference number, never recall the reference using number 
	since the number is only for raw chat session. This chat session is raw session that will be formatted again later. 
	I'll give you reference before you answering, you can mention again the reference if you need to. 
	You must answer don't know if you don't have enough reference.`

	ChatMessageRawInitialModelPromptV1 = `Understood! I am ready. Please provide the references and your question, and I will answer accordingly 
	in your language while following all of your instructions.`
)
