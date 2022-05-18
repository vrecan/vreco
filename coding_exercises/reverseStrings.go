package codingexercises

func reverseString(s string) string {
	runes := []rune(s)
	size := len(runes)
	newRunes := make([]rune, size)
	for i, r := range runes {
		newRunes[(size-i)-1] = r
	}
	return string(newRunes)
}

func inplaceReverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
