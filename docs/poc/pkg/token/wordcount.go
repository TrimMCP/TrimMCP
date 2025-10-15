package token

import "strings"

// WordCountTokenizer implements Tokenizer using simple whitespace-based word counting.
// This is a language-agnostic approximation useful for quick estimates.
type WordCountTokenizer struct{}

// NewWordCount creates a new WordCountTokenizer.
func NewWordCount() *WordCountTokenizer {
	return &WordCountTokenizer{}
}

// Count returns the number of words (tokens) by splitting on whitespace.
// Empty strings return 0.
func (w *WordCountTokenizer) Count(text string) (int, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, nil
	}

	words := strings.Fields(text)
	return len(words), nil
}

// Name returns "word-count".
func (w *WordCountTokenizer) Name() string {
	return "word-count"
}
