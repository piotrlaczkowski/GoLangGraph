# RAG System Example

This example demonstrates how to build a **Retrieval-Augmented Generation (RAG)** system using GoLangGraph. The RAG system combines document retrieval with LLM generation to provide accurate, context-aware responses based on your own documents.

## 🎯 What You'll Learn

- **Document Processing**: Load and chunk documents for optimal retrieval
- **Vector Embeddings**: Convert text to vector representations for semantic search
- **Retrieval System**: Find relevant documents based on query similarity
- **Context Integration**: Combine retrieved context with LLM prompts
- **RAG Pipeline**: End-to-end question-answering with your documents

## 🏗️ Architecture

```
Query → Embedding → Vector Search → Context Retrieval → LLM Generation → Response
```

## 🚀 Features

- **Multiple Document Formats**: Support for text, markdown, and structured data
- **Intelligent Chunking**: Smart text segmentation with overlap
- **Semantic Search**: Vector-based similarity matching
- **Context Ranking**: Relevance scoring and filtering
- **Source Attribution**: Track which documents provided answers
- **Streaming Responses**: Real-time answer generation

## 📋 Prerequisites

1. **Ollama Installation**:
   ```bash
   # Install Ollama
   curl -fsSL https://ollama.com/install.sh | sh
   
   # Pull required models
   ollama pull gemma3:1b                    # Main LLM
   ollama pull nomic-embed-text:latest      # Embedding model
   ```

2. **Sample Documents**: Place your documents in the `documents/` folder

## 🔧 Configuration

The RAG system supports multiple configurations:

### Basic Configuration
- **Chunk Size**: 500 tokens with 50 token overlap
- **Retrieval Count**: Top 3 most relevant chunks
- **Similarity Threshold**: 0.7 minimum relevance score

### Advanced Configuration
- **Hybrid Search**: Combines vector and keyword search
- **Reranking**: Secondary relevance scoring
- **Context Compression**: Removes redundant information

## 💻 Usage

### Basic RAG Query
```bash
cd examples/04-rag-system
go run main.go

# Interactive mode
> What is the main topic of the documents?
> Explain the key concepts mentioned in section 2
> Find information about implementation details
```

### Command Line Options
```bash
# Process specific documents
go run main.go --docs ./path/to/documents

# Configure chunk size
go run main.go --chunk-size 1000

# Set retrieval count
go run main.go --top-k 5

# Enable debug mode
go run main.go --debug
```

## 📁 Project Structure

```
04-rag-system/
├── main.go              # Main RAG application
├── config.go            # RAG configuration
├── document_processor.go # Document loading and chunking
├── embeddings.go        # Vector embedding generation
├── retriever.go         # Document retrieval logic
├── vector_store.go      # In-memory vector storage
├── documents/           # Sample documents
│   ├── sample1.txt
│   ├── sample2.md
│   └── technical_doc.txt
└── README.md           # This file
```

## 🔍 Example Interactions

### Document Analysis
```
You: What are the main topics covered in the documents?

RAG System: Based on the uploaded documents, the main topics include:
1. Software Architecture Patterns (from technical_doc.txt)
2. Machine Learning Fundamentals (from sample1.txt)  
3. Best Practices for API Design (from sample2.md)

The documents cover both theoretical concepts and practical implementation guidelines.

Sources: technical_doc.txt (lines 1-50), sample1.txt (lines 25-75)
```

### Specific Information Retrieval
```
You: How do I implement a microservices architecture?

RAG System: According to the technical documentation, implementing microservices architecture involves:

1. **Service Decomposition**: Break down monolithic applications into smaller, independent services
2. **API Gateway**: Use a centralized entry point for client requests
3. **Service Discovery**: Implement mechanisms for services to find each other
4. **Data Management**: Each service should own its data and database

Key considerations include fault tolerance, monitoring, and deployment strategies.

Sources: technical_doc.txt (lines 120-180), sample2.md (lines 45-90)
```

## ⚙️ Advanced Features

