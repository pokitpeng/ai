# AI Terminal Tool

A command line tool that allows programmers to interact directly with AI large language models in the terminal.

## Features

- Ask questions directly to AI models
- Ask questions based on file content
- Support for multiple AI model management
- Support for asking multiple models simultaneously
- Model-specific default settings
- Based on Cobra framework, with good extensibility
- Support for conversation sessions and history management
- Session management

## Installation

```bash
go install github.com/pokitpeng/ai@latest
```

## Usage

### Basic Question
```bash
ai "How to implement quicksort algorithm?"
```

### View Available Models
```bash
ai model list
```

### Set Default Model
```bash
ai model set openai
```

### Add New Model
```bash
# Basic model addition
ai model add openai-gpt4 https://api.openai.com your-api-key

# Add model with options
ai model add openai-gpt4 https://api.openai.com your-api-key --default --temperature 0.5 --max-tokens 4096 --stream
```

### Remove Model
```bash
ai model remove openai-gpt4
```

### Set Model Options
```bash
# Set default chat options for a model
ai model options openai-gpt4 --temperature 0.2 --max-tokens 4096 --stream

# Make a model the default
ai model options openai-gpt4 --default
```

### Ask Questions Based on File
```bash
ai file main.go "Explain what this code does"
```

### Ask Multiple Models Simultaneously
```bash
ai multi openai,anthropic "What is functional programming?"
```

### Upgrade Tool
```bash
ai upgrade
```

### Session Management
```bash
# Start a new session
ai new

# Continue the conversation in the current session
ai session chat "What is functional programming?"

# List all sessions
ai session list

# Switch to a specific session
ai session switch <session-id>
```

## Configuration

Configuration file is located at `~/.ai/config.yaml`

### Model Configuration Options

Each model can have its own default settings:

- **DefaultEnabled**: When true, this model will be used as the default model
- **DefaultChatOptions**: Default options for chat requests
  - **Temperature**: Controls randomness (0.0-1.0)
  - **MaxTokens**: Maximum number of tokens in the response
  - **Stream**: Whether to stream the response in real-time

## License

MIT
