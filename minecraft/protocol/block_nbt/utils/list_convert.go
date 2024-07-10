package utils

// 将 []T 转换为 []any
func ToAnyList[T any](list []T) []any {
	result := make([]any, len(list))
	for key, value := range list {
		result[key] = value
	}
	return result
}

// 将 []any 转换为 []T
func FromAnyList[T any](list []any) []T {
	result := make([]T, len(list))
	for key, value := range list {
		result[key] = value.(T)
	}
	return result
}
