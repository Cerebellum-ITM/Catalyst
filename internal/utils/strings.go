package utils

func InsertString(original string, insertion string, index int) string {
	if index < 0 || index > len(original) {
		return original
	}
	return original[:index] + insertion + original[index:]
}
