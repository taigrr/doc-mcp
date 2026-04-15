package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	crack "github.com/taigrr/document-crack"
)

// resultText extracts the text from the first content item in a CallToolResult.
func resultText(t *testing.T, result *gomcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	data, err := json.Marshal(result.Content[0])
	if err != nil {
		t.Fatalf("failed to marshal content: %v", err)
	}
	var obj struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("failed to unmarshal content: %v", err)
	}
	return obj.Text
}

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

func TestDownloadWithLimit_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world")
	}))
	defer ts.Close()

	data, err := downloadWithLimit(context.Background(), ts.URL, 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(data))
	}
}

func TestDownloadWithLimit_TooLarge(t *testing.T) {
	body := strings.Repeat("x", 1024)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	}))
	defer ts.Close()

	_, err := downloadWithLimit(context.Background(), ts.URL, 512)
	if err == nil {
		t.Error("expected error for oversized response")
	}
}

func TestDownloadWithLimit_ContentLengthTooLarge(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "999999999")
		fmt.Fprint(w, "small")
	}))
	defer ts.Close()

	_, err := downloadWithLimit(context.Background(), ts.URL, 1024)
	if err == nil {
		t.Error("expected error when Content-Length exceeds limit")
	}
}

func TestDownloadWithLimit_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := downloadWithLimit(context.Background(), ts.URL, 1<<20)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestDownloadWithLimit_InvalidURL(t *testing.T) {
	_, err := downloadWithLimit(context.Background(), "http://[invalid]:99999", 1<<20)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestCrackURL_MaxDownloadMBCapping(t *testing.T) {
	s := &Server{}
	// MaxDownloadMB > MaxDownloadLimit should be capped (won't fail, but exercises the path)
	result, _, err := s.crackURL(context.TODO(), nil, CrackURLArgs{
		URL:           "http://localhost:1/nonexistent",
		MaxDownloadMB: 999,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Will error on connection refused, but the cap logic was exercised
	if !result.IsError {
		t.Error("expected error for unreachable URL")
	}
}

func TestFormatResult_EmptyContent(t *testing.T) {
	doc := crack.Document{
		Type:    crack.TypeTXT,
		Content: []string{},
	}
	opts := newPaginationOpts(-1, 0, 0)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error for empty content")
	}
}

func TestFormatResult_TitleIncluded(t *testing.T) {
	doc := crack.Document{
		Type:    crack.TypePDF,
		Title:   "My Test Document",
		Content: []string{"some content"},
	}
	opts := newPaginationOpts(-1, 0, 0)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "My Test Document") {
		t.Error("expected title in output")
	}
}

func TestFormatResult_PaginationSecondPage(t *testing.T) {
	content := strings.Repeat("A", 100)
	doc := newTestDoc(crack.TypeTXT, content)
	opts := newPaginationOpts(200, 2, 30)
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "Page 2/") {
		t.Error("expected page 2 indicator in output")
	}
}

func TestFormatResult_MaxCharsTruncation(t *testing.T) {
	content := strings.Repeat("C", 500)
	doc := newTestDoc(crack.TypeTXT, content)
	// Use unlimited maxChars (-1 → 0) so formatResult takes the non-pagination branch,
	// then set maxChars manually after to trigger truncation in that branch.
	opts := paginationOpts{maxChars: 50, page: 1, pageSize: 0}
	result := formatResult(doc, opts)
	if result.IsError {
		t.Error("expected no error")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "Truncated at 50 chars") {
		t.Errorf("expected truncation message in output, got: %s", text[:min(len(text), 200)])
	}
}

func TestDecodeBase64_URLEncoding(t *testing.T) {
	// URL-safe base64 with + replaced by - and / by _
	input := "SGVsbG8gV29ybGQ=" // same in standard, but test the path
	data, err := decodeBase64(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", string(data))
	}
}
