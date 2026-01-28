package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	crack "github.com/taigrr/document-crack"
)

// Defaults
const (
	DefaultMaxChars    = 50000 // 50k chars default
	DefaultPageSize    = 10000 // 10k chars per page
	DefaultMaxDownload = 10    // 10MB default download limit
	MaxDownloadLimit   = 100   // 100MB hard limit
)

// Tool argument types with optional pagination/limits
type CrackFileArgs struct {
	Path     string `json:"path" jsonschema:"Path to the document file to extract text from"`
	MaxChars int    `json:"max_chars,omitempty" jsonschema:"Maximum characters to return (default 50000, 0 for unlimited)"`
	Page     int    `json:"page,omitempty" jsonschema:"Page number for paginated results (1-indexed, default 1)"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Characters per page (default 10000)"`
}

type CrackURLArgs struct {
	URL           string `json:"url" jsonschema:"URL of the document to download and extract text from"`
	MaxChars      int    `json:"max_chars,omitempty" jsonschema:"Maximum characters to return (default 50000, 0 for unlimited)"`
	Page          int    `json:"page,omitempty" jsonschema:"Page number for paginated results (1-indexed, default 1)"`
	PageSize      int    `json:"page_size,omitempty" jsonschema:"Characters per page (default 10000)"`
	MaxDownloadMB int    `json:"max_download_mb,omitempty" jsonschema:"Maximum download size in MB (default 10, max 100)"`
}

type CrackBase64Args struct {
	Data     string `json:"data" jsonschema:"Base64-encoded document content"`
	MaxChars int    `json:"max_chars,omitempty" jsonschema:"Maximum characters to return (default 50000, 0 for unlimited)"`
	Page     int    `json:"page,omitempty" jsonschema:"Page number for paginated results (1-indexed, default 1)"`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Characters per page (default 10000)"`
}

// paginationOpts holds pagination settings
type paginationOpts struct {
	maxChars int
	page     int
	pageSize int
}

func newPaginationOpts(maxChars, page, pageSize int) paginationOpts {
	opts := paginationOpts{
		maxChars: DefaultMaxChars,
		page:     1,
		pageSize: DefaultPageSize,
	}
	if maxChars > 0 {
		opts.maxChars = maxChars
	} else if maxChars == -1 {
		opts.maxChars = 0 // unlimited
	}
	if page > 0 {
		opts.page = page
	}
	if pageSize > 0 {
		opts.pageSize = pageSize
	}
	return opts
}

type Server struct {
	mcpServer *mcp.Server
}

func NewServer() *mcp.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "doc-mcp",
		Version: "1.1.0",
	}, nil)

	s := &Server{
		mcpServer: mcpServer,
	}

	s.setupTools()

	return mcpServer
}

func (s *Server) setupTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "crack_file",
		Description: "Extract text content from a local document file. Supports PDF, DOCX, DOC, PPTX, ODT, and TXT. Use pagination options to manage large documents.",
	}, s.crackFile)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "crack_url",
		Description: "Download a document from a URL and extract its text content. Supports PDF, DOCX, DOC, PPTX, ODT, and TXT. Use pagination and download limit options for large files.",
	}, s.crackURL)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "crack_base64",
		Description: "Extract text from base64-encoded document data. Supports PDF, DOCX, DOC, PPTX, ODT, and TXT. Use pagination options to manage large documents.",
	}, s.crackBase64)
}

func (s *Server) crackFile(ctx context.Context, req *mcp.CallToolRequest, args CrackFileArgs) (*mcp.CallToolResult, any, error) {
	if args.Path == "" {
		return errorResult("path is required"), nil, nil
	}

	doc, err := crack.FromFile(args.Path)
	if err != nil {
		return errorResult(fmt.Sprintf("Error extracting document: %v", err)), nil, nil
	}

	opts := newPaginationOpts(args.MaxChars, args.Page, args.PageSize)
	return formatResult(doc, opts), nil, nil
}

