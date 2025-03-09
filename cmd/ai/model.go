package ai

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/pokitpeng/ai/pkg/models"
	"github.com/spf13/cobra"
)

// modelCmd represents the model subcommand
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage AI models, including listing, adding, deleting, and setting default models",
	Long:  `Manage AI models, including listing, adding, deleting, and setting default models.`,
}

// modelListCmd lists available models
var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available AI models",
	Long:  `List all configured AI models and their information.`,
	Run: func(cmd *cobra.Command, args []string) {
		modelsList := modelManager.ListModels()
		defaultName := modelManager.GetDefaultModelName()

		if len(modelsList) == 0 {
			fmt.Println("No models configured. Use 'ai model add' to add a model.")
			return
		}

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
			{Number: 1, WidthMax: 6, WidthMin: 6, Align: text.AlignCenter},
			{Number: 2, WidthMax: 25, WidthMin: 10},
			{Number: 3, WidthMax: 30, WidthMin: 10, Transformer: truncateString(30)},
			{Number: 4, WidthMax: 20, WidthMin: 10, Transformer: truncateString(20)},
			{Number: 5, WidthMax: 30, WidthMin: 15},
		})

		// Add header
		t.AppendHeader(table.Row{"Default", "Name", "URL", "API Key", "Parameters"})

		// Get global default options
		globalDefaults := models.DefaultChatOptions()

		// Add data rows
		for name, config := range modelsList {
			defaultMark := " "
			if name == defaultName {
				defaultMark = "✓"
			}

			// Mask API Key
			apiKeyMasked := maskAPIKey(config.APIKey)

			// Prepare options info
			var optionsInfo string
			if config.DefaultChatOptions != nil {
				// Show model-specific default options
				optionsInfo = fmt.Sprintf("Temp:%.2f MaxTokens:%d",
					config.DefaultChatOptions.Temperature,
					config.DefaultChatOptions.MaxTokens)
			} else {
				// Show global default values
				optionsInfo = fmt.Sprintf("Global defaults(Temp:%.2f MaxTokens:%d)",
					globalDefaults.Temperature,
					globalDefaults.MaxTokens)
			}

			// Handle long model names
			shortName := name
			if len(name) > 25 {
				parts := strings.Split(name, "/")
				if len(parts) > 1 {
					shortName = parts[len(parts)-1]
				} else {
					shortName = name[:22] + "..."
				}
			}

			t.AppendRow(table.Row{
				defaultMark,
				shortName,
				config.URL,
				apiKeyMasked,
				optionsInfo,
			})
		}

		// Add footer
		// t.AppendFooter(table.Row{"", "Total", strconv.Itoa(len(modelsList)), "", ""})

		// Set title
		// t.SetTitle("AI Model List")

		// Set caption
		// if defaultName != "" {
		// 	t.SetCaption("Current default model: %s", defaultName)
		// } else {
		// 	t.SetCaption("No default model set")
		// }

		// Render table
		t.Render()
	},
}

// truncateString returns a function to truncate strings
func truncateString(maxLen int) func(val interface{}) string {
	return func(val interface{}) string {
		str, ok := val.(string)
		if !ok {
			return fmt.Sprintf("%v", val)
		}
		if len(str) <= maxLen {
			return str
		}
		return str[:maxLen-3] + "..."
	}
}

// addCmd adds a new model
var addCmd = &cobra.Command{
	Use:   "add <name> <url> <apikey>",
	Short: "Add a new AI model",
	Long:  `Add a new AI model configuration, providing the name, API URL, and API key.`,
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		url := args[1]
		apiKey := args[2]

		// Get flags
		defaultEnabled, _ := cmd.Flags().GetBool("default")

		// Create default chat options
		var chatOptions *models.ChatOptions
		if cmd.Flags().Changed("temperature") || cmd.Flags().Changed("max-tokens") || cmd.Flags().Changed("stream") {
			temperature, _ := cmd.Flags().GetFloat64("temperature")
			maxTokens, _ := cmd.Flags().GetInt("max-tokens")
			stream, _ := cmd.Flags().GetBool("stream")

			chatOptions = &models.ChatOptions{
				Temperature: temperature,
				MaxTokens:   maxTokens,
				Stream:      stream,
			}
		}

		err := modelManager.AddModel(name, url, apiKey, defaultEnabled, chatOptions)
		if err != nil {
			fmt.Printf("Failed to add model: %v\n", err)
			return
		}

		fmt.Printf("Model '%s' added successfully\n", name)
	},
}

