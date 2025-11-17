package testdata

// calculateSum is a Type-2 clone: same structure, different variable names
func calculateSum(nums []int) int {
	sum := 0
	for _, n := range nums {
		sum = sum + n
	}
	return sum
}

// getBiggest is a Type-2 clone: same logic, renamed variables
func getBiggest(data []int) int {
	if len(data) == 0 {
		return 0
	}
	biggest := data[0]
	for idx := 1; idx < len(data); idx++ {
		if data[idx] > biggest {
			biggest = data[idx]
		}
	}
	return biggest
}
