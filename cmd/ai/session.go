package ai

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage session history",
	Long:  `View, switch or delete historical conversation sessions.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default to showing session list
		listSessions()
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all historical sessions",
	Long:  `List all available historical sessions, including creation time and preview content.`,
	Run: func(cmd *cobra.Command, args []string) {
		listSessions()
	},
}

var switchCmd = &cobra.Command{
	Use:   "switch [session_id or number]",
	Short: "Switch to a specified session",
	Long:  `Switch to a historical session specified by ID or number. You can use either the session ID or the number shown in the list.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionIdentifier := args[0]
		switchToSession(sessionIdentifier)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [session_id or number]",
	Short: "Delete a specified session",
	Long:  `Delete a historical session specified by ID or number. You can use either the session ID or the number shown in the list.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionIdentifier := args[0]
		deleteSession(sessionIdentifier)
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(switchCmd)
	sessionCmd.AddCommand(deleteCmd)
}

// List all sessions
func listSessions() {
	sessions, err := historyManager.ListSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get session list: %v\n", err)
		return
	}

	if len(sessions) == 0 {
		fmt.Println("No history sessions found")
		return
	}

	currentID := historyManager.GetCurrentSessionID()

	// Create table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// Set table style
	t.SetStyle(table.StyleLight)

	// Customize table style
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateFooter = true
	t.Style().Options.SeparateHeader = true
	t.Style().Options.SeparateRows = true

	// Set column configurations
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMax: 6, WidthMin: 6, Align: text.AlignCenter},           // Current
		{Number: 2, WidthMax: 8, WidthMin: 8, Align: text.AlignCenter},           // No.
		{Number: 3, WidthMax: 30, WidthMin: 30},                                  // ID
		{Number: 4, WidthMax: 20, WidthMin: 20},                                  // Updated At
		{Number: 5, WidthMax: 10, WidthMin: 10, Align: text.AlignCenter},         // Messages
		{Number: 6, WidthMax: 40, WidthMin: 20, Transformer: truncateString(40)}, // Preview
	})

	// Add header
	t.AppendHeader(table.Row{"Current", "No.", "ID", "Updated At", "Messages", "Preview"})

	// Add data rows
	for i, session := range sessions {
		// Format time
		timeStr := session.UpdatedAt.Format("2006-01-02 15:04:05")

		// Mark current session
		currentMarker := " "
		if session.ID == currentID {
			currentMarker = "✓"
		}

		// Add session info to table
		t.AppendRow(table.Row{
			currentMarker,
			i + 1,
			session.ID,
			timeStr,
			session.MessageCount,
			session.Preview,
		})
	}

	// Render table
	t.Render()

	fmt.Println("✓ indicates current session")
	fmt.Println("Use 'ai session switch <number or ID>' to switch session")
	fmt.Println("Use 'ai session delete <number or ID>' to delete session")
}

// Switch to the specified session
func switchToSession(identifier string) {
	// Check if it's a number (number)
	if index, err := strconv.Atoi(identifier); err == nil {
		sessions, err := historyManager.ListSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get session list: %v\n", err)
			return
		}

		// Check if the number is valid
		if index < 1 || index > len(sessions) {
			fmt.Fprintf(os.Stderr, "Invalid session number: %d\n", index)
			return
		}

		// Use the session ID corresponding to the number
		identifier = sessions[index-1].ID
	}

	// Switch session
	if err := historyManager.SwitchSession(identifier); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to switch session: %v\n", err)
		return
	}

	fmt.Printf("Switched to session: %s\n", identifier)
}

// Delete the specified session
func deleteSession(identifier string) {
	// Check if it's a number (number)
	if index, err := strconv.Atoi(identifier); err == nil {
		sessions, err := historyManager.ListSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get session list: %v\n", err)
			return
		}

		// Check if the number is valid
		if index < 1 || index > len(sessions) {
			fmt.Fprintf(os.Stderr, "Invalid session number: %d\n", index)
			return
		}

		// Use the session ID corresponding to the number
		identifier = sessions[index-1].ID
	}

	// Check if it's the current session
	if identifier == historyManager.GetCurrentSessionID() {
		fmt.Fprintf(os.Stderr, "Cannot delete the current session, please switch to another session first\n")
		return
	}

	// Delete session
	if err := historyManager.DeleteSession(identifier); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete session: %v\n", err)
		return
	}

	fmt.Printf("Deleted session: %s\n", identifier)
}