func (s *Server) crackURL(ctx context.Context, req *mcp.CallToolRequest, args CrackURLArgs) (*mcp.CallToolResult, any, error) {
	if args.URL == "" {
		return errorResult("url is required"), nil, nil
	}

	// Validate URL scheme
	if !strings.HasPrefix(args.URL, "http://") && !strings.HasPrefix(args.URL, "https://") {
		return errorResult("url must use http or https scheme"), nil, nil
	}

	// Determine download limit
	maxDownloadMB := DefaultMaxDownload
	if args.MaxDownloadMB > 0 {
		maxDownloadMB = args.MaxDownloadMB
		if maxDownloadMB > MaxDownloadLimit {
			maxDownloadMB = MaxDownloadLimit
		}
	}
	maxBytes := int64(maxDownloadMB) << 20

	// Download with limit
	data, err := downloadWithLimit(ctx, args.URL, maxBytes)
	if err != nil {
		return errorResult(fmt.Sprintf("Error downloading document: %v", err)), nil, nil
	}

	doc, err := crack.FromBytes(data)
	if err != nil {
		return errorResult(fmt.Sprintf("Error extracting document: %v", err)), nil, nil
	}

	opts := newPaginationOpts(args.MaxChars, args.Page, args.PageSize)
	return formatResult(doc, opts), nil, nil
}

func (s *Server) crackBase64(ctx context.Context, req *mcp.CallToolRequest, args CrackBase64Args) (*mcp.CallToolResult, any, error) {
	if args.Data == "" {
		return errorResult("data is required"), nil, nil
	}

	data, err := decodeBase64(args.Data)
	if err != nil {
		return errorResult(fmt.Sprintf("Error decoding base64: %v", err)), nil, nil
	}

	doc, err := crack.FromBytes(data)
	if err != nil {
		return errorResult(fmt.Sprintf("Error extracting document: %v", err)), nil, nil
	}

	opts := newPaginationOpts(args.MaxChars, args.Page, args.PageSize)
	return formatResult(doc, opts), nil, nil
}

func downloadWithLimit(ctx context.Context, url string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Check Content-Length header first
	if resp.ContentLength > maxBytes {
		return nil, fmt.Errorf("file too large: %d bytes (max %d MB)", resp.ContentLength, maxBytes>>20)
	}

	// Read with limit
	limitedReader := io.LimitReader(resp.Body, maxBytes+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("file too large (max %d MB)", maxBytes>>20)
	}

	return data, nil
}

// decodeBase64 tries multiple base64 encodings (standard, URL-safe, with/without padding)
func decodeBase64(s string) ([]byte, error) {
	if data, err := base64.StdEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	if data, err := base64.URLEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	if data, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	if data, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return data, nil
	}
	return nil, fmt.Errorf("invalid base64 encoding")
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Error: " + msg},
		},
		IsError: true,
	}
}

func formatResult(doc crack.Document, opts paginationOpts) *mcp.CallToolResult {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Document Type: %s\n", doc.Type))
	if doc.Title != "" {
		sb.WriteString(fmt.Sprintf("Title: %s\n", doc.Title))
	}
	sb.WriteString("\n--- Content ---\n")

	// Build full content first
	var contentBuilder strings.Builder
	for i, content := range doc.Content {
		if len(doc.Content) > 1 {
			contentBuilder.WriteString(fmt.Sprintf("\n[Page %d]\n", i+1))
		}
		contentBuilder.WriteString(content)
		contentBuilder.WriteString("\n")
	}
	fullContent := contentBuilder.String()
	totalChars := len(fullContent)

	// Apply pagination
	if opts.maxChars > 0 && opts.pageSize > 0 {
		totalPages := (totalChars + opts.pageSize - 1) / opts.pageSize
		if totalPages == 0 {
			totalPages = 1
		}

		page := opts.page
		if page < 1 {
			page = 1
		}
		if page > totalPages {
			page = totalPages
		}

		start := (page - 1) * opts.pageSize
		end := start + opts.pageSize
		if end > totalChars {
			end = totalChars
		}
		if start > totalChars {
			start = totalChars
		}

		pageContent := fullContent[start:end]

		// Truncate if still over maxChars
		if len(pageContent) > opts.maxChars {
			pageContent = pageContent[:opts.maxChars]
		}

		sb.WriteString(pageContent)

		// Add pagination info
		sb.WriteString(fmt.Sprintf("\n\n--- Page %d/%d (%d total chars) ---", page, totalPages, totalChars))
		if page < totalPages {
			sb.WriteString(fmt.Sprintf("\nUse page=%d to continue", page+1))
		}
	} else {
		// No pagination, but still respect maxChars
		content := fullContent
		if opts.maxChars > 0 && len(content) > opts.maxChars {
			content = content[:opts.maxChars]
			sb.WriteString(content)
			sb.WriteString(fmt.Sprintf("\n\n--- Truncated at %d chars (%d total) ---", opts.maxChars, totalChars))
			sb.WriteString("\nUse page/page_size options to paginate, or max_chars=-1 for unlimited")
		} else {
			sb.WriteString(content)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}
}
