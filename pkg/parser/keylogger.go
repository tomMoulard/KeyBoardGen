package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/tommoulard/keyboardgen/pkg/genetic"
)

// KeyloggerParser handles parsing of keylogger data files.
type KeyloggerParser struct {
	charFreq    map[rune]int
	bigramFreq  map[string]int
	trigramFreq map[string]int
	totalChars  int
}

// NewKeyloggerParser creates a new parser instance.
func NewKeyloggerParser() *KeyloggerParser {
	return &KeyloggerParser{
		charFreq:    make(map[rune]int),
		bigramFreq:  make(map[string]int),
		trigramFreq: make(map[string]int),
		totalChars:  0,
	}
}

// ParseFormat defines supported keylogger formats.
type ParseFormat int

const (
	RawTextFormat ParseFormat = iota
	VimCommandFormat
	TimestampedFormat
	JSONFormat
)

// ParseConfig holds parsing configuration.
type ParseConfig struct {
	Format             ParseFormat `json:"format"`
	IgnoreCase         bool        `json:"ignore_case"`
	IgnoreNumbers      bool        `json:"ignore_numbers"`
	IgnoreSpecialChars bool        `json:"ignore_special_chars"`
	MinWordLength      int         `json:"min_word_length"`
	MaxLineLength      int         `json:"max_line_length"`
}

// DefaultConfig returns sensible parsing defaults.
func DefaultConfig() ParseConfig {
	return ParseConfig{
		Format:             RawTextFormat,
		IgnoreCase:         true,
		IgnoreNumbers:      true,
		IgnoreSpecialChars: true,
		MinWordLength:      1,
		MaxLineLength:      1000,
	}
}

// ProgrammingConfig returns configuration optimized for programming text.
func ProgrammingConfig() ParseConfig {
	return ParseConfig{
		Format:             RawTextFormat,
		IgnoreCase:         false, // Programming is case-sensitive
		IgnoreNumbers:      false, // Include numbers
		IgnoreSpecialChars: false, // Include special characters
		MinWordLength:      1,
		MaxLineLength:      1000,
	}
}

// FullKeyboardConfig returns configuration for full keyboard layout.
func FullKeyboardConfig() ParseConfig {
	return ParseConfig{
		Format:             RawTextFormat,
		IgnoreCase:         false,
		IgnoreNumbers:      false,
		IgnoreSpecialChars: false,
		MinWordLength:      1,
		MaxLineLength:      1000,
	}
}

// KeyloggerData holds parsed keylogger information.
type KeyloggerData struct {
	CharFrequency map[rune]int   `json:"char_frequency"`
	BigramFreq    map[string]int `json:"bigram_frequency"`
	TrigramFreq   map[string]int `json:"trigram_frequency"`
	TotalChars    int            `json:"total_chars"`
	WordCount     int            `json:"word_count"`
	LineCount     int            `json:"line_count"`
	Metadata      map[string]any `json:"metadata"`
}

// Parse processes keylogger data from a reader.
func (kp *KeyloggerParser) Parse(reader io.Reader, config ParseConfig) (*KeyloggerData, error) {
	scanner := bufio.NewScanner(reader)
	lineCount := 0
	wordCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Skip overly long lines
		if len(line) > config.MaxLineLength {
			continue
		}

		// Process line based on format
		processedText, err := kp.processLine(line, config)
		if err != nil {
			continue // Skip malformed lines
		}

		// Count words
		words := strings.Fields(processedText)
		wordCount += len(words)

		// Extract character and n-gram frequencies
		kp.extractFrequencies(processedText, config)
	}

	err := scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return &KeyloggerData{
		CharFrequency: kp.charFreq,
		BigramFreq:    kp.bigramFreq,
		TrigramFreq:   kp.trigramFreq,
		TotalChars:    kp.totalChars,
		WordCount:     wordCount,
		LineCount:     lineCount,
		Metadata: map[string]any{
			"unique_chars":    len(kp.charFreq),
			"unique_bigrams":  len(kp.bigramFreq),
			"unique_trigrams": len(kp.trigramFreq),
		},
	}, nil
}

// processLine processes a single line based on format.
func (kp *KeyloggerParser) processLine(line string, config ParseConfig) (string, error) {
	switch config.Format {
	case RawTextFormat:
		return kp.processRawText(line, config), nil
	case VimCommandFormat:
		return kp.processVimCommands(line, config), nil
	case TimestampedFormat:
		return kp.processTimestamped(line, config)
	case JSONFormat:
		return kp.processJSON(line, config)
	default:
		return kp.processRawText(line, config), nil
	}
}

// processRawText handles plain text input.
func (kp *KeyloggerParser) processRawText(line string, config ParseConfig) string {
	text := line

	if config.IgnoreCase {
		text = strings.ToLower(text)
	}

	// Filter characters
	var filtered strings.Builder

	for _, char := range text {
		if kp.shouldIncludeChar(char, config) {
			filtered.WriteRune(char)
		}
	}

	return filtered.String()
}

// processVimCommands handles vim-style command sequences.
func (kp *KeyloggerParser) processVimCommands(line string, config ParseConfig) string {
	// Remove vim command markers and escape sequences
	vimEscapePattern := regexp.MustCompile(`<[^>]+>`)
	cleaned := vimEscapePattern.ReplaceAllString(line, "")

	// Process remaining text
	return kp.processRawText(cleaned, config)
}

