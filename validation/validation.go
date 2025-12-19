package validation

import (
	"regexp"
	"strings"
	"unicode"
)

// IsValidPrompt checks if a prompt makes sense (not gibberish)
// Returns true if the prompt appears to be valid, false if it's likely gibberish
func IsValidPrompt(prompt string) bool {
	// Trim whitespace
	trimmed := strings.TrimSpace(prompt)
	
	// Check minimum length (at least 3 characters)
	if len(trimmed) < 3 {
		return false
	}
	
	// Check maximum reasonable length (prevent extremely long gibberish)
	if len(trimmed) > 10000 {
		return false
	}
	
	// Check if it's all whitespace
	if len(trimmed) == 0 {
		return false
	}
	
	// Check for minimum word count (at least 2 words for a meaningful prompt)
	words := strings.Fields(trimmed)
	if len(words) < 2 {
		// Single word might be valid if it's long enough and has meaning
		if len(words) == 1 && len(words[0]) >= 3 {
			// Check if it's not just repeated characters
			if !isRepeatedCharacters(words[0]) {
				return true
			}
		}
		return false
	}
	
	// Check for excessive character repetition (e.g., "aaaaaa", "111111")
	if hasExcessiveRepetition(trimmed) {
		return false
	}
	
	// Check for too many special characters (more than 50% special chars is suspicious)
	if hasTooManySpecialChars(trimmed) {
		return false
	}
	
	// Check for valid sentence structure
	// Should have some letters (at least 30% of characters should be letters)
	letterCount := 0
	totalChars := 0
	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			letterCount++
		}
		if !unicode.IsSpace(r) {
			totalChars++
		}
	}
	
	if totalChars == 0 {
		return false
	}
	
	letterRatio := float64(letterCount) / float64(totalChars)
	if letterRatio < 0.3 {
		return false
	}
	
	// Check for common patterns that indicate gibberish
	if isGibberishPattern(trimmed) {
		return false
	}
	
	// Check for valid word patterns (words should have reasonable length)
	// Too many very short words (1-2 chars) or very long words (>30 chars) might indicate gibberish
	shortWordCount := 0
	longWordCount := 0
	for _, word := range words {
		cleanWord := strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(cleanWord) <= 2 && len(cleanWord) > 0 {
			shortWordCount++
		}
		if len(cleanWord) > 30 {
			longWordCount++
		}
	}
	
	// If more than 70% are very short words, it's suspicious
	if len(words) > 0 && float64(shortWordCount)/float64(len(words)) > 0.7 {
		return false
	}
	
	// If too many extremely long words, it's suspicious
	if len(words) > 0 && float64(longWordCount)/float64(len(words)) > 0.3 {
		return false
	}
	
	// Check for keyboard mashing patterns (e.g., "asdfgh", "qwerty", "zxcvbn")
	if hasKeyboardMashing(trimmed) {
		return false
	}
	
	// Check for excessive numbers (more than 50% numbers is suspicious)
	digitCount := 0
	for _, r := range trimmed {
		if unicode.IsDigit(r) {
			digitCount++
		}
	}
	if totalChars > 0 && float64(digitCount)/float64(totalChars) > 0.5 {
		return false
	}
	
	// Check for valid punctuation usage
	// Should have reasonable punctuation (not excessive)
	if hasExcessivePunctuation(trimmed) {
		return false
	}
	
	// Check for common English words or patterns
	// If it contains some common words, it's more likely to be valid
	if hasCommonWords(trimmed) {
		return true
	}
	
	// Check for question patterns (questions are usually valid)
	if isQuestion(trimmed) {
		return true
	}
	
	// Check for imperative patterns (commands are usually valid)
	if isImperative(trimmed) {
		return true
	}
	
	// If it passes all the negative checks and has reasonable structure, consider it valid
	// This is a lenient check - we'd rather process a slightly odd prompt than reject a valid one
	return true
}

// isRepeatedCharacters checks if a string is just repeated characters
func isRepeatedCharacters(s string) bool {
	if len(s) < 3 {
		return false
	}
	firstChar := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != firstChar {
			return false
		}
	}
	return true
}

// hasExcessiveRepetition checks for patterns like "aaaa", "1111", "ababab"
func hasExcessiveRepetition(s string) bool {
	// Check for 4+ consecutive identical characters using a simpler approach
	// Instead of backreferences, check character by character
	if len(s) < 4 {
		return false
	}
	
	// Check for 4+ consecutive identical characters
	for i := 0; i <= len(s)-4; i++ {
		char := s[i]
		count := 1
		for j := i + 1; j < len(s) && s[j] == char; j++ {
			count++
		}
		if count >= 4 {
			return true
		}
	}
	
	// Check for simple repeating patterns (2-3 char patterns repeated 4+ times)
	// Check 2-char patterns
	for i := 0; i <= len(s)-8; i++ {
		if len(s)-i >= 8 {
			pattern := s[i : i+2]
			repeats := 1
			for j := i + 2; j <= len(s)-2 && s[j:j+2] == pattern; j += 2 {
				repeats++
			}
			if repeats >= 4 {
				return true
			}
		}
	}
	
	// Check 3-char patterns
	for i := 0; i <= len(s)-12; i++ {
		if len(s)-i >= 12 {
			pattern := s[i : i+3]
			repeats := 1
			for j := i + 3; j <= len(s)-3 && s[j:j+3] == pattern; j += 3 {
				repeats++
			}
			if repeats >= 4 {
				return true
			}
		}
	}
	
	return false
}

