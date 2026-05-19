package rag

import (
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T117 | spec.md: §4.11 RAG 检索 DTO

type SearchRequest struct {
	ProjectID shared.ID `json:"project_id" validate:"required,uuid"`
	Query     string    `json:"query" validate:"required"`
	TopK      int       `json:"top_k"`
}

type SearchResponse struct {
	Results []SearchResultItem `json:"results"`
}

type SearchResultItem struct {
	ChunkID  shared.ID `json:"chunk_id"`
	Content  string    `json:"content"`
	Score    float32   `json:"score"`
	FileName string    `json:"file_name"`
}
