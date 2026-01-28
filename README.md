# doc-mcp

An MCP server for extracting text content from documents.

## Supported Formats

- **PDF** - Portable Document Format
- **DOCX** - Microsoft Word (Open XML)
- **DOC** - Microsoft Word (Legacy)
- **PPTX** - Microsoft PowerPoint (Open XML)
- **ODT** - OpenDocument Text
- **TXT** - Plain text

## Installation

```bash
go install github.com/taigrr/doc-mcp@latest
```

## Usage

### With [Crush](https://github.com/charmbracelet/crush)

Add to your Crush config:

```json
{
  "mcp": {
    "doc-mcp": {
      "command": "doc-mcp"
    }
  }
}
```

## Tools

All tools support pagination to manage large documents and avoid filling context.

### `crack_file`

Extract text content from a local file.

**Input:**
- `path` (string, required): Path to the document file
- `max_chars` (int, optional): Maximum characters to return (default 50000, use -1 for unlimited)
- `page` (int, optional): Page number for paginated results (1-indexed, default 1)
- `page_size` (int, optional): Characters per page (default 10000)

### `crack_url`

Download and extract text content from a URL.

**Input:**
- `url` (string, required): URL of the document to download (http/https only)
- `max_chars` (int, optional): Maximum characters to return (default 50000, use -1 for unlimited)
- `page` (int, optional): Page number for paginated results (1-indexed, default 1)
- `page_size` (int, optional): Characters per page (default 10000)
- `max_download_mb` (int, optional): Maximum download size in MB (default 10, max 100)

### `crack_base64`

Extract text from base64-encoded document data.

**Input:**
- `data` (string, required): Base64-encoded document content (standard or URL-safe)
- `max_chars` (int, optional): Maximum characters to return (default 50000, use -1 for unlimited)
- `page` (int, optional): Page number for paginated results (1-indexed, default 1)
- `page_size` (int, optional): Characters per page (default 10000)

## Defaults

| Setting | Default | Notes |
|---------|---------|-------|
| `max_chars` | 50,000 | Use -1 for unlimited |
| `page_size` | 10,000 | Characters per page |
| `max_download_mb` | 10 MB | Hard limit: 100 MB |

## License

0BSD
