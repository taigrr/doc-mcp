package mcp

import (
	"context"
	"testing"

	crack "github.com/taigrr/document-crack"
)

func newTestDoc(docType crack.FileType, content string) crack.Document {
	return crack.Document{
		Type:    docType,
		Content: []string{content},
	}
}

func TestNewPaginationOpts_Defaults(t *testing.T) {
	opts := newPaginationOpts(0, 0, 0)
	if opts.maxChars != DefaultMaxChars {
		t.Errorf("expected maxChars=%d, got %d", DefaultMaxChars, opts.maxChars)
	}
	if opts.page != 1 {
		t.Errorf("expected page=1, got %d", opts.page)
	}
	if opts.pageSize != DefaultPageSize {
		t.Errorf("expected pageSize=%d, got %d", DefaultPageSize, opts.pageSize)
	}
}

func TestNewPaginationOpts_Custom(t *testing.T) {
	opts := newPaginationOpts(1000, 3, 500)
	if opts.maxChars != 1000 {
		t.Errorf("expected maxChars=1000, got %d", opts.maxChars)
	}
	if opts.page != 3 {
		t.Errorf("expected page=3, got %d", opts.page)
	}
	if opts.pageSize != 500 {
		t.Errorf("expected pageSize=500, got %d", opts.pageSize)
	}
}

func TestNewPaginationOpts_Unlimited(t *testing.T) {
	opts := newPaginationOpts(-1, 1, 0)
	if opts.maxChars != 0 {
		t.Errorf("expected maxChars=0 (unlimited), got %d", opts.maxChars)
	}
}

func TestDecodeBase64_Standard(t *testing.T) {
	input := "SGVsbG8gV29ybGQ="
	data, err := decodeBase64(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", string(data))
	}
}

func TestDecodeBase64_RawStd(t *testing.T) {
	input := "SGVsbG8"
	data, err := decodeBase64(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "Hello" {
		t.Errorf("expected 'Hello', got %q", string(data))
	}
}

func TestDecodeBase64_Invalid(t *testing.T) {
	_, err := decodeBase64("!!!not-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestErrorResult(t *testing.T) {
	result := errorResult("test error")
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
}

func TestFormatResult_NoPagination(t *testing.T) {
	doc := newTestDoc(crack.TypeTXT, "Hello World")
	opts := newPaginationOpts(-1, 0, 0)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
}

func TestFormatResult_Pagination(t *testing.T) {
	content := make([]byte, 100)
	for i := range content {
		content[i] = 'A'
	}
	doc := newTestDoc(crack.TypeTXT, string(content))
	opts := newPaginationOpts(200, 1, 30)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error")
	}
}

func TestFormatResult_Truncation(t *testing.T) {
	content := make([]byte, 200)
	for i := range content {
		content[i] = 'B'
	}
	doc := newTestDoc(crack.TypeTXT, string(content))
	opts := newPaginationOpts(50, 0, 0)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error")
	}
}

func TestCrackFile_EmptyPath(t *testing.T) {
	s := &Server{}
	result, _, err := s.crackFile(context.TODO(), nil, CrackFileArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for empty path")
	}
}

func TestCrackFile_NonexistentFile(t *testing.T) {
	s := &Server{}
	result, _, err := s.crackFile(context.TODO(), nil, CrackFileArgs{Path: "/nonexistent/file.pdf"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}

func TestCrackURL_EmptyURL(t *testing.T) {
	s := &Server{}
	result, _, err := s.crackURL(context.TODO(), nil, CrackURLArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for empty URL")
	}
}

func TestCrackURL_InvalidScheme(t *testing.T) {
	s := &Server{}
	result, _, err := s.crackURL(context.TODO(), nil, CrackURLArgs{URL: "ftp://example.com/file.pdf"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid scheme")
	}
}

func TestCrackBase64_EmptyData(t *testing.T) {
	s := &Server{}
	result, _, err := s.crackBase64(context.TODO(), nil, CrackBase64Args{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for empty data")
	}
}

func TestCrackBase64_InvalidBase64(t *testing.T) {
	s := &Server{}
	result, _, err := s.crackBase64(context.TODO(), nil, CrackBase64Args{Data: "!!!invalid!!!"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid base64")
	}
}

func TestNewServer(t *testing.T) {
	server := NewServer()
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestFormatResult_MultiPage(t *testing.T) {
	doc := crack.Document{
		Type:  "pdf",
		Title: "Test Doc",
		Content: []string{
			"Page one content here",
			"Page two content here",
		},
	}
	opts := newPaginationOpts(-1, 0, 0)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error")
	}
}

func TestFormatResult_PageBeyondRange(t *testing.T) {
	doc := newTestDoc(crack.TypeTXT, "short")
	opts := newPaginationOpts(100, 999, 10)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error even for out-of-range page")
	}
}
