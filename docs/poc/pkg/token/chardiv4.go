package token

import "unicode/utf8"

// CharDivFourTokenizer implements Tokenizer using character count divided by 4.
// This is a quick estimation method based on the rule-of-thumb that 1 token ≈ 4 characters
// for English text.
type CharDivFourTokenizer struct{}

// NewCharDivFour creates a new CharDivFourTokenizer.
func NewCharDivFour() *CharDivFourTokenizer {
	return &CharDivFourTokenizer{}
}

// Count returns the number of tokens estimated as character_count / 4.
// Uses UTF-8 character counting (not byte counting).
func (c *CharDivFourTokenizer) Count(text string) (int, error) {
	charCount := utf8.RuneCountInString(text)
	tokenCount := (charCount + 3) / 4 // Round up using integer division
	return tokenCount, nil
}

// Name returns "char-div-4".
func (c *CharDivFourTokenizer) Name() string {
	return "char-div-4"
}