// hasTooManySpecialChars checks if more than 50% of characters are special
func hasTooManySpecialChars(s string) bool {
	specialCount := 0
	totalNonSpace := 0
	
	for _, r := range s {
		if !unicode.IsSpace(r) {
			totalNonSpace++
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsPunct(r) {
				specialCount++
			} else if unicode.IsPunct(r) {
				// Count excessive punctuation as special
				specialCount++
			}
		}
	}
	
	if totalNonSpace == 0 {
		return false
	}
	
	ratio := float64(specialCount) / float64(totalNonSpace)
	return ratio > 0.5
}

// isGibberishPattern checks for common gibberish patterns
func isGibberishPattern(s string) bool {
	lower := strings.ToLower(s)
	
	// Check for patterns like "asdf", "qwerty", "zxcv" (keyboard patterns)
	keyboardPatterns := []string{
		"asdf", "qwerty", "zxcv", "hjkl", "fghj", "dfgh",
		"asdfgh", "qwertyui", "zxcvbnm",
	}
	
	for _, pattern := range keyboardPatterns {
		if strings.Contains(lower, pattern) && len(s) < 20 {
			// If the string is mostly this pattern, it's gibberish
			if strings.Count(lower, pattern)*len(pattern) > len(lower)/2 {
				return true
			}
		}
	}
	
	return false
}

// hasKeyboardMashing checks for keyboard mashing patterns
func hasKeyboardMashing(s string) bool {
	lower := strings.ToLower(s)
	
	// Common keyboard mashing sequences
	mashingPatterns := []string{
		"asdfghjkl", "qwertyuiop", "zxcvbnm",
		"asdf", "qwer", "zxcv", "hjkl",
	}
	
	for _, pattern := range mashingPatterns {
		if strings.Contains(lower, pattern) {
			// If the string is short and contains these patterns, likely mashing
			if len(s) < 30 {
				return true
			}
		}
	}
	
	return false
}

// hasExcessivePunctuation checks for too much punctuation
func hasExcessivePunctuation(s string) bool {
	punctuationCount := 0
	totalChars := 0
	
	for _, r := range s {
		if !unicode.IsSpace(r) {
			totalChars++
			if unicode.IsPunct(r) {
				punctuationCount++
			}
		}
	}
	
	if totalChars == 0 {
		return false
	}
	
	// More than 30% punctuation is excessive
	return float64(punctuationCount)/float64(totalChars) > 0.3
}

// hasCommonWords checks if the prompt contains common English words
func hasCommonWords(s string) bool {
	lower := strings.ToLower(s)
	
	// Common English words that indicate meaningful text
	commonWords := []string{
		"the", "is", "are", "was", "were", "be", "been", "have", "has", "had",
		"do", "does", "did", "will", "would", "could", "should", "may", "might",
		"can", "this", "that", "these", "those", "what", "which", "who", "when",
		"where", "why", "how", "i", "you", "he", "she", "it", "we", "they",
		"and", "or", "but", "if", "then", "because", "so", "with", "from",
		"for", "to", "of", "in", "on", "at", "by", "about", "into", "through",
		"want", "need", "get", "make", "give", "take", "go", "come", "see",
		"know", "think", "say", "tell", "ask", "help", "show", "find", "use",
	}
	
	for _, word := range commonWords {
		// Use word boundaries to match whole words
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
		if re.MatchString(lower) {
			return true
		}
	}
	
	return false
}

// isQuestion checks if the prompt is a question
func isQuestion(s string) bool {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) == 0 {
		return false
	}
	
	// Check for question mark
	if strings.HasSuffix(trimmed, "?") {
		return true
	}
	
	// Check for question words at the start
	lower := strings.ToLower(trimmed)
	questionWords := []string{"what", "who", "when", "where", "why", "how", "which", "whose"}
	for _, qw := range questionWords {
		if strings.HasPrefix(lower, qw+" ") || strings.HasPrefix(lower, qw+"?") {
			return true
		}
	}
	
	// Check for "is", "are", "can", "do", "does", "did" at start (common question patterns)
	questionStarters := []string{"is ", "are ", "can ", "do ", "does ", "did ", "will ", "would ", "could ", "should "}
	for _, qs := range questionStarters {
		if strings.HasPrefix(lower, qs) {
			return true
		}
	}
	
	return false
}

// isImperative checks if the prompt is an imperative (command)
func isImperative(s string) bool {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) == 0 {
		return false
	}
	
	lower := strings.ToLower(trimmed)
	
	// Common imperative verbs
	imperativeVerbs := []string{
		"show", "display", "list", "get", "give", "tell", "explain", "describe",
		"create", "make", "generate", "build", "write", "send", "find", "search",
		"help", "assist", "provide", "calculate", "compute", "analyze",
	}
	
	words := strings.Fields(lower)
	if len(words) == 0 {
		return false
	}
	
	firstWord := words[0]
	for _, verb := range imperativeVerbs {
		if firstWord == verb {
			return true
		}
	}
	
	return false
}