// removeCmd removes a model
var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an AI model",
	Long:  `Remove a configured AI model.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := modelManager.RemoveModel(name)
		if err != nil {
			fmt.Printf("Failed to remove model: %v\n", err)
			return
		}

		fmt.Printf("Model '%s' removed successfully\n", name)
	},
}

// setCmd sets the default model
var setCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set the default AI model",
	Long:  `Set the default AI model to use.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := modelManager.SetDefaultModel(name)
		if err != nil {
			fmt.Printf("Failed to set default model: %v\n", err)
			return
		}

		fmt.Printf("Set '%s' as the default model\n", name)
	},
}

// optionsCmd sets model options
var optionsCmd = &cobra.Command{
	Use:   "options <name>",
	Short: "Set options for an AI model",
	Long:  `Set default chat options for an AI model, such as temperature, max tokens, and streaming.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Get the current model config
		config, err := modelManager.GetModelConfig(name)
		if err != nil {
			fmt.Printf("Failed to get model: %v\n", err)
			return
		}

		// Check if we need to create or update chat options
		if config.DefaultChatOptions == nil {
			config.DefaultChatOptions = models.DefaultChatOptions()
		}

		// Update options based on flags
		if cmd.Flags().Changed("temperature") {
			temperature, _ := cmd.Flags().GetFloat64("temperature")
			config.DefaultChatOptions.Temperature = temperature
		}

		if cmd.Flags().Changed("max-tokens") {
			maxTokens, _ := cmd.Flags().GetInt("max-tokens")
			config.DefaultChatOptions.MaxTokens = maxTokens
		}

		if cmd.Flags().Changed("stream") {
			stream, _ := cmd.Flags().GetBool("stream")
			config.DefaultChatOptions.Stream = stream
		}

		if cmd.Flags().Changed("default") {
			defaultEnabled, _ := cmd.Flags().GetBool("default")
			config.DefaultEnabled = defaultEnabled
		}

		// Update the model config
		err = modelManager.UpdateModelConfig(name, config)
		if err != nil {
			fmt.Printf("Failed to update model options: %v\n", err)
			return
		}

		fmt.Printf("Updated options for model '%s'\n", name)
		fmt.Printf("Temperature: %.2f, MaxTokens: %d, Stream: %v, Default: %v\n",
			config.DefaultChatOptions.Temperature,
			config.DefaultChatOptions.MaxTokens,
			config.DefaultChatOptions.Stream,
			config.DefaultEnabled)
	},
}

// Register commands in init
func init() {
	rootCmd.AddCommand(modelCmd)
	modelCmd.AddCommand(modelListCmd)
	modelCmd.AddCommand(addCmd)
	modelCmd.AddCommand(removeCmd)
	modelCmd.AddCommand(setCmd)
	modelCmd.AddCommand(optionsCmd)

	// Add flags for add command
	addCmd.Flags().Bool("default", false, "Set this model as the default")
	addCmd.Flags().Float64("temperature", 0.2, "Set default temperature (0.0-1.0)")
	addCmd.Flags().Int("max-tokens", 2048, "Set default maximum tokens")
	addCmd.Flags().Bool("stream", true, "Enable streaming output by default")

	// Add flags for options command
	optionsCmd.Flags().Float64("temperature", 0.2, "Set default temperature (0.0-1.0)")
	optionsCmd.Flags().Int("max-tokens", 2048, "Set default maximum tokens")
	optionsCmd.Flags().Bool("stream", true, "Enable streaming output by default")
	optionsCmd.Flags().Bool("default", false, "Set this model as the default")
}

// maskAPIKey masks the API key
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}

	prefix := apiKey[:4]
	suffix := apiKey[len(apiKey)-4:]
	masked := prefix + strings.Repeat("*", len(apiKey)-8) + suffix

	return masked
}
