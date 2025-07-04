// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - RAG System Example

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Document represents a document in the knowledge base
type Document struct {
	ID       string
	Title    string
	Content  string
	Metadata map[string]string
}

// Chunk represents a chunk of a document
type Chunk struct {
	ID         string
	DocumentID string
	Content    string
	Vector     []float64
	Metadata   map[string]string
}

// VectorStore manages document chunks and their embeddings
type VectorStore struct {
	chunks []Chunk
}

// RAGSystem combines retrieval and generation
type RAGSystem struct {
	vectorStore *VectorStore
	endpoint    string
	model       string
	documents   []Document
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Chunk      Chunk
	Similarity float64
}

func main() {
	fmt.Println("📚 GoLangGraph RAG System")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("Welcome to the Retrieval-Augmented Generation (RAG) system!")
	fmt.Println()
	fmt.Println("This system can:")
	fmt.Println("  📄 Process and store documents")
	fmt.Println("  🔍 Search through knowledge base")
	fmt.Println("  🧠 Generate answers using retrieved context")
	fmt.Println("  📊 Show relevance scores and sources")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit  - Exit the system")
	fmt.Println("  /help          - Show help message")
	fmt.Println("  /docs          - List all documents")
	fmt.Println("  /search        - Search the knowledge base")
	fmt.Println()

	// Initialize the RAG system
	fmt.Println("🔍 Checking Ollama connection...")
	ragSystem := NewRAGSystem("http://localhost:11434", "gemma3:1b")

	if err := ragSystem.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the model with: ollama pull gemma3:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	// Load sample documents
	fmt.Println("📚 Loading sample documents...")
	ragSystem.loadSampleDocuments()
	fmt.Printf("✅ Loaded %d documents with %d chunks\n", len(ragSystem.documents), len(ragSystem.vectorStore.chunks))
	fmt.Println("✅ RAG system ready for queries")
	fmt.Println()

	// Start interactive session
	ragSystem.startInteractiveSession()
}

// NewRAGSystem creates a new RAG system
func NewRAGSystem(endpoint, model string) *RAGSystem {
	return &RAGSystem{
		vectorStore: &VectorStore{chunks: make([]Chunk, 0)},
		endpoint:    endpoint,
		model:       model,
		documents:   make([]Document, 0),
	}
}

// validateConnection checks if Ollama is running and accessible
func (r *RAGSystem) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", r.endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status: %d", resp.StatusCode)
	}

	return nil
}

// loadSampleDocuments loads sample documents into the system
func (r *RAGSystem) loadSampleDocuments() {
	sampleDocs := []Document{
		{
			ID:       "doc1",
			Title:    "Artificial Intelligence Overview",
			Content:  `Artificial Intelligence (AI) is a branch of computer science that aims to create intelligent machines capable of performing tasks that typically require human intelligence. AI systems can learn, reason, perceive, and make decisions. Machine learning is a subset of AI that enables systems to learn from data without explicit programming. Deep learning, a subset of machine learning, uses neural networks with multiple layers to process complex patterns. AI applications include natural language processing, computer vision, robotics, and autonomous systems. The field has seen rapid advancement in recent years, with applications in healthcare, finance, transportation, and entertainment.`,
			Metadata: map[string]string{"category": "technology", "author": "AI Expert"},
		},
		{
			ID:       "doc2",
			Title:    "Machine Learning Fundamentals",
			Content:  `Machine Learning (ML) is a method of data analysis that automates analytical model building. It uses algorithms that iteratively learn from data, allowing computers to find hidden insights without being explicitly programmed. Supervised learning uses labeled training data to learn a mapping function. Unsupervised learning finds hidden patterns in data without labeled examples. Reinforcement learning learns through interaction with an environment. Common algorithms include linear regression, decision trees, random forests, support vector machines, and neural networks. Feature engineering is crucial for model performance. Cross-validation helps assess model generalization. Overfitting and underfitting are common challenges in ML.`,
			Metadata: map[string]string{"category": "technology", "author": "ML Researcher"},
		},
		{
			ID:       "doc3",
			Title:    "Go Programming Language",
			Content:  `Go, also known as Golang, is an open-source programming language developed by Google. It was designed for simplicity, efficiency, and scalability. Go features strong static typing, garbage collection, and excellent concurrency support through goroutines and channels. The language emphasizes readability and maintainability with its clean syntax. Go compiles to native machine code, providing excellent performance. It has a rich standard library and a growing ecosystem of third-party packages. Go is particularly well-suited for web servers, microservices, cloud applications, and system programming. The language includes built-in testing support and powerful tooling for development and deployment.`,
			Metadata: map[string]string{"category": "programming", "author": "Go Developer"},
		},
		{
			ID:       "doc4",
			Title:    "Cloud Computing Concepts",
			Content:  `Cloud computing delivers computing services over the internet, including servers, storage, databases, networking, software, and analytics. The main service models are Infrastructure as a Service (IaaS), Platform as a Service (PaaS), and Software as a Service (SaaS). Cloud deployment models include public, private, hybrid, and multi-cloud. Benefits include cost reduction, scalability, flexibility, and global accessibility. Major cloud providers include Amazon Web Services (AWS), Microsoft Azure, and Google Cloud Platform. Key technologies include virtualization, containerization, microservices, and serverless computing. Security, compliance, and data governance are critical considerations in cloud adoption.`,
			Metadata: map[string]string{"category": "technology", "author": "Cloud Architect"},
		},
	}

	for _, doc := range sampleDocs {
		r.addDocument(doc)
	}
}

