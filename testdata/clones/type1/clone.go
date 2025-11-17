package testdata

// calculateSum adds two numbers and returns the result
func calculateSum(a int, b int) int {
	result := a + b
	return result
}

// This is an EXACT clone of processData - Type-1 clone (identical)
func validateItems(items []string) []string {
	var filtered []string
	for _, item := range items {
		if len(item) > 0 {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
