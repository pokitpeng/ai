package history

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Message represents a single message in the conversation
type Message struct {
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`   // message content
	Timestamp time.Time `json:"timestamp"` // when the message was sent
}

// Session represents a conversation session
type Session struct {
	ID        string    `json:"id"`         // unique session ID
	Messages  []Message `json:"messages"`   // messages in this session
	CreatedAt time.Time `json:"created_at"` // when the session was created
	UpdatedAt time.Time `json:"updated_at"` // when the session was last updated
}

// SessionInfo contains basic information about a session
type SessionInfo struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Preview      string    `json:"preview"`
	MessageCount int       `json:"message_count"`
}

// Manager handles conversation history
type Manager struct {
	currentSession *Session
	storagePath    string
}

// IsEmpty checks if the current session has any messages
func (m *Manager) IsEmpty() bool {
	return m.currentSession == nil || len(m.currentSession.Messages) == 0
}

// NewManager creates a new history manager
func NewManager(storagePath string) (*Manager, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, err
	}

	// Create sessions directory if it doesn't exist
	sessionsPath := filepath.Join(storagePath, "sessions")
	if err := os.MkdirAll(sessionsPath, 0755); err != nil {
		return nil, err
	}

	manager := &Manager{
		storagePath: storagePath,
	}

	// Try to load the current session
	if err := manager.loadCurrentSession(); err != nil {
		// If there's no current session, create a new one
		manager.New()
	}

	return manager, nil
}

// AddUserMessage adds a user message to the current session
func (m *Manager) AddUserMessage(content string) {
	m.currentSession.Messages = append(m.currentSession.Messages, Message{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	})
	m.currentSession.UpdatedAt = time.Now()
	m.saveCurrentSession()
	// Also save to sessions directory
	m.saveSessionToFile(m.currentSession)
}

// AddAssistantMessage adds an assistant message to the current session
func (m *Manager) AddAssistantMessage(content string) {
	m.currentSession.Messages = append(m.currentSession.Messages, Message{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	})
	m.currentSession.UpdatedAt = time.Now()
	m.saveCurrentSession()
	// Also save to sessions directory
	m.saveSessionToFile(m.currentSession)
}

// GetMessages returns all messages in the current session
func (m *Manager) GetMessages() []Message {
	return m.currentSession.Messages
}

// New creates a new session
func (m *Manager) New() {
	m.currentSession = &Session{
		ID:        generateSessionID(),
		Messages:  []Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.saveCurrentSession()
}

// ListSessions returns a list of all available sessions
func (m *Manager) ListSessions() ([]SessionInfo, error) {
	sessionsPath := filepath.Join(m.storagePath, "sessions")
	files, err := os.ReadDir(sessionsPath)
	if err != nil {
		return nil, err
	}

	var sessions []SessionInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			sessionID := strings.TrimSuffix(file.Name(), ".json")
			session, err := m.loadSessionFromFile(sessionID)
			if err != nil {
				continue // Skip sessions that can't be loaded
			}

			// Create a preview from the first user message
			preview := ""
			messageCount := len(session.Messages)
			if messageCount > 0 {
				for _, msg := range session.Messages {
					if msg.Role == "user" {
						// Truncate long messages
						if len(msg.Content) > 50 {
							preview = msg.Content[:50] + "..."
						} else {
							preview = msg.Content
						}
						break
					}
				}
			}

			sessions = append(sessions, SessionInfo{
				ID:           session.ID,
				CreatedAt:    session.CreatedAt,
				UpdatedAt:    session.UpdatedAt,
				Preview:      preview,
				MessageCount: messageCount,
			})
		}
	}

	// Sort sessions by UpdatedAt (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// SwitchSession switches to a different session
func (m *Manager) SwitchSession(sessionID string) error {
	session, err := m.loadSessionFromFile(sessionID)
	if err != nil {
		return err
	}

	m.currentSession = session
	return m.saveCurrentSession()
}

// DeleteSession deletes a session
func (m *Manager) DeleteSession(sessionID string) error {
	// Don't allow deleting the current session
	if m.currentSession != nil && m.currentSession.ID == sessionID {
		return errors.New("cannot delete the current session")
	}

	sessionPath := filepath.Join(m.storagePath, "sessions", sessionID+".json")
	return os.Remove(sessionPath)
}

// GetCurrentSessionID returns the ID of the current session
func (m *Manager) GetCurrentSessionID() string {
	if m.currentSession == nil {
		return ""
	}
	return m.currentSession.ID
}

// Internal methods for saving and loading sessions
func (m *Manager) saveCurrentSession() error {
	sessionPath := filepath.Join(m.storagePath, "current_session.json")
	data, err := json.MarshalIndent(m.currentSession, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionPath, data, 0644)
}

func (m *Manager) loadCurrentSession() error {
	sessionPath := filepath.Join(m.storagePath, "current_session.json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return err
	}

	session := &Session{}
	if err := json.Unmarshal(data, session); err != nil {
		return err
	}

	m.currentSession = session
	return nil
}

func (m *Manager) saveSessionToFile(session *Session) error {
	sessionPath := filepath.Join(m.storagePath, "sessions", session.ID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(sessionPath, data, 0644)
}

func (m *Manager) loadSessionFromFile(sessionID string) (*Session, error) {
	sessionPath := filepath.Join(m.storagePath, "sessions", sessionID+".json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}

	session := &Session{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil, err
	}

	return session, nil
}

// Helper function to generate a unique session ID
func generateSessionID() string {
	return time.Now().Format("20060102-150405-") + randomString(6)
}

// Helper function to generate a random string
func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// FormatHistoryAsPrompt formats the conversation history as a prompt
// that can be prepended to the current question
func (m *Manager) FormatHistoryAsPrompt() string {
	if len(m.currentSession.Messages) == 0 {
		return ""
	}

	var formattedHistory strings.Builder
	formattedHistory.WriteString("Previous conversation:\n\n")

	for _, msg := range m.currentSession.Messages {
		if msg.Role == "user" {
			formattedHistory.WriteString("User: ")
		} else {
			formattedHistory.WriteString("Assistant: ")
		}
		formattedHistory.WriteString(msg.Content)
		formattedHistory.WriteString("\n\n")
	}

	formattedHistory.WriteString("Current question: ")
	return formattedHistory.String()
}
