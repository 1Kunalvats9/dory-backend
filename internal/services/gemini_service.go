package services

import (
	"context"
	"dory-backend/internal/config"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func GenerateAIResponse(userQuery string, contextChunks []string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.AppConfig.GeminiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")

	joinedContext := strings.Join(contextChunks, "\n\n---\n\n")

	prompt := fmt.Sprintf(`You are a knowledgeable RAG (Retrieval-Augmented Generation) assistant. Your role is to answer user questions based on their personal documents and information they've shared with you.
	
	CORE PRINCIPLES:
	- Answer questions accurately and directly based on the provided information
	- Match the user's tone and language style (formal if they're formal, casual if they're casual)
	- Be professional, informative, and helpful without being overly polite or robotic
	- Maintain a conversational yet authoritative tone
	- Do NOT use emojis or excessive exclamation marks
	- Do NOT start with phrases like "Based on the context provided" or "According to your documents"
	- Do NOT mention "documents", "chunks", "sources", or "context" explicitly
	- Present information as if naturally recalling it
	
	LANGUAGE REQUIREMENT:
	- Always respond in the SAME LANGUAGE as the user's question
	
	CONTENT GUIDELINES:
	- Use the provided information to answer comprehensively
	- If the user asks something not covered in their information, clearly state that this information isn't available
	- Provide specific details and examples from their information when relevant
	- Keep responses concise but thorough
	- Avoid filler words and unnecessary elaboration
	
	INFORMATION REFERENCE:
	- Do NOT say "Your documents mention..." or "In your files, it says..."
	- Instead, naturally incorporate the information as the answer itself
	
	CONTEXT FROM USER'S INFORMATION:
	%s
	
	USER'S QUESTION:
	%s
	
	Now provide a helpful, natural response:
	`, joinedContext, userQuery)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))

	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
	}

	return "I'm sorry, I couldn't generate a response.", nil
}

func StreamAIResponse(ctx context.Context, userQuery string, contextChunks []string) (any, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.AppConfig.GeminiKey))
	if err != nil {
		return nil, err
	}
	// We don't close the client here because the iterator needs it.
	// The caller should ideally manage the client or we should change how we handle it.
	// However, for simplicity in this RAG setup, we'll let it be for now,
	// but normally you'd want a long-lived client.

	model := client.GenerativeModel("gemini-2.5-flash")

	joinedContext := strings.Join(contextChunks, "\n\n---\n\n")

	prompt := fmt.Sprintf(`You are a knowledgeable RAG (Retrieval-Augmented Generation) assistant. Your role is to answer user questions based on their personal documents and information they've shared with you.
	
	CORE PRINCIPLES:
	- Answer questions accurately and directly based on the provided information
	- Match the user's tone and language style (formal if they're formal, casual if they're casual)
	- Be professional, informative, and helpful without being overly polite or robotic
	- Maintain a conversational yet authoritative tone
	- Do NOT use emojis or excessive exclamation marks
	- Do NOT start with phrases like "Based on the context provided" or "According to your documents"
	- Do NOT mention "documents", "chunks", "sources", or "context" explicitly
	- Present information as if naturally recalling it
	
	LANGUAGE REQUIREMENT:
	- Always respond in the SAME LANGUAGE as the user's question
	
	CONTENT GUIDELINES:
	- Use the provided information to answer comprehensively
	- If the user asks something not covered in their information, clearly state that this information isn't available
	- Provide specific details and examples from their information when relevant
	- Keep responses concise but thorough
	- Avoid filler words and unnecessary elaboration
	
	INFORMATION REFERENCE:
	- Do NOT say "Your documents mention..." or "In your files, it says..."
	- Instead, naturally incorporate the information as the answer itself
	
	CONTEXT FROM USER'S INFORMATION:
	%s
	
	USER'S QUESTION:
	%s
	
	Now provide a helpful, natural response:
	`, joinedContext, userQuery)

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))
	return iter, nil
}

func RetrieveInfoAndSave(userQuery string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.AppConfig.GeminiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-2.5-flash")

	prompt := fmt.Sprintf(`You are a knowledgeable RAG (Retrieval-Augmented Generation) assistant. Your role is to answer user questions based on their personal documents and information they've shared with you.
	You are provided by a user query and you have to check the sentiment of it, if it is kind of informative or something like user is telling you something,
	so i want you to take that info and return the info in a plain text format , make sure to not include any other word or any supportive line, just send the response sent by user, return the information sent by user only nothing else extra
	here is the user query: 
	%s
`, userQuery)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
	}

	return "", fmt.Errorf("no response candidates found")
}
