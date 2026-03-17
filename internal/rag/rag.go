package rag

// Stubs for embedder, retriever, reranker, generator interfaces

type Embedder interface {
	Embed(texts []string) ([][]float32, error)
}

type Retriever interface {
	Retrieve(query string, topK int) ([]any, error)
}

type Reranker interface {
	Rerank(items []any, query string, topK int) ([]any, error)
}

type Generator interface {
	GenerateAnswer(query string, context string) (string, error)
}
