package helpers

import "strings"

var bigDigits = map[rune][5]string{
	'0': {
		" _--_ ",
		"/    \\",
		"|    |",
		"\\    /",
		" -__- ",
	},
	'1': {
		"   __ ",
		"  / | ",
		" /  | ",
		"    | ",
		"    | ",
	},
	'2': {
		" ___  ",
		"/   \\ ",
		"   _/ ",
		" _/   ",
		"/____ ",
	},
	'3': {
		" ___  ",
		"/   \\ ",
		" ___/ ",
		"    \\ ",
		"\\___/ ",
	},
	'4': {
		"  ___ ",
		" /  | ",
		"/___| ",
		"    | ",
		"    | ",
	},
	'5': {
		" ____ ",
		"|     ",
		"|___  ",
		"    \\ ",
		"\\___/ ",
	},
	'6': {
		"  __  ",
		" /  \\ ",
		"/___  ",
		"|   \\ ",
		"\\___/ ",
	},
	'7': {
		"_____ ",
		"    / ",
		"   /  ",
		"  /   ",
		" /    ",
	},
	'8': {
		" ___  ",
		"/   \\ ",
		"\\___/ ",
		"/   \\ ",
		"\\___/ ",
	},
	'9': {
		" ___  ",
		"/   \\ ",
		"\\___| ",
		"    | ",
		"\\__/  ",
	},
	':': {
		" ",
		"o",
		" ",
		"o",
		" ",
	},
	' ': {
		" ",
		" ",
		" ",
		" ",
		" ",
	},
}

// RenderBigTime convierte una cadena como "15:04:05" en texto ASCII art de 5 líneas
func RenderBigTime(s string) string {
	var lines [5]string

	for _, ch := range s {
		glyph, ok := bigDigits[ch]
		if !ok {
			glyph = bigDigits[' ']
		}
		for row := 0; row < 5; row++ {
			if len(lines[row]) > 0 {
				lines[row] += "  "
			}
			lines[row] += glyph[row]
		}
	}

	return strings.Join(lines[:], "\n")
}
