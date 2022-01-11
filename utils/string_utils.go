package utils

import (
	"strings"
)

//CreateTickerKeywords creates a list of keys to search on.
func CreateTickerKeywords(names []string) []string {

	var keys []string
	for _, name := range names {
		nkeys := createKeyWords(name)
		for _, nkey := range nkeys {
			keys = append(keys, nkey)
		}
		words := strings.Split(name, " ")
		for _, word := range words {
			wkeys := createKeyWords(word)
			for _, wkey := range wkeys {
				keys = append(keys, wkey)
			}
		}
	}
	return keys
}

func createKeyWords(word string) []string {

	var keys []string
	if len(word) == 0 {
		return keys
	}
	var key string
	var i = 0
	key = string(word[i])
	for j := i + 1; j < len(word); j++ {
		key = key + string(word[j])
		if len(key) >= 3 {
			// log.Println(key)
			keys = append(keys, key)
		}
	}
	// }
	return keys
}
