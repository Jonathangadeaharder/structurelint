package testdata

// simpleValidation performs basic input validation
func simpleValidation(input string) bool {
	if len(input) == 0 {
		return false
	}
	if input == "invalid" {
		return false
	}
	return true
}
