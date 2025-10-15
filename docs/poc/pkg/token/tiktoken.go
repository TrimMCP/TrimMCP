package token

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

// TiktokenTokenizer implements Tokenizer using OpenAI's tiktoken cl100k_base encoding.
// This is the most accurate tokenization for GPT-4 and provides real-world cost estimates.
type TiktokenTokenizer struct {
	encoding *tiktoken.Tiktoken
}

// NewTiktoken creates a new TiktokenTokenizer with cl100k_base encoding.
// Returns error if encoding initialization fails.
func NewTiktoken() (*TiktokenTokenizer, error) {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tiktoken encoding: %w", err)
	}

	return &TiktokenTokenizer{
		encoding: enc,
	}, nil
}

// Count returns the number of tokens in the given text using cl100k_base encoding.
func (t *TiktokenTokenizer) Count(text string) (int, error) {
	if t.encoding == nil {
		return 0, fmt.Errorf("tiktoken encoding not initialized")
	}

	tokens := t.encoding.Encode(text, nil, nil)
	return len(tokens), nil
}

// Name returns "tiktoken-cl100k".
func (t *TiktokenTokenizer) Name() string {
	return "tiktoken-cl100k"
}
