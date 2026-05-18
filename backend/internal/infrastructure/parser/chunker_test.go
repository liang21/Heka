// tasks.md: T077 | spec.md: §3.5 Text chunking with overlap
package parser

import (
	"strings"
	"testing"

	"github.com/liang21/heka/internal/domain/rag"
	"github.com/stretchr/testify/assert"
)

func TestChunker_SemanticOverlap(t *testing.T) {
	t.Parallel()

	config := rag.DefaultChunkConfig()

	text := `This is paragraph one. It has multiple sentences.

This is paragraph two. It also has content.

This is paragraph three.`

	chunker := NewChunker()
	chunks := chunker.Chunk(text, config)

	// Should split by paragraphs first
	assert.True(t, len(chunks) >= 1)
	if len(chunks) > 0 {
		assert.Contains(t, chunks[0].Content, "paragraph one")
	}
	if len(chunks) > 1 {
		assert.Contains(t, chunks[1].Content, "paragraph two")
	}
}

func TestChunker_LongParagraphSplit(t *testing.T) {
	t.Parallel()

	config := rag.DefaultChunkConfig()

	// Create a long paragraph without newlines
	// Use actual text that will exceed token limit when processed
	var longText strings.Builder
	sentence := "This is a sentence that will be repeated many times to create a long paragraph. "
	for i := 0; i < 50; i++ {
		longText.WriteString(sentence)
	}

	chunker := NewChunker()
	chunks := chunker.Chunk(longText.String(), config)

	// Long paragraph should be split into multiple chunks
	assert.True(t, len(chunks) > 1, "Expected multiple chunks, got %d", len(chunks))
}

func TestChunker_Overlap(t *testing.T) {
	t.Parallel()

	config := rag.ChunkConfig{
		MaxTokens:    500,
		Overlap:      50,
		MinChunkSize: 100,
	}

	text := strings.Repeat("A", 100) + " " + strings.Repeat("B", 100)

	chunker := NewChunker()
	chunks := chunker.Chunk(text, config)

	// Verify chunks are created
	assert.True(t, len(chunks) >= 1)
	// Verify index is sequential
	for i := 0; i < len(chunks)-1; i++ {
		assert.Equal(t, chunks[i].Index+1, chunks[i+1].Index)
	}
}

func TestChunker_MinChunkSize(t *testing.T) {
	t.Parallel()

	config := rag.ChunkConfig{
		MaxTokens:    500,
		Overlap:      50,
		MinChunkSize: 100,
	}

	// Text that's too short
	shortText := "Short text"

	chunker := NewChunker()
	chunks := chunker.Chunk(shortText, config)

	// Small chunks should be discarded or merged
	assert.True(t, len(chunks) == 0 || len(chunks) == 1)
}

func TestChunker_EmptyInput(t *testing.T) {
	t.Parallel()

	config := rag.DefaultChunkConfig()
	chunker := NewChunker()

	chunks := chunker.Chunk("", config)

	assert.Empty(t, chunks)
}

func TestChunker_TokensCount(t *testing.T) {
	t.Parallel()

	config := rag.DefaultChunkConfig()

	text := "This is a test text for token counting."

	chunker := NewChunker()
	chunks := chunker.Chunk(text, config)

	// Verify token count respects MaxTokens
	for _, chunk := range chunks {
		assert.True(t, chunk.Tokens <= config.MaxTokens)
	}
}
