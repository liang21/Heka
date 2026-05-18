// tasks.md: T078 | spec.md: Text chunking with overlap
package parser

import (
	"strings"
	"unicode"

	"github.com/liang21/heka/internal/domain/rag"
)

// Chunker splits text into chunks with semantic overlap
type Chunker struct{}

// NewChunker creates a new chunker
func NewChunker() *Chunker {
	return &Chunker{}
}

// CountTokens estimates token count (rough approximation: ~4 chars per token)
func CountTokens(text string) int {
	return len(text) / 4
}

// Chunk splits text into chunks according to config
func (c *Chunker) Chunk(text string, config rag.ChunkConfig) []*rag.DocumentChunk {
	if text == "" {
		return nil
	}

	// Normalize whitespace
	text = strings.Join(strings.Fields(text), " ")

	// Split by paragraphs first
	paragraphs := strings.Split(text, "\n\n")

	var chunks []*rag.DocumentChunk
	var currentChunk strings.Builder
	var currentTokens int
	chunkIndex := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		paraTokens := CountTokens(para)

		// If single paragraph exceeds max, split by sentences
		if paraTokens > int(config.MaxTokens) {
			sentences := c.splitBySentences(para)
			for _, sentence := range sentences {
				sentenceTokens := CountTokens(sentence)
				newTokens := currentTokens + sentenceTokens

				if newTokens > int(config.MaxTokens) && currentTokens > 0 {
					// Save current chunk if large enough
					if currentTokens >= int(config.MinChunkSize) {
						chunks = append(chunks, &rag.DocumentChunk{
							Content: strings.TrimSpace(currentChunk.String()),
							Tokens:  currentTokens,
							Index:   chunkIndex,
						})
						chunkIndex++
					}

					// Start new chunk with overlap
					currentChunk.Reset()
					currentTokens = 0
					if config.Overlap > 0 {
						overlapText := c.getOverlapText(chunks, config.Overlap)
						currentChunk.WriteString(overlapText)
						currentTokens = CountTokens(overlapText)
					}
				}

				currentChunk.WriteString(sentence + " ")
				currentTokens += sentenceTokens
			}
		} else {
			// Check if adding paragraph would exceed max
			newTokens := currentTokens + paraTokens
			if newTokens > int(config.MaxTokens) && currentTokens > 0 {
				// Save current chunk
				if currentTokens >= int(config.MinChunkSize) {
					chunks = append(chunks, &rag.DocumentChunk{
						Content: strings.TrimSpace(currentChunk.String()),
						Tokens:  currentTokens,
						Index:   chunkIndex,
					})
					chunkIndex++
				}

				// Start new chunk with overlap
				currentChunk.Reset()
				currentTokens = 0
				if config.Overlap > 0 && len(chunks) > 0 {
					overlapText := c.getOverlapText(chunks, config.Overlap)
					currentChunk.WriteString(overlapText)
					currentTokens = CountTokens(overlapText)
				}
			}

			currentChunk.WriteString(para + "\n\n")
			currentTokens += paraTokens
		}
	}

	// Add final chunk
	// If we have no chunks yet, add the current content even if below MinChunkSize
	// This ensures we don't return empty results for small inputs
	if len(chunks) == 0 && currentTokens > 0 {
		chunks = append(chunks, &rag.DocumentChunk{
			Content: strings.TrimSpace(currentChunk.String()),
			Tokens:  currentTokens,
			Index:   chunkIndex,
		})
	} else if currentTokens >= int(config.MinChunkSize) {
		chunks = append(chunks, &rag.DocumentChunk{
			Content: strings.TrimSpace(currentChunk.String()),
			Tokens:  currentTokens,
			Index:   chunkIndex,
		})
	}

	return chunks
}

// splitBySentences splits text into sentences
func (c *Chunker) splitBySentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i, r := range runes {
		current.WriteRune(r)

		// Check for sentence ending
		if r == '.' || r == '!' || r == '?' {
			// Check if next char is space or end of string
			if i+1 >= len(runes) || unicode.IsSpace(runes[i+1]) {
				sent := strings.TrimSpace(current.String())
				if sent != "" {
					sentences = append(sentences, sent)
				}
				current.Reset()
			}
		}
	}

	// Add remaining text
	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		sentences = append(sentences, remaining)
	}

	return sentences
}

// getOverlapText extracts overlap from the end of the last chunk
func (c *Chunker) getOverlapText(chunks []*rag.DocumentChunk, overlapTokens int) string {
	if len(chunks) == 0 {
		return ""
	}

	lastChunk := chunks[len(chunks)-1].Content
	words := strings.Fields(lastChunk)

	if len(words) == 0 {
		return ""
	}

	// Estimate tokens in overlap (rough approximation)
	tokensPerWord := 4
	wordsToTake := (overlapTokens / tokensPerWord) + 1

	if wordsToTake > len(words) {
		wordsToTake = len(words)
	}

	overlapWords := words[len(words)-wordsToTake:]
	return strings.Join(overlapWords, " ") + " "
}
