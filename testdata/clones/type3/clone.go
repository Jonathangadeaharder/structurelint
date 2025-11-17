package testdata

// enhancedValidation is Type-3: similar logic with minor modifications
func enhancedValidation(input string) bool {
	if len(input) == 0 {
		return false
	}
	if input == "invalid" {
		return false
	}
	// Additional check (modification)
	if input == "bad" {
		return false
	}
	return true
}