// addDocument adds a document to the RAG system
func (r *RAGSystem) addDocument(doc Document) {
	r.documents = append(r.documents, doc)

	// Chunk the document
	chunks := r.chunkDocument(doc)

	// Generate embeddings for each chunk
	for _, chunk := range chunks {
		chunk.Vector = r.generateEmbedding(chunk.Content)
		r.vectorStore.chunks = append(r.vectorStore.chunks, chunk)
	}
}

// chunkDocument splits a document into smaller chunks
func (r *RAGSystem) chunkDocument(doc Document) []Chunk {
	var chunks []Chunk

	// Simple sentence-based chunking
	sentences := strings.Split(doc.Content, ". ")
	chunkSize := 3 // Group 3 sentences per chunk

	for i := 0; i < len(sentences); i += chunkSize {
		end := i + chunkSize
		if end > len(sentences) {
			end = len(sentences)
		}

		chunkContent := strings.Join(sentences[i:end], ". ")
		if !strings.HasSuffix(chunkContent, ".") && end < len(sentences) {
			chunkContent += "."
		}

		chunk := Chunk{
			ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, i/chunkSize),
			DocumentID: doc.ID,
			Content:    strings.TrimSpace(chunkContent),
			Metadata:   doc.Metadata,
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// generateEmbedding generates a simple embedding for text (simplified TF-IDF approach)
func (r *RAGSystem) generateEmbedding(text string) []float64 {
	// Simple word frequency-based embedding
	words := strings.Fields(strings.ToLower(text))
	wordFreq := make(map[string]int)

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 { // Ignore very short words
			wordFreq[word]++
		}
	}

	// Create a fixed-size vector (simplified)
	vector := make([]float64, 100)
	i := 0
	for word, freq := range wordFreq {
		if i >= 100 {
			break
		}
		// Simple hash-based positioning
		pos := hash(word) % 100
		vector[pos] += float64(freq)
		i++
	}

	// Normalize vector
	norm := 0.0
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector
}

// hash generates a simple hash for a string
func hash(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}

// search performs similarity search in the vector store
func (r *RAGSystem) search(query string, topK int) []SearchResult {
	queryVector := r.generateEmbedding(query)

	var results []SearchResult
	for _, chunk := range r.vectorStore.chunks {
		similarity := cosineSimilarity(queryVector, chunk.Vector)
		results = append(results, SearchResult{
			Chunk:      chunk,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK]
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// generateRAGResponse generates a response using retrieved context
func (r *RAGSystem) generateRAGResponse(query string, searchResults []SearchResult) string {
	// Build context from search results
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Based on the following information:\n\n")

	for i, result := range searchResults {
		contextBuilder.WriteString(fmt.Sprintf("Source %d (Relevance: %.2f):\n%s\n\n",
			i+1, result.Similarity, result.Chunk.Content))
	}

	contextBuilder.WriteString(fmt.Sprintf("Question: %s\n\n", query))
	contextBuilder.WriteString("Please provide a comprehensive answer based on the above sources. If the sources don't contain enough information, please indicate that.")

	return r.callOllama(contextBuilder.String())
}

// callOllama makes a request to the Ollama API
func (r *RAGSystem) callOllama(prompt string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := OllamaRequest{
		Model:  r.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Sprintf("Error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("Error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Ollama returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return fmt.Sprintf("Error unmarshaling response: %v", err)
	}

	return strings.TrimSpace(ollamaResp.Response)
}

// startInteractiveSession runs the interactive RAG session
func (r *RAGSystem) startInteractiveSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("📚 RAG Session Started")
	fmt.Println("Ask questions about AI, ML, Go, or Cloud Computing")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println()

	for {
		fmt.Print("Question: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		if userInput == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(userInput, "/") {
			if userInput == "/quit" || userInput == "/exit" {
				fmt.Println("\n👋 RAG session ended.")
				break
			}

			if r.processCommand(userInput) {
				continue
			}

			fmt.Printf("❓ Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands.")
			continue
		}

		// Process RAG query
		r.processRAGQuery(userInput)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processCommand handles system commands
func (r *RAGSystem) processCommand(command string) bool {
	switch strings.ToLower(command) {
	case "/help":
		r.showHelp()
		return true
	case "/docs":
		r.showDocuments()
		return true
	case "/search":
		r.interactiveSearch()
		return true
	default:
		return false
	}
}

// processRAGQuery processes a user query using RAG
func (r *RAGSystem) processRAGQuery(query string) {
	fmt.Printf("\n🔍 Searching knowledge base for: %s\n", query)
	fmt.Println("─────────────────────────────────────────")

	startTime := time.Now()

	// Retrieve relevant chunks
	searchResults := r.search(query, 3)

	// Display search results
	fmt.Println("📊 Retrieved Sources:")
	for i, result := range searchResults {
		fmt.Printf("%d. Relevance: %.2f | Document: %s\n",
			i+1, result.Similarity, r.getDocumentTitle(result.Chunk.DocumentID))
		fmt.Printf("   Content: %s...\n", truncateText(result.Chunk.Content, 100))
	}
	fmt.Println()

	// Generate response
	fmt.Println("🧠 Generating response...")
	response := r.generateRAGResponse(query, searchResults)

	retrievalTime := time.Since(startTime)

	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("🤖 RAG Response:\n%s\n", response)
	fmt.Printf("⏱️  Processing time: %s\n", formatDuration(retrievalTime))
	fmt.Println()
}

// getDocumentTitle returns the title of a document by ID
func (r *RAGSystem) getDocumentTitle(docID string) string {
	for _, doc := range r.documents {
		if doc.ID == docID {
			return doc.Title
		}
	}
	return "Unknown Document"
}

// truncateText truncates text to a maximum length
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

// interactiveSearch allows users to search without generating responses
func (r *RAGSystem) interactiveSearch() {
	fmt.Print("Enter search query: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		query := strings.TrimSpace(scanner.Text())
		if query != "" {
			results := r.search(query, 5)
			fmt.Printf("\n🔍 Search Results for: %s\n", query)
			fmt.Println("─────────────────────────────────────────")

			for i, result := range results {
				fmt.Printf("%d. Similarity: %.3f\n", i+1, result.Similarity)
				fmt.Printf("   Document: %s\n", r.getDocumentTitle(result.Chunk.DocumentID))
				fmt.Printf("   Content: %s\n", result.Chunk.Content)
				fmt.Println()
			}
		}
	}
}

// showHelp displays help information
func (r *RAGSystem) showHelp() {
	fmt.Println("\n📚 Help - RAG System")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("This RAG system combines document retrieval with AI generation.")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit, /exit   - Exit the system")
	fmt.Println("  /help          - Show this help message")
	fmt.Println("  /docs          - List all documents in the knowledge base")
	fmt.Println("  /search        - Search the knowledge base without generating response")
	fmt.Println()
	fmt.Println("Example questions:")
	fmt.Println("  • 'What is machine learning?'")
	fmt.Println("  • 'How does Go handle concurrency?'")
	fmt.Println("  • 'What are the benefits of cloud computing?'")
	fmt.Println("  • 'Explain the difference between AI and ML'")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  1. Your question is converted to a vector representation")
	fmt.Println("  2. Similar document chunks are retrieved using cosine similarity")
	fmt.Println("  3. Retrieved context is provided to the AI for response generation")
	fmt.Println("  4. The AI generates an answer based on the retrieved information")
	fmt.Println()
}

// showDocuments displays all documents in the knowledge base
func (r *RAGSystem) showDocuments() {
	fmt.Println("\n📄 Knowledge Base Documents")
	fmt.Println("===========================")
	fmt.Println()

	for i, doc := range r.documents {
		fmt.Printf("%d. %s\n", i+1, doc.Title)
		fmt.Printf("   ID: %s\n", doc.ID)
		fmt.Printf("   Category: %s\n", doc.Metadata["category"])
		fmt.Printf("   Author: %s\n", doc.Metadata["author"])
		fmt.Printf("   Content: %s...\n", truncateText(doc.Content, 150))
		fmt.Println()
	}

	fmt.Printf("Total: %d documents, %d chunks\n", len(r.documents), len(r.vectorStore.chunks))
	fmt.Println()
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
