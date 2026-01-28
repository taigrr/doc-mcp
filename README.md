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

### With Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "doc-mcp": {
      "command": "doc-mcp"
    }
  }
}
```

### With Crush

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

### `crack_file`

Extract text content from a local file.

**Input:**
- `path` (string, required): Path to the document file

**Output:**
- Document type, title (if available), and extracted text content

### `crack_url`

Download and extract text content from a URL.

**Input:**
- `url` (string, required): URL of the document to download

**Output:**
- Document type, title (if available), and extracted text content

### `crack_base64`

Extract text from base64-encoded document data.

**Input:**
- `data` (string, required): Base64-encoded document content

**Output:**
- Document type, title (if available), and extracted text content

## License

0BSD