// processTimestamped handles timestamped keylogger format.
func (kp *KeyloggerParser) processTimestamped(line string, config ParseConfig) (string, error) {
	// Example format: "2023-01-01 12:00:00 | Hello World"
	timestampPattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\s*\|\s*(.*)$`)

	matches := timestampPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", errors.New("invalid timestamp format")
	}

	return kp.processRawText(matches[1], config), nil
}

// processJSON handles JSON-formatted keylogger data.
func (kp *KeyloggerParser) processJSON(line string, config ParseConfig) (string, error) {
	// Simplified JSON parsing - in real implementation, use json package
	// Example: {"timestamp": "...", "text": "Hello World"}
	jsonPattern := regexp.MustCompile(`"text":\s*"([^"]*)"`)

	matches := jsonPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", errors.New("no text field found in JSON")
	}

	return kp.processRawText(matches[1], config), nil
}

// shouldIncludeChar determines if a character should be included based on character set.
func (kp *KeyloggerParser) shouldIncludeChar(char rune, config ParseConfig) bool {
	// Get the character set for this configuration
	charset := genetic.FullKeyboardCharset()

	// Use character set to determine inclusion
	return charset.Contains(char)
}

// extractFrequencies extracts character and n-gram frequencies based on character set.
func (kp *KeyloggerParser) extractFrequencies(text string, config ParseConfig) {
	runes := []rune(text)
	charset := genetic.FullKeyboardCharset()

	// Extract character frequencies
	for _, char := range runes {
		if charset.Contains(char) {
			kp.charFreq[char]++
			kp.totalChars++
		}
	}

	// Extract bigram frequencies
	for i := range len(runes) - 1 {
		char1, char2 := runes[i], runes[i+1]
		if charset.Contains(char1) && charset.Contains(char2) {
			bigram := string([]rune{char1, char2})
			kp.bigramFreq[bigram]++
		}
	}

	// Extract trigram frequencies
	for i := range len(runes) - 2 {
		char1, char2, char3 := runes[i], runes[i+1], runes[i+2]
		if charset.Contains(char1) && charset.Contains(char2) && charset.Contains(char3) {
			trigram := string([]rune{char1, char2, char3})
			kp.trigramFreq[trigram]++
		}
	}
}

// GetCharFreq returns frequency of a character.
func (kd *KeyloggerData) GetCharFreq(char rune) int {
	return kd.CharFrequency[char]
}

// GetBigramFreq returns frequency of a bigram.
func (kd *KeyloggerData) GetBigramFreq(bigram string) int {
	return kd.BigramFreq[bigram]
}

// GetTrigramFreq returns frequency of a trigram.
func (kd *KeyloggerData) GetTrigramFreq(trigram string) int {
	return kd.TrigramFreq[trigram]
}

// GetTotalChars returns total character count.
func (kd *KeyloggerData) GetTotalChars() int {
	return kd.TotalChars
}

// GetAllBigrams returns all bigram frequencies.
func (kd *KeyloggerData) GetAllBigrams() map[string]int {
	return kd.BigramFreq
}

// GetMostFrequentChars returns the N most frequent characters.
func (kd *KeyloggerData) GetMostFrequentChars(n int) []struct {
	Char rune
	Freq int
} {
	type charFreq struct {
		Char rune
		Freq int
	}

	var pairs []charFreq
	for char, freq := range kd.CharFrequency {
		pairs = append(pairs, charFreq{char, freq})
	}

	// Sort by frequency (descending)
	for i := range len(pairs) - 1 {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].Freq > pairs[i].Freq {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	// Return top N
	if n > len(pairs) {
		n = len(pairs)
	}

	result := make([]struct {
		Char rune
		Freq int
	}, n)

	for i := range n {
		result[i] = struct {
			Char rune
			Freq int
		}{pairs[i].Char, pairs[i].Freq}
	}

	return result
}

// GetMostFrequentBigrams returns the N most frequent bigrams.
func (kd *KeyloggerData) GetMostFrequentBigrams(n int) []struct {
	Bigram string
	Freq   int
} {
	type bigramFreq struct {
		Bigram string
		Freq   int
	}

	var pairs []bigramFreq
	for bigram, freq := range kd.BigramFreq {
		pairs = append(pairs, bigramFreq{bigram, freq})
	}

	// Sort by frequency (descending)
	for i := range len(pairs) - 1 {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].Freq > pairs[i].Freq {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	// Return top N
	if n > len(pairs) {
		n = len(pairs)
	}

	result := make([]struct {
		Bigram string
		Freq   int
	}, n)

	for i := range n {
		result[i] = struct {
			Bigram string
			Freq   int
		}{pairs[i].Bigram, pairs[i].Freq}
	}

	return result
}

// Validate checks if the parsed data is sufficient for GA.
func (kd *KeyloggerData) Validate() error {
	if kd.TotalChars < 100 {
		return fmt.Errorf("insufficient data: only %d characters parsed, need at least 100", kd.TotalChars)
	}

	if len(kd.CharFrequency) < 10 {
		return fmt.Errorf("insufficient character diversity: only %d unique characters", len(kd.CharFrequency))
	}

	if len(kd.BigramFreq) < 20 {
		return fmt.Errorf("insufficient bigram data: only %d unique bigrams", len(kd.BigramFreq))
	}

	return nil
}
