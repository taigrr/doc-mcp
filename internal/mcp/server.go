package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	crack "github.com/taigrr/document-crack"
)

// Tool argument types
type CrackFileArgs struct {
	Path string `json:"path" jsonschema:"Path to the document file to extract text from"`
}

type CrackURLArgs struct {
	URL string `json:"url" jsonschema:"URL of the document to download and extract text from"`
}

type CrackBase64Args struct {
	Data string `json:"data" jsonschema:"Base64-encoded document content"`
}

type Server struct {
	mcpServer *mcp.Server
}

func NewServer() *mcp.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "doc-mcp",
		Version: "1.0.0",
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
		Description: "Extract text content from a local document file. Supports PDF, DOCX, DOC, PPTX, ODT, and TXT.",
	}, s.crackFile)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "crack_url",
		Description: "Download a document from a URL and extract its text content. Supports PDF, DOCX, DOC, PPTX, ODT, and TXT.",
	}, s.crackURL)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "crack_base64",
		Description: "Extract text from base64-encoded document data. Supports PDF, DOCX, DOC, PPTX, ODT, and TXT.",
	}, s.crackBase64)
}

func (s *Server) crackFile(ctx context.Context, req *mcp.CallToolRequest, args CrackFileArgs) (*mcp.CallToolResult, any, error) {
	if args.Path == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: path is required"},
			},
			IsError: true,
		}, nil, nil
	}

	doc, err := crack.FromFile(args.Path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error extracting document: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return formatResult(doc), nil, nil
}

func (s *Server) crackURL(ctx context.Context, req *mcp.CallToolRequest, args CrackURLArgs) (*mcp.CallToolResult, any, error) {
	if args.URL == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: url is required"},
			},
			IsError: true,
		}, nil, nil
	}

	doc, err := crack.FromURL(ctx, args.URL)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error downloading/extracting document: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return formatResult(doc), nil, nil
}

func (s *Server) crackBase64(ctx context.Context, req *mcp.CallToolRequest, args CrackBase64Args) (*mcp.CallToolResult, any, error) {
	if args.Data == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: data is required"},
			},
			IsError: true,
		}, nil, nil
	}

	data, err := base64.StdEncoding.DecodeString(args.Data)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error decoding base64: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	doc, err := crack.FromBytes(data)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error extracting document: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return formatResult(doc), nil, nil
}

func formatResult(doc crack.Document) *mcp.CallToolResult {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Document Type: %s\n", doc.Type))
	if doc.Title != "" {
		sb.WriteString(fmt.Sprintf("Title: %s\n", doc.Title))
	}
	sb.WriteString("\n--- Content ---\n")
	for i, content := range doc.Content {
		if len(doc.Content) > 1 {
			sb.WriteString(fmt.Sprintf("\n[Page %d]\n", i+1))
		}
		sb.WriteString(content)
		sb.WriteString("\n")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}
}
