# Tracer Bullet

A Go-based command-line tool for tracking agile project activities, including story management, commits, and pair programming sessions.

![Example](.github/tracer_example.gif)

## Project Structure

```
tracer-bullet/
├── cmd/
│   └── tracer/          # Main CLI application
├── internal/
│   ├── commands/        # CLI command implementations
│   ├── config/          # Configuration management
│   ├── jira/           # JIRA integration
│   ├── story/          # Story management
│   └── utils/          # Utility functions
├── stories/            # Story data storage
└── tests/              # Test data
```

## Tech Stack

- **Language**: Go 1.16+
- **Dependencies**:
  - [Cobra](https://github.com/spf13/cobra) - CLI framework
  - [Viper](https://github.com/spf13/viper) - Configuration management
  - [Testify](https://github.com/stretchr/testify) - Testing utilities
  - [GolangCI-Lint](https://golangci-lint.run/) - Code linting

## Installation

### Prerequisites

- Go 1.16 or later
- Git

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/helmedeiros/tracer-bullet.git
cd tracer-bullet
```

2. Build the project:

```bash
make build
```

3. Move the binary to your PATH:

```bash
sudo mv tracer /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/helmedeiros/tracer-bullet/cmd/tracer@latest
```

## Usage

After installation, you can use the `tracer` command:

```bash
tracer --help
```

### Available Commands

#### Configure

```bash
tracer configure --project <name> --user <name>
```

#### Story Management

```bash
# Create a new story
tracer story new --title "Story Title" --description "Description" --tags "tag1,tag2"

# List stories by author
tracer story by --author "author-name"

# Show story files
tracer story files --id <story-id>

# Show story commits
tracer story commits --id <story-id>

# Show story development diary
tracer story diary --id <story-id> [--since <time>] [--until <time>]

# Show story changes
tracer story diff --id <story-id> [--from <time>] [--to <time>]

# Show stories after a commit
tracer story after-hash --hash <commit-hash>
```

#### Commit Management

```bash
# Create a commit
tracer commit create --type <type> --scope <scope> --message "message" [--body "body"] [--breaking] [--jira]
```

#### Pair Programming

```bash
# Start a pair session
tracer pair start <partner-name>

# Show pair status
tracer pair status

# Stop pair session
tracer pair stop
```

#### JIRA Integration

```bash
# Configure JIRA
tracer jira configure --host <jira-host> --token <api-token>

# Link story to JIRA issue
tracer jira link --story <story-id> --issue <jira-issue-id>
```

## Configuration

The tool creates a configuration directory at `~/.tracer` where it stores:

- `config.yaml` - Main configuration file
- `stories/` - Story data files
- `jira/` - JIRA configuration

### Configuration Options

- `project`: Project name
- `user`: User name
- `jira.host`: JIRA instance URL
- `jira.token`: JIRA API token
- `jira.project`: JIRA project key
- `jira.user`: JIRA username

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
