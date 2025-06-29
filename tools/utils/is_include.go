package utils

func IsInclude[T comparable](data []T, target T) bool {
	for _, datum := range data {
		if datum == target {
			return true
		}
	}

	return false
}