### 1. Custom Embedding Models
```go
// Use different embedding models
config := &RAGConfig{
    EmbeddingModel: "all-MiniLM-L6-v2",
    EmbeddingDim:   384,
}
```

### 2. Hybrid Search
```go
// Combine vector and keyword search
retriever := NewHybridRetriever(
    vectorStore,
    keywordIndex,
    0.7, // vector weight
    0.3, // keyword weight
)
```

### 3. Context Compression
```go
// Compress retrieved context
compressor := NewContextCompressor()
compressedContext := compressor.Compress(retrievedChunks)
```

## 📊 Performance Metrics

The system tracks several metrics:

- **Retrieval Latency**: Time to find relevant documents
- **Embedding Generation**: Time to create vector representations
- **Context Relevance**: Similarity scores of retrieved chunks
- **Response Quality**: Coherence and accuracy of generated answers
- **Source Coverage**: Percentage of documents utilized

## 🛠️ Customization Options

### Document Processors
- **TextProcessor**: Plain text files
- **MarkdownProcessor**: Markdown with structure preservation
- **PDFProcessor**: PDF extraction (requires additional dependencies)
- **JSONProcessor**: Structured data processing

### Chunking Strategies
- **FixedSizeChunker**: Fixed token count chunks
- **SentenceChunker**: Sentence boundary preservation
- **SemanticChunker**: Topic-based segmentation
- **HierarchicalChunker**: Document structure awareness

### Retrieval Methods
- **VectorRetriever**: Pure semantic search
- **BM25Retriever**: Keyword-based search
- **HybridRetriever**: Combined approach
- **ReRankingRetriever**: Two-stage retrieval

## 🐛 Troubleshooting

### Common Issues

1. **No Documents Found**
   ```
   Error: No documents found in directory
   Solution: Ensure documents are in the correct folder and format
   ```

2. **Embedding Model Not Available**
   ```
   Error: Failed to load embedding model
   Solution: Pull the embedding model with: ollama pull nomic-embed-text
   ```

3. **Low Relevance Scores**
   ```
   Issue: Retrieved documents seem irrelevant
   Solution: Lower similarity threshold or improve document quality
   ```

4. **Memory Issues**
   ```
   Issue: Out of memory with large document sets
   Solution: Reduce chunk size or implement disk-based vector store
   ```

### Performance Optimization

1. **Batch Processing**: Process multiple documents simultaneously
2. **Caching**: Cache embeddings for frequently accessed chunks
3. **Indexing**: Use approximate nearest neighbor search for large datasets
4. **Pruning**: Remove low-quality or duplicate chunks

## 🔗 Integration Examples

### With Persistence
```go
// Save vector store to disk
vectorStore.SaveToDisk("./vector_index.db")

// Load from disk
vectorStore.LoadFromDisk("./vector_index.db")
```

### With External APIs
```go
// Use external embedding service
embeddingService := NewOpenAIEmbeddings(apiKey)
ragSystem.SetEmbeddingService(embeddingService)
```

### With Multiple Models
```go
// Use different models for different tasks
ragConfig := &RAGConfig{
    EmbeddingModel: "nomic-embed-text",
    GenerationModel: "gemma3:1b",
    RerankingModel: "ms-marco-MiniLM-L-12-v2",
}
```

## 📚 Learning Resources

- **Vector Databases**: Understanding similarity search
- **Text Embeddings**: How semantic representations work
- **Information Retrieval**: Classical and modern approaches
- **Prompt Engineering**: Optimizing context integration
- **RAG Evaluation**: Measuring system performance

## 🚀 Next Steps

After mastering this example:
1. Explore **05-streaming** for real-time responses
2. Try **06-persistence** for production-ready storage
3. Check **07-tools-integration** for enhanced capabilities
4. Review **08-production-ready** for deployment strategies

## 🤝 Contributing

Improve this example by:
- Adding new document processors
- Implementing better chunking strategies
- Contributing evaluation metrics
- Sharing performance optimizations

---

**Happy Building!** 🎉

This RAG system provides a solid foundation for building intelligent document-based question-answering systems with GoLangGraph. 