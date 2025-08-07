package genetic

// CharacterSet defines the characters supported by keyboard layouts.
type CharacterSet struct {
	Characters []rune       `json:"characters"`
	CharToPos  map[rune]int `json:"char_to_pos"`
	PosToChar  map[int]rune `json:"pos_to_char"`
	Size       int          `json:"size"`
	Name       string       `json:"name"`
}

// NewCharacterSet creates a new character set from a slice of runes.
func NewCharacterSet(name string, chars []rune) *CharacterSet {
	cs := &CharacterSet{
		Characters: make([]rune, len(chars)),
		CharToPos:  make(map[rune]int),
		PosToChar:  make(map[int]rune),
		Size:       len(chars),
		Name:       name,
	}

	copy(cs.Characters, chars)

	for i, char := range chars {
		cs.CharToPos[char] = i
		cs.PosToChar[i] = char
	}

	return cs
}

// Contains checks if a character is in this character set.
func (cs *CharacterSet) Contains(char rune) bool {
	_, exists := cs.CharToPos[char]

	return exists
}

// GetPosition returns the position of a character in the set.
func (cs *CharacterSet) GetPosition(char rune) (int, bool) {
	pos, exists := cs.CharToPos[char]

	return pos, exists
}

// GetCharacter returns the character at a given position.
func (cs *CharacterSet) GetCharacter(pos int) (rune, bool) {
	char, exists := cs.PosToChar[pos]

	return char, exists
}

// IsValid checks if all positions are filled with valid characters.
func (cs *CharacterSet) IsValid(layout []rune) bool {
	if len(layout) != cs.Size {
		return false
	}

	seen := make(map[rune]bool)

	for _, char := range layout {
		if char == 0 {
			return false // Null character
		}

		if !cs.Contains(char) {
			return false // Character not in set
		}

		if seen[char] {
			return false // Duplicate
		}

		seen[char] = true
	}

	return len(seen) == cs.Size
}

// Predefined character sets

// AlphabetOnly returns a character set with only lowercase letters a-z.
func AlphabetOnly() *CharacterSet {
	chars := []rune("abcdefghijklmnopqrstuvwxyz")

	return NewCharacterSet("alphabet", chars)
}

// AlphabetWithNumbers returns letters and numbers.
func AlphabetWithNumbers() *CharacterSet {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	return NewCharacterSet("alphanumeric", chars)
}

// ProgrammingCharset returns letters, numbers, and common programming symbols.
func ProgrammingCharset() *CharacterSet {
	// Include common programming symbols
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	numbers := "0123456789"
	symbols := "!@#$%^&*()_+-=[]{}|;':\",./<>?`~"

	chars := []rune(alphabet + numbers + symbols)

	return NewCharacterSet("programming", chars)
}

// FullKeyboardCharset returns the complete keyboard character set.
func FullKeyboardCharset() *CharacterSet {
	// Complete US QWERTY keyboard layout
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	numbers := "0123456789"
	symbols := "!@#$%^&*()_+-=[]{}\\|;':\",./<>?`~ "

	chars := []rune(alphabet + numbers + symbols)

	return NewCharacterSet("full_keyboard", chars)
}

// CustomCharset allows creating a custom character set from a string.
func CustomCharset(name, charString string) *CharacterSet {
	// Remove duplicates while preserving order
	seen := make(map[rune]bool)

	var uniqueChars []rune

	for _, char := range charString {
		if !seen[char] && char != 0 { // Skip null characters
			seen[char] = true
			uniqueChars = append(uniqueChars, char)
		}
	}

	return NewCharacterSet(name, uniqueChars)
}

// GetCharsetByName returns a predefined character set by name.
func GetCharsetByName(name string) *CharacterSet {
	switch name {
	case "alphabet":
		return AlphabetOnly()
	case "alphanumeric":
		return AlphabetWithNumbers()
	case "programming":
		return ProgrammingCharset()
	case "full", "full_keyboard":
		return FullKeyboardCharset()
	default:
		return AlphabetOnly() // Default fallback
	}
}

// Common character groups for reference.
var (
	Letters      = "abcdefghijklmnopqrstuvwxyz"
	Numbers      = "0123456789"
	BasicSymbols = "!@#$%^&*()"
	Punctuation  = ".,;:'\""
	Brackets     = "()[]{}"
	Operators    = "+-=*/<>?|&^%"
	Special      = "`~_\\#"
	Space        = " "
)
