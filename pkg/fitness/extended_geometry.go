package fitness

import "github.com/tommoulard/keyboardgen/pkg/genetic"

// ExtendedGeometry defines keyboard geometry supporting special characters.
func ExtendedGeometry() KeyboardGeometry {
	positions := make(map[int][2]float64)
	fingerMap := make(map[int]int)

	// Number row (0-9 and symbols)
	numberRowChars := []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	numberRowSymbols := []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')'}

	for i, char := range numberRowChars {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i), -1} // Row above QWERTY
			fingerMap[pos] = getFingerForColumn(i)
		}
	}

	for i, char := range numberRowSymbols {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i), -1} // Same position as numbers
			fingerMap[pos] = getFingerForColumn(i)
		}
	}

	// Top row (QWERTY)
	topRowChars := []rune{'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p'}
	topRowSymbols := []rune{'[', ']', '\\'}

	for i, char := range topRowChars {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i), 0}
			fingerMap[pos] = getFingerForColumn(i)
		}
	}
	// Brackets and backslash
	for i, char := range topRowSymbols {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i + 10), 0}
			fingerMap[pos] = 7 // Right pinky
		}
	}

	// Middle row (ASDF)
	middleRowChars := []rune{'a', 's', 'd', 'f', 'g', 'h', 'j', 'k', 'l'}
	middleRowSymbols := []rune{';', '\''}

	for i, char := range middleRowChars {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i) + 0.5, 1}
			fingerMap[pos] = getFingerForColumn(i)
		}
	}

	for i, char := range middleRowSymbols {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i+9) + 0.5, 1}
			fingerMap[pos] = 7 // Right pinky
		}
	}

	// Bottom row (ZXCV)
	bottomRowChars := []rune{'z', 'x', 'c', 'v', 'b', 'n', 'm'}
	bottomRowSymbols := []rune{',', '.', '/'}

	for i, char := range bottomRowChars {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i) + 1, 2}
			fingerMap[pos] = getFingerForColumn(i)
		}
	}

	for i, char := range bottomRowSymbols {
		pos := getCharPosition(char)
		if pos >= 0 {
			positions[pos] = [2]float64{float64(i+7) + 1, 2}
			fingerMap[pos] = getFingerForColumn(i + 7)
		}
	}

	// Special characters
	specialChars := []struct {
		char   rune
		x, y   float64
		finger int
	}{
		{'`', -1, -1, 0}, // Backtick (left of 1)
		{'~', -1, -1, 0}, // Tilde (shift + backtick)
		{'-', 10, -1, 7}, // Minus (right of 0)
		{'_', 10, -1, 7}, // Underscore (shift + minus)
		{'=', 11, -1, 7}, // Equals (right of minus)
		{'+', 11, -1, 7}, // Plus (shift + equals)
		{'{', 10, 0, 7},  // Left brace (shift + [)
		{'}', 11, 0, 7},  // Right brace (shift + ])
		{'|', 12, 0, 7},  // Pipe (shift + \)
		{':', 10, 1, 7},  // Colon (shift + ;)
		{'"', 11, 1, 7},  // Quote (shift + ')
		{'<', 7, 2, 5},   // Less than (shift + ,)
		{'>', 8, 2, 6},   // Greater than (shift + .)
		{'?', 9, 2, 7},   // Question mark (shift + /)
		{' ', 5, 3, 4},   // Space bar (thumbs)
	}

	for _, sc := range specialChars {
		pos := getCharPosition(sc.char)
		if pos >= 0 {
			positions[pos] = [2]float64{sc.x, sc.y}
			fingerMap[pos] = sc.finger
		}
	}

	return KeyboardGeometry{
		KeyPositions: positions,
		FingerMap:    fingerMap,
	}
}

// Helper function to get character position in full character set.
func getCharPosition(char rune) int {
	// This assumes we're using the full keyboard character set
	fullCharset := genetic.FullKeyboardCharset()

	pos, exists := fullCharset.GetPosition(char)
	if !exists {
		return -1
	}

	return pos
}

// Helper function to map column to finger.
func getFingerForColumn(col int) int {
	// Finger mapping: 0=left pinky, 1=left ring, 2=left middle, 3=left index
	//                 4=right index, 5=right middle, 6=right ring, 7=right pinky
	fingerMap := []int{0, 1, 2, 3, 3, 4, 4, 5, 6, 7}
	if col < 0 || col >= len(fingerMap) {
		return 7 // Default to right pinky for out of bounds
	}

	return fingerMap[col]
}

// ProgrammingGeometry returns a geometry optimized for programming with symbols.
func ProgrammingGeometry() KeyboardGeometry {
	// Use extended geometry but adjust some symbol positions for better programming ergonomics
	geom := ExtendedGeometry()

	// Move frequently used programming symbols to more accessible positions
	// This is a starting point - could be further optimized based on usage patterns

	return geom
}

// GetGeometryForCharset returns appropriate geometry for a character set.
func GetGeometryForCharset(charset *genetic.CharacterSet) KeyboardGeometry {
	if charset == nil {
		return StandardGeometry()
	}

	switch charset.Size {
	case 26: // Alphabet only
		return StandardGeometry()
	case 36: // Alphanumeric
		return ExtendedGeometry()
	default: // Programming or full keyboard
		if charset.Name == "programming" {
			return ProgrammingGeometry()
		}

		return ExtendedGeometry()
	}
}
