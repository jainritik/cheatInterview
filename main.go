package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const openAIAPIKey = "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"
const openAIEndpoint = "https://api.openai.com/v1/chat/completions"

type ExtractRequest struct {
	ImageDataList []string `json:"imageDataList"`
	Language      string   `json:"language"`
}

type GPTRequest struct {
	Model     string       `json:"model"`
	Messages  []GPTMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
}

type GPTMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type GPTImage struct {
	Type string            `json:"type"`
	Data map[string]string `json:"image_url"`
}

type GPTText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type GPTResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// Util to send to GPT
func callGPTWithImages(imageDataList []string, prompt string) (string, error) {
	var contentItems []interface{}

	log.Printf("[GPT] Prompt: %s", prompt)

	// Add text prompt
	contentItems = append(contentItems, GPTText{Type: "text", Text: prompt})

	// Add image content
	for i, base64img := range imageDataList {
		log.Printf("[GPT] Image %d: base64 length = %d", i+1, len(base64img))
		contentItems = append(contentItems, GPTImage{
			Type: "image_url",
			Data: map[string]string{
				"url": "data:image/png;base64," + base64img,
			},
		})
	}

	// Select model
	modelToUse := "gpt-3.5-turbo"
	if len(imageDataList) > 0 {
		modelToUse = "gpt-4-turbo"
	}
	log.Printf("[GPT] Using model: %s", modelToUse)

	// Build request
	request := GPTRequest{
		Model:     modelToUse,
		Messages:  []GPTMessage{{Role: "user", Content: contentItems}},
		MaxTokens: 1000,
	}

	reqBytes, _ := json.MarshalIndent(request, "", "  ")
	log.Printf("[GPT] Payload JSON: %s", string(reqBytes))

	req, err := http.NewRequest("POST", openAIEndpoint, strings.NewReader(string(reqBytes)))
	if err != nil {
		log.Printf("[GPT ERROR] Failed to create request: %v", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+openAIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[GPT ERROR] Request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("[GPT] Raw response: %s", string(bodyBytes))

	var parsed GPTResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		log.Printf("[GPT ERROR] Failed to parse response JSON: %v", err)
		return "", err
	}

	if len(parsed.Choices) > 0 {
		log.Printf("[GPT] Response extracted successfully.")
		return parsed.Choices[0].Message.Content, nil
	}

	log.Printf("[GPT ERROR] No choices returned by OpenAI.")
	return "", fmt.Errorf("no response from GPT")
}

func extractHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[API] /api/extract called")

	var req ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[API] Received %d images, language: %s", len(req.ImageDataList), req.Language)

	responseText, err := callGPTWithImages(req.ImageDataList, "Extract the problem statement from this image, in clean text format.")
	if err != nil {
		log.Printf("[ERROR] GPT error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("[API] Extraction successful")
	json.NewEncoder(w).Encode(map[string]string{
		"problem": responseText,
	})
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[API] /api/generate called")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[ERROR] Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	questionVal, ok := payload["question"]
	if !ok || questionVal == nil {
		log.Println("[ERROR] Missing 'question' field")
		http.Error(w, "Missing 'question' field", http.StatusBadRequest)
		return
	}
	question, ok := questionVal.(string)
	if !ok || strings.TrimSpace(question) == "" {
		log.Println("[ERROR] Empty 'question' string")
		http.Error(w, "'question' must be a non-empty string", http.StatusBadRequest)
		return
	}

	log.Printf("[API] Generating code for question: %.60s...", question)

	prompt := fmt.Sprintf(`
	Write a clean, properly formatted, and idiomatic C++ solution to the following problem. 
	Make sure the code includes:
	- Proper indentation and spacing
	- Descriptive comments where necessary
	- No markdown, no code block formatting — just raw code

	Use this response format exactly:
	CODE:
	<Insert only raw, cleanly formatted C++ code here>

	EXPLANATION:
	<Insert a brief plain-text explanation of the code logic and approach>

	Problem:
	%s
	`, question)

	responseText, err := callGPTWithImages([]string{}, prompt)
	if err != nil {
		log.Printf("[ERROR] GPT call failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	code := ""
	explanation := ""
	if strings.Contains(responseText, "EXPLANATION:") {
		parts := strings.SplitN(responseText, "EXPLANATION:", 2)
		code = strings.TrimSpace(strings.TrimPrefix(parts[0], "CODE:"))
		explanation = strings.TrimSpace(parts[1])
	} else {
		code = strings.TrimSpace(responseText)
	}

	log.Println("[API] Code generation successful")
	json.NewEncoder(w).Encode(map[string]string{
		"code":        code,
		"explanation": explanation,
	})
}

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	log.Printf("✅ Server running at http://localhost:%s", port)
	http.Handle("/api/extract", loggingMiddleware(http.HandlerFunc(extractHandler)))
	http.Handle("/api/generate", loggingMiddleware(http.HandlerFunc(generateHandler)))

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
