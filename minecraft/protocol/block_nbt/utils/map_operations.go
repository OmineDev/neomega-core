package utils

import (
	"maps"
)

// 将一个或多个 map[string]any 合并到一起
func MergeMaps(mapping ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, value := range mapping {
		maps.Copy(result, value)
	}
	return result
}
