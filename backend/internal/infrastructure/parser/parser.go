package parser

import (
	"context"
	"fmt"
	"io"
)

type Parser interface {
	Parse(ctx context.Context, reader io.Reader) (string, error)
	SupportedTypes() []string
}

type Registry struct {
	parsers map[string]Parser
}

func NewRegistry() *Registry {
	r := &Registry{parsers: make(map[string]Parser)}

	// Register all parsers
	r.Register("pdf", &PDFParser{})
	r.Register("docx", &DOCXParser{})
	r.Register("xlsx", &XLSXParser{})
	r.Register("image", &ImageParser{})

	return r
}

func (r *Registry) Register(fileType string, p Parser) {
	r.parsers[fileType] = p
}

func (r *Registry) GetParser(fileType string) (Parser, error) {
	p, ok := r.parsers[fileType]
	if !ok {
		return nil, fmt.Errorf("no parser registered for file type: %s", fileType)
	}
	return p, nil
}
