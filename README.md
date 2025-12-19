# Confluence Exporter

Confluence Exporter is a CLI tool designed to export content from Confluence into Markdown format. This project provides a simple way to retrieve pages and their content from Confluence and convert them into a more portable format.

## Features

- Fetch pages from Confluence
- Convert Confluence content to Markdown
- Export to multiple formats:
  - **File**: Save as individual Markdown files
  - **Database**: Store in DuckDB database
  - **MeiliSearch**: Export as JSON with UIDs for MeiliSearch indexing
- Easy configuration through environment variables or config files

## Project Structure

```
confluence-exporter
├── cmd
│   └── exporter
│       └── main.go          # Entry point of the CLI application
├── internal
│   ├── api
│   │   └── confluence.go    # Functions to interact with the Confluence API
│   ├── converter
│   │   └── markdown.go      # Convert Confluence content to Markdown
│   ├── config
│   │   └── config.go        # Configuration settings for the application
│   └── models
│       └── page.go          # Data structures for Confluence pages
├── pkg
│   └── utils
│       ├── auth.go          # Utility functions for authentication
│       └── logger.go        # Logging functionality
├── go.mod                    # Module definition and dependencies
├── go.sum                    # Checksums for module dependencies
└── README.md                 # Project documentation
```

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/Ilhicas/confluence-exporter.git
   cd confluence-exporter
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Configuration

Create a `config.json` file based on `config-sample.json`:

```json
{
  "confluence": {
    "baseUrl": "https://your-domain.atlassian.net/wiki",
    "apiToken": "your-api-token",
    "username": "your-email@example.com"
  },
  "export": {
    "spaceKey": "TEAM",
    "outputDir": "./output",
    "outputType": "meilisearch",
    "recursive": true,
    "includeAttachments": false,
    "concurrentRequests": 5,
    "format": {
      "includeFrontMatter": true,
      "preserveLinks": true
    }
  },
  "logging": {
    "level": "info",
    "file": "confluence-export.log"
  }
}
```

### Output Types

- **`file`**: Exports pages as individual Markdown files in a directory structure
- **`db`**: Stores pages in a DuckDB database file (`confluence_pages.db`)
- **`meilisearch`**: Exports all pages as a single JSON file (`confluence_pages_meilisearch.json`) with UIDs for MeiliSearch indexing

## Usage

To run the application, use the following command:

```
go run cmd/exporter/main.go --config config.json
```

Replace `config.json` with the path to your configuration file containing the necessary API credentials.

## License

This project is licensed under the MIT License. See the LICENSE file for details.# confluence-exporter

## Disclaimer

The project is still in development and may not be fully functional. Use at your own risk.
This project was generated with AI assistance from Claude 3.7 Sonnet Thinking model.
