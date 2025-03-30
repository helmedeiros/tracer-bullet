# Tracer Bullet

A Go-based command-line tool for tracking agile project activities, including story management, commits, and pair programming sessions.

![Example](.github/tracer_example.gif)

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
go build -o tracer cmd/tracer/main.go
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

- `configure`: Configure tracer settings
- `commit`: Create a commit with story information
- `pair`: Manage pair programming sessions
- `story`: Manage stories and their tracking
  - `new`: Create a new story
  - `after-hash`: Show stories after a specific commit hash
  - `by`: Show stories by author
  - `files`: Show files associated with a story
  - `commits`: Show commits associated with a story
  - `diary`: Show story development diary
  - `diff`: Show story changes

## Configuration

The tool creates a configuration directory at `~/.tracer` where it stores:

- Configuration file (`config.yaml`)
- Story files
- Pair programming session data

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
