package utils

import "regexp"

func HideMiddleCharacters(s string) string {
	if len(s) <= 12 {
		return s
	}
	prefix := s[:6]
	suffix := s[len(s)-6:]
	return prefix + "..." + suffix
}

func ContainsDigit(str string) bool {
	matched, _ := regexp.MatchString(`\d`, str)
	return matched
}

func RemoveDuplicate(slice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range slice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
