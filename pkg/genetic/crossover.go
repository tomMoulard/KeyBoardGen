package genetic

import (
	"math/rand/v2"
)

// CrossoverMethod defines different crossover strategies.
type CrossoverMethod int

const (
	OrderCrossover CrossoverMethod = iota
	PartiallyMatchedCrossover
	CycleCrossover
	UniformCrossover
)

// Crossover handles genetic crossover operations.
type Crossover struct {
	method CrossoverMethod
}

// NewCrossover creates a new crossover operator.
func NewCrossover(method CrossoverMethod) *Crossover {
	return &Crossover{method: method}
}

// Apply performs crossover between two parents.
func (c *Crossover) Apply(parent1, parent2 Individual) Individual {
	switch c.method {
	case OrderCrossover:
		return c.orderCrossover(parent1, parent2)
	case PartiallyMatchedCrossover:
		return c.partiallyMatchedCrossover(parent1, parent2)
	case CycleCrossover:
		return c.cycleCrossover(parent1, parent2)
	case UniformCrossover:
		return c.uniformCrossover(parent1, parent2)
	default:
		return c.orderCrossover(parent1, parent2)
	}
}

// orderCrossover (OX) - preserves order of elements.
func (c *Crossover) orderCrossover(parent1, parent2 Individual) Individual {
	length := len(parent1.Layout)

	// Choose two random crossover points
	point1 := rand.IntN(length)

	point2 := rand.IntN(length)
	if point1 > point2 {
		point1, point2 = point2, point1
	}

	// Initialize child with same size layout as parents
	child := Individual{
		Layout:  make([]rune, length),
		Charset: parent1.Charset,
		Age:     0,
	}
	used := make(map[rune]bool)

	// Copy segment from parent1
	for i := point1; i <= point2; i++ {
		child.Layout[i] = parent1.Layout[i]
		used[parent1.Layout[i]] = true
	}

	// Fill remaining positions from parent2 in order
	childIdx := 0
	for parentIdx := range length {
		// Skip positions already filled
		if childIdx >= point1 && childIdx <= point2 {
			childIdx = point2 + 1
		}

		if childIdx >= length {
			break
		}

		char := parent2.Layout[parentIdx]
		if !used[char] {
			child.Layout[childIdx] = char
			used[char] = true
			childIdx++
		}
	}

	// Handle wrap-around
	if childIdx < point1 {
		for parentIdx := 0; parentIdx < length && childIdx < point1; parentIdx++ {
			char := parent2.Layout[parentIdx]
			if !used[char] {
				child.Layout[childIdx] = char
				used[char] = true
				childIdx++
			}
		}
	}

	return child
}

// partiallyMatchedCrossover (PMX) - creates mapping between parents.
func (c *Crossover) partiallyMatchedCrossover(parent1, parent2 Individual) Individual {
	length := len(parent1.Layout)

	// Choose two random crossover points
	point1 := rand.IntN(length)

	point2 := rand.IntN(length)
	if point1 > point2 {
		point1, point2 = point2, point1
	}

	// Initialize child as copy of parent1
	child := Individual{
		Layout:  make([]rune, length),
		Charset: parent1.Charset,
		Age:     0,
	}
	copy(child.Layout, parent1.Layout)

	// Create mapping between parents in the crossover segment
	mapping := make(map[rune]rune)

	for i := point1; i <= point2; i++ {
		char1 := parent1.Layout[i]
		char2 := parent2.Layout[i]

		if char1 != char2 {
			mapping[char1] = char2
			mapping[char2] = char1
		}
	}

	// Copy segment from parent2
	for i := point1; i <= point2; i++ {
		child.Layout[i] = parent2.Layout[i]
	}

	// Resolve conflicts outside the crossover segment
	for i := range length {
		if i >= point1 && i <= point2 {
			continue // Skip crossover segment
		}

		char := child.Layout[i]
		for {
			if newChar, exists := mapping[char]; exists {
				char = newChar
			} else {
				break
			}
		}

		child.Layout[i] = char
	}

	return child
}

