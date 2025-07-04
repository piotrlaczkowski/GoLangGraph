// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Persistence Example

package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
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

// Conversation represents a conversation session
type Conversation struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	Role           string    `json:"role"` // "user" or "assistant"
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	Metadata       string    `json:"metadata,omitempty"`
}

// PersistentAgent manages conversations with persistence
type PersistentAgent struct {
	db                  *sql.DB
	endpoint            string
	model               string
	currentConversation *Conversation
	conversationHistory []Message
}

func main() {
	fmt.Println("💾 GoLangGraph Persistent Agent")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Welcome to the persistent conversation agent!")
	fmt.Println()
	fmt.Println("This agent demonstrates:")
	fmt.Println("  💾 Conversation persistence")
	fmt.Println("  📚 Session management")
	fmt.Println("  🔄 State restoration")
	fmt.Println("  📊 Conversation analytics")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit    - Exit the agent")
	fmt.Println("  /help            - Show help message")
	fmt.Println("  /new             - Start new conversation")
	fmt.Println("  /list            - List all conversations")
	fmt.Println("  /load <id>       - Load conversation by ID")
	fmt.Println("  /delete <id>     - Delete conversation by ID")
	fmt.Println("  /export <id>     - Export conversation to JSON")
	fmt.Println("  /stats           - Show conversation statistics")
	fmt.Println("  /search <query>  - Search conversations")
	fmt.Println()

	// Initialize the persistent agent
	fmt.Println("🔍 Checking Ollama connection...")
	agent, err := NewPersistentAgent("http://localhost:11434", "gemma3:1b")
	if err != nil {
		fmt.Printf("❌ Failed to initialize agent: %v\n", err)
		return
	}
	defer agent.Close()

	if err := agent.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the model with: ollama pull gemma3:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	fmt.Println("✅ Database initialized")
	fmt.Println("✅ Persistent agent ready")
	fmt.Println()

	// Start interactive session
	agent.startPersistentSession()
}

// NewPersistentAgent creates a new persistent agent
func NewPersistentAgent(endpoint, model string) (*PersistentAgent, error) {
	// Initialize SQLite database
	db, err := sql.Open("sqlite3", "conversations.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	agent := &PersistentAgent{
		db:       db,
		endpoint: endpoint,
		model:    model,
	}

	// Create tables
	if err := agent.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return agent, nil
}

// Close closes the database connection
func (p *PersistentAgent) Close() error {
	return p.db.Close()
}

