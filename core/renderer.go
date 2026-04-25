package core

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// MarkdownRenderer wraps goldmark with full GFM support.
type MarkdownRenderer struct {
	md goldmark.Markdown
}

// NewMarkdownRenderer creates a new MarkdownRenderer.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Table,
				extension.Strikethrough,
				extension.Linkify,
				extension.TaskList,
				extension.DefinitionList,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
			),
		),
	}
}

// Render converts markdown to HTML.
func (r *MarkdownRenderer) Render(content string) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
