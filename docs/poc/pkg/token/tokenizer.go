package token

// Tokenizer counts tokens in text using a specific tokenization method.
type Tokenizer interface {
	// Count returns the number of tokens in the given text.
	// Returns error if tokenization fails.
	Count(text string) (int, error)

	// Name returns the human-readable name of this tokenization method.
	// Examples: "tiktoken-cl100k", "word-count", "char-div-4"
	Name() string
}