// cycleCrossover (CX) - maintains position relationships.
func (c *Crossover) cycleCrossover(parent1, parent2 Individual) Individual {
	length := len(parent1.Layout)
	child := Individual{
		Layout:  make([]rune, length),
		Charset: parent1.Charset,
		Age:     0,
	}
	used := make([]bool, length)

	// Find cycles
	for start := range length {
		if used[start] {
			continue
		}

		// Determine which parent to use for this cycle
		useParent1 := (start%2 == 0)

		// Trace the cycle
		current := start
		for !used[current] {
			used[current] = true

			if useParent1 {
				child.Layout[current] = parent1.Layout[current]
			} else {
				child.Layout[current] = parent2.Layout[current]
			}

			// Find next position in cycle
			target := parent1.Layout[current]
			if !useParent1 {
				target = parent2.Layout[current]
			}

			// Find where this target appears in the other parent
			next := -1

			otherParent := parent2.Layout
			if !useParent1 {
				otherParent = parent1.Layout
			}

			for i, char := range otherParent {
				if char == target {
					next = i

					break
				}
			}

			if next == -1 || next == current {
				break
			}

			current = next
		}
	}

	return child
}

// uniformCrossover - randomly chooses each gene from either parent.
func (c *Crossover) uniformCrossover(parent1, parent2 Individual) Individual {
	length := len(parent1.Layout)
	child := Individual{
		Layout:  make([]rune, length),
		Charset: parent1.Charset,
		Age:     0,
	}
	used := make(map[rune]bool)
	remaining1 := make([]rune, 0, length)
	remaining2 := make([]rune, 0, length)

	// First pass: randomly select genes that don't create conflicts
	for i := range length {
		char1 := parent1.Layout[i]
		char2 := parent2.Layout[i]

		if rand.Float64() < 0.5 {
			// Try to use parent1
			if !used[char1] {
				child.Layout[i] = char1
				used[char1] = true
			} else {
				remaining2 = append(remaining2, char2)
			}
		} else {
			// Try to use parent2
			if !used[char2] {
				child.Layout[i] = char2
				used[char2] = true
			} else {
				remaining1 = append(remaining1, char1)
			}
		}
	}

	// Second pass: fill empty positions
	remaining := append(remaining1, remaining2...)
	remainingIdx := 0

	for i := range length {
		if child.Layout[i] == 0 { // Empty position
			// Find next unused character
			for remainingIdx < len(remaining) {
				char := remaining[remainingIdx]
				remainingIdx++

				if !used[char] {
					child.Layout[i] = char
					used[char] = true

					break
				}
			}
		}
	}

	// Final pass: ensure all positions are filled
	for i := range length {
		if child.Layout[i] == 0 {
			// Find any unused character
			for char := 'a'; char <= 'z'; char++ {
				if !used[char] {
					child.Layout[i] = char
					used[char] = true

					break
				}
			}
		}
	}

	return child
}

// TwoPointCrossover performs traditional two-point crossover (for comparison).
func (c *Crossover) TwoPointCrossover(parent1, parent2 Individual) Individual {
	return c.orderCrossover(parent1, parent2)
}

// ValidateChild ensures the child has a valid keyboard layout.
func ValidateChild(child Individual) Individual {
	if child.Charset == nil {
		child.Charset = AlphabetOnly() // Default to alphabet
	}

	// If already valid, return as-is
	if child.Charset.IsValid(child.Layout) {
		return child
	}

	// Fix the layout using the charset
	validChars := child.Charset.Characters
	seen := make(map[rune]bool)
	missing := make([]rune, 0, len(validChars))
	invalidPositions := make([]int, 0)

	// Initialize missing with all valid characters
	missing = append(missing, validChars...)

	// Find invalid positions (duplicates, nulls, invalid chars) and track seen characters
	for i, char := range child.Layout {
		if char == 0 || !child.Charset.Contains(char) || seen[char] {
			// Invalid position: null character, not in charset, or duplicate
			invalidPositions = append(invalidPositions, i)
		} else {
			seen[char] = true
			// Remove from missing list
			for j, missingChar := range missing {
				if missingChar == char {
					missing = append(missing[:j], missing[j+1:]...)

					break
				}
			}
		}
	}

	// Fix invalid positions by replacing with missing characters
	missingIdx := 0
	for _, pos := range invalidPositions {
		if missingIdx < len(missing) {
			child.Layout[pos] = missing[missingIdx]
			missingIdx++
		}
	}

	return child
}