// createTables creates the necessary database tables
func (p *PersistentAgent) createTables() error {
	conversationsTable := `
	CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		message_count INTEGER DEFAULT 0
	);`

	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER NOT NULL,
		role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
	);`

	indexQuery := `
	CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON conversations(updated_at);`

	queries := []string{conversationsTable, messagesTable, indexQuery}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// validateConnection checks if Ollama is running and accessible
func (p *PersistentAgent) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/api/tags", nil)
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

// startPersistentSession runs the interactive persistent session
func (p *PersistentAgent) startPersistentSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("💾 Persistent Session Started")
	fmt.Println("Your conversations will be saved automatically")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println()

	// Show recent conversations
	p.showRecentConversations()

	for {
		prompt := "Chat: "
		if p.currentConversation != nil {
			prompt = fmt.Sprintf("Chat [%s]: ", p.currentConversation.Title)
		}

		fmt.Print(prompt)
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
				fmt.Println("\n👋 Persistent session ended.")
				break
			}

			if p.processCommand(userInput) {
				continue
			}

			fmt.Printf("❓ Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands.")
			continue
		}

		// Process conversation input
		p.processConversationInput(userInput)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processCommand handles system commands
func (p *PersistentAgent) processCommand(command string) bool {
	parts := strings.Fields(command)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		p.showHelp()
		return true
	case "/new":
		p.startNewConversation()
		return true
	case "/list":
		p.listConversations()
		return true
	case "/load":
		if len(parts) > 1 {
			if id, err := strconv.Atoi(parts[1]); err == nil {
				p.loadConversation(id)
			} else {
				fmt.Printf("❌ Invalid conversation ID: %s\n", parts[1])
			}
		} else {
			fmt.Println("Usage: /load <conversation_id>")
		}
		return true
	case "/delete":
		if len(parts) > 1 {
			if id, err := strconv.Atoi(parts[1]); err == nil {
				p.deleteConversation(id)
			} else {
				fmt.Printf("❌ Invalid conversation ID: %s\n", parts[1])
			}
		} else {
			fmt.Println("Usage: /delete <conversation_id>")
		}
		return true
	case "/export":
		if len(parts) > 1 {
			if id, err := strconv.Atoi(parts[1]); err == nil {
				p.exportConversation(id)
			} else {
				fmt.Printf("❌ Invalid conversation ID: %s\n", parts[1])
			}
		} else {
			fmt.Println("Usage: /export <conversation_id>")
		}
		return true
	case "/stats":
		p.showStatistics()
		return true
	case "/search":
		if len(parts) > 1 {
			query := strings.Join(parts[1:], " ")
			p.searchConversations(query)
		} else {
			fmt.Println("Usage: /search <query>")
		}
		return true
	default:
		return false
	}
}

// processConversationInput handles user conversation input
func (p *PersistentAgent) processConversationInput(input string) {
	startTime := time.Now()

	// Create new conversation if none exists
	if p.currentConversation == nil {
		title := generateConversationTitle(input)
		if err := p.createConversation(title); err != nil {
			fmt.Printf("❌ Failed to create conversation: %v\n", err)
			return
		}
	}

	// Save user message
	userMessage := Message{
		ConversationID: p.currentConversation.ID,
		Role:           "user",
		Content:        input,
		CreatedAt:      time.Now(),
	}

	if err := p.saveMessage(userMessage); err != nil {
		fmt.Printf("❌ Failed to save user message: %v\n", err)
		return
	}

	// Build context from conversation history
	context := p.buildConversationContext(input)

	// Get response from Ollama
	response, err := p.callOllama(context)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// Save assistant message
	assistantMessage := Message{
		ConversationID: p.currentConversation.ID,
		Role:           "assistant",
		Content:        response,
		CreatedAt:      time.Now(),
	}

	if err := p.saveMessage(assistantMessage); err != nil {
		fmt.Printf("❌ Failed to save assistant message: %v\n", err)
		return
	}

	// Update conversation history
	p.conversationHistory = append(p.conversationHistory, userMessage, assistantMessage)

	// Update conversation metadata
	p.updateConversation()

	responseTime := time.Since(startTime)
	fmt.Printf("\n🤖 Assistant: %s\n", response)
	fmt.Printf("⏱️  Response time: %s | Saved to conversation: %s\n", formatDuration(responseTime), p.currentConversation.Title)
	fmt.Println()
}

// createConversation creates a new conversation
func (p *PersistentAgent) createConversation(title string) error {
	query := `INSERT INTO conversations (title) VALUES (?) RETURNING id, created_at, updated_at`

	conversation := &Conversation{
		Title: title,
	}

	err := p.db.QueryRow(query, title).Scan(&conversation.ID, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	p.currentConversation = conversation
	p.conversationHistory = make([]Message, 0)

	fmt.Printf("✅ Started new conversation: %s (ID: %d)\n", title, conversation.ID)
	return nil
}

// saveMessage saves a message to the database
func (p *PersistentAgent) saveMessage(message Message) error {
	query := `INSERT INTO messages (conversation_id, role, content, metadata) VALUES (?, ?, ?, ?)`

	_, err := p.db.Exec(query, message.ConversationID, message.Role, message.Content, message.Metadata)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// updateConversation updates conversation metadata
func (p *PersistentAgent) updateConversation() error {
	query := `UPDATE conversations SET updated_at = CURRENT_TIMESTAMP, message_count = (
		SELECT COUNT(*) FROM messages WHERE conversation_id = ?
	) WHERE id = ?`

	_, err := p.db.Exec(query, p.currentConversation.ID, p.currentConversation.ID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// loadConversation loads a conversation by ID
func (p *PersistentAgent) loadConversation(id int) {
	// Load conversation metadata
	query := `SELECT id, title, created_at, updated_at, message_count FROM conversations WHERE id = ?`

	conversation := &Conversation{}
	err := p.db.QueryRow(query, id).Scan(&conversation.ID, &conversation.Title, &conversation.CreatedAt, &conversation.UpdatedAt, &conversation.MessageCount)
	if err != nil {
		fmt.Printf("❌ Failed to load conversation: %v\n", err)
		return
	}

	// Load messages
	messagesQuery := `SELECT id, conversation_id, role, content, created_at, metadata FROM messages WHERE conversation_id = ? ORDER BY created_at`

	rows, err := p.db.Query(messagesQuery, id)
	if err != nil {
		fmt.Printf("❌ Failed to load messages: %v\n", err)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		var metadata sql.NullString

		err := rows.Scan(&message.ID, &message.ConversationID, &message.Role, &message.Content, &message.CreatedAt, &metadata)
		if err != nil {
			fmt.Printf("❌ Failed to scan message: %v\n", err)
			continue
		}

		if metadata.Valid {
			message.Metadata = metadata.String
		}

		messages = append(messages, message)
	}

	p.currentConversation = conversation
	p.conversationHistory = messages

	fmt.Printf("✅ Loaded conversation: %s (ID: %d)\n", conversation.Title, conversation.ID)
	fmt.Printf("   Messages: %d | Created: %s\n", len(messages), conversation.CreatedAt.Format("2006-01-02 15:04"))

	if len(messages) > 0 {
		fmt.Println("   Recent messages:")
		start := len(messages) - 4
		if start < 0 {
			start = 0
		}

		for i := start; i < len(messages); i++ {
			msg := messages[i]
			role := "🤖"
			if msg.Role == "user" {
				role = "👤"
			}

			content := msg.Content
			if len(content) > 60 {
				content = content[:60] + "..."
			}

			fmt.Printf("   %s %s\n", role, content)
		}
	}
	fmt.Println()
}

// buildConversationContext builds context from conversation history
func (p *PersistentAgent) buildConversationContext(currentInput string) string {
	var context strings.Builder

	// Add system prompt
	context.WriteString("You are a helpful and friendly AI assistant. Provide clear, concise, and helpful responses based on the conversation history.\n\n")

	// Add conversation history (last 10 messages)
	start := len(p.conversationHistory) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(p.conversationHistory); i++ {
		msg := p.conversationHistory[i]
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		context.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	// Add current input
	context.WriteString(fmt.Sprintf("User: %s\n", currentInput))
	context.WriteString("Assistant:")

	return context.String()
}

// callOllama makes a request to the Ollama API
func (p *PersistentAgent) callOllama(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := OllamaRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// startNewConversation starts a new conversation
func (p *PersistentAgent) startNewConversation() {
	p.currentConversation = nil
	p.conversationHistory = make([]Message, 0)
	fmt.Println("✅ Ready to start a new conversation")
	fmt.Println("   Type your first message to begin")
}

// listConversations lists all conversations
func (p *PersistentAgent) listConversations() {
	query := `SELECT id, title, created_at, updated_at, message_count FROM conversations ORDER BY updated_at DESC LIMIT 20`

	rows, err := p.db.Query(query)
	if err != nil {
		fmt.Printf("❌ Failed to list conversations: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n📚 Recent Conversations")
	fmt.Println("======================")

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt, &conv.MessageCount)
		if err != nil {
			fmt.Printf("❌ Failed to scan conversation: %v\n", err)
			continue
		}
		conversations = append(conversations, conv)
	}

	if len(conversations) == 0 {
		fmt.Println("No conversations found.")
		return
	}

	for _, conv := range conversations {
		current := ""
		if p.currentConversation != nil && p.currentConversation.ID == conv.ID {
			current = " (current)"
		}

		fmt.Printf("🔹 [%d] %s%s\n", conv.ID, conv.Title, current)
		fmt.Printf("   Messages: %d | Updated: %s\n", conv.MessageCount, conv.UpdatedAt.Format("2006-01-02 15:04"))
	}

	fmt.Printf("\nTotal: %d conversations\n", len(conversations))
	fmt.Println("Use /load <id> to load a conversation")
	fmt.Println()
}

// showRecentConversations shows recent conversations on startup
func (p *PersistentAgent) showRecentConversations() {
	query := `SELECT id, title, updated_at, message_count FROM conversations ORDER BY updated_at DESC LIMIT 5`

	rows, err := p.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.Title, &conv.UpdatedAt, &conv.MessageCount)
		if err != nil {
			continue
		}
		conversations = append(conversations, conv)
	}

	if len(conversations) > 0 {
		fmt.Println("📚 Recent Conversations:")
		for _, conv := range conversations {
			fmt.Printf("   [%d] %s (%d messages)\n", conv.ID, conv.Title, conv.MessageCount)
		}
		fmt.Println("   Use /load <id> to continue a conversation")
		fmt.Println()
	}
}

// deleteConversation deletes a conversation
func (p *PersistentAgent) deleteConversation(id int) {
	// Check if conversation exists
	var title string
	err := p.db.QueryRow("SELECT title FROM conversations WHERE id = ?", id).Scan(&title)
	if err != nil {
		fmt.Printf("❌ Conversation not found: %d\n", id)
		return
	}

	// Delete conversation (messages will be deleted due to CASCADE)
	_, err = p.db.Exec("DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		fmt.Printf("❌ Failed to delete conversation: %v\n", err)
		return
	}

	// Clear current conversation if it was deleted
	if p.currentConversation != nil && p.currentConversation.ID == id {
		p.currentConversation = nil
		p.conversationHistory = make([]Message, 0)
	}

	fmt.Printf("✅ Deleted conversation: %s (ID: %d)\n", title, id)
}

// exportConversation exports a conversation to JSON
func (p *PersistentAgent) exportConversation(id int) {
	// Load conversation
	var conv Conversation
	err := p.db.QueryRow("SELECT id, title, created_at, updated_at, message_count FROM conversations WHERE id = ?", id).Scan(
		&conv.ID, &conv.Title, &conv.CreatedAt, &conv.UpdatedAt, &conv.MessageCount)
	if err != nil {
		fmt.Printf("❌ Conversation not found: %d\n", id)
		return
	}

	// Load messages
	rows, err := p.db.Query("SELECT role, content, created_at FROM messages WHERE conversation_id = ? ORDER BY created_at", id)
	if err != nil {
		fmt.Printf("❌ Failed to load messages: %v\n", err)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	// Create export data
	exportData := map[string]interface{}{
		"conversation": conv,
		"messages":     messages,
		"exported_at":  time.Now(),
	}

	// Write to file
	filename := fmt.Sprintf("conversation_%d_%s.json", id, time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ Failed to create export file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		fmt.Printf("❌ Failed to write export data: %v\n", err)
		return
	}

	fmt.Printf("✅ Exported conversation to: %s\n", filename)
}

// searchConversations searches conversations by content
func (p *PersistentAgent) searchConversations(query string) {
	searchQuery := `
	SELECT DISTINCT c.id, c.title, c.updated_at, c.message_count
	FROM conversations c
	JOIN messages m ON c.id = m.conversation_id
	WHERE m.content LIKE ? OR c.title LIKE ?
	ORDER BY c.updated_at DESC
	LIMIT 10`

	searchTerm := "%" + query + "%"
	rows, err := p.db.Query(searchQuery, searchTerm, searchTerm)
	if err != nil {
		fmt.Printf("❌ Failed to search conversations: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("\n🔍 Search Results for: %s\n", query)
	fmt.Println("==============================")

	var found bool
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.Title, &conv.UpdatedAt, &conv.MessageCount)
		if err != nil {
			continue
		}

		fmt.Printf("🔹 [%d] %s\n", conv.ID, conv.Title)
		fmt.Printf("   Messages: %d | Updated: %s\n", conv.MessageCount, conv.UpdatedAt.Format("2006-01-02 15:04"))
		found = true
	}

	if !found {
		fmt.Println("No conversations found matching your search.")
	}
	fmt.Println()
}

// showStatistics shows conversation statistics
func (p *PersistentAgent) showStatistics() {
	fmt.Println("\n📊 Conversation Statistics")
	fmt.Println("=========================")

	// Total conversations
	var totalConversations int
	p.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&totalConversations)

	// Total messages
	var totalMessages int
	p.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&totalMessages)

	// Average messages per conversation
	var avgMessages float64
	if totalConversations > 0 {
		avgMessages = float64(totalMessages) / float64(totalConversations)
	}

	// Most active conversation
	var mostActiveTitle string
	var mostActiveCount int
	p.db.QueryRow(`
		SELECT c.title, COUNT(m.id) as msg_count
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		GROUP BY c.id, c.title
		ORDER BY msg_count DESC
		LIMIT 1
	`).Scan(&mostActiveTitle, &mostActiveCount)

	// Recent activity
	var recentCount int
	p.db.QueryRow("SELECT COUNT(*) FROM messages WHERE created_at > datetime('now', '-7 days')").Scan(&recentCount)

	fmt.Printf("Total Conversations: %d\n", totalConversations)
	fmt.Printf("Total Messages: %d\n", totalMessages)
	fmt.Printf("Average Messages/Conversation: %.1f\n", avgMessages)
	fmt.Printf("Messages in Last 7 Days: %d\n", recentCount)

	if mostActiveTitle != "" {
		fmt.Printf("Most Active Conversation: %s (%d messages)\n", mostActiveTitle, mostActiveCount)
	}

	fmt.Println()
}

// showHelp displays help information
func (p *PersistentAgent) showHelp() {
	fmt.Println("\n📚 Help - Persistent Agent")
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println("This agent automatically saves all conversations to a local SQLite database.")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit, /exit     - Exit the persistent session")
	fmt.Println("  /help            - Show this help message")
	fmt.Println("  /new             - Start a new conversation")
	fmt.Println("  /list            - List all conversations")
	fmt.Println("  /load <id>       - Load a specific conversation")
	fmt.Println("  /delete <id>     - Delete a conversation")
	fmt.Println("  /export <id>     - Export conversation to JSON")
	fmt.Println("  /stats           - Show conversation statistics")
	fmt.Println("  /search <query>  - Search conversations")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  ✅ Automatic conversation saving")
	fmt.Println("  ✅ Session restoration")
	fmt.Println("  ✅ Conversation management")
	fmt.Println("  ✅ Full-text search")
	fmt.Println("  ✅ JSON export")
	fmt.Println("  ✅ Analytics and statistics")
	fmt.Println()
	fmt.Println("Database: conversations.db (SQLite)")
	fmt.Println()
}

// generateConversationTitle generates a title from the first user input
func generateConversationTitle(input string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return "New Conversation"
	}

	// Take first 5 words or until we reach 50 characters
	var title strings.Builder
	for i, word := range words {
		if i >= 5 || title.Len()+len(word) > 50 {
			break
		}
		if i > 0 {
			title.WriteString(" ")
		}
		title.WriteString(word)
	}

	result := title.String()
	if len(result) > 50 {
		result = result[:47] + "..."
	}

	return result
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
