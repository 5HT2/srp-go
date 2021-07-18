package main

// AppendIfMissing will append a string to slice if the slice does not already have that string
func AppendIfMissing(slice []string, str string) []string {
	for _, ele := range slice {
		if ele == str {
			return slice
		}
	}
	return append(slice, str)
}

// ToInt will return an int from a float rounded to the nearest whole
func ToInt(f float64) int {
	return int(f + 0.5)
}
