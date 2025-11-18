package testdata

// computeTotal calculates the sum of all numbers in the slice
func computeTotal(numbers []int) int {
	total := 0
	for _, num := range numbers {
		total = total + num
	}
	return total
}

// findMaximum returns the largest value in the slice
func findMaximum(values []int) int {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for i := 1; i < len(values); i++ {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}
