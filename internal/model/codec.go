package model

import (
	"encoding/json"
	"sort"
)

func Encode(value any) ([]byte, error) {
	return json.Marshal(value)
}

func Decode(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

func NormalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}
