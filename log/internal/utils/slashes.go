package utils

func Addslashes(str string) string {
	runes := make([]rune, 0, len(str))
	for _, ch := range str {
		switch ch {
		case '\'':
			runes = append(runes, []rune(`\'`)...)
		case '"':
			runes = append(runes, []rune(`\"`)...)
		case '\\':
			runes = append(runes, []rune(`\\`)...)
		case '\n':
			runes = append(runes, []rune(`\n`)...)
		case '\t':
			runes = append(runes, []rune(`\t`)...)
		case '\r':
			runes = append(runes, []rune(`\r`)...)
		default:
			runes = append(runes, ch)
		}
	}

	return string(runes)
}
