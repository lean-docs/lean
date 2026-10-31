package ir

import "strconv"

func BuiltinStyle(id string) (Style, bool) {
	style := Style{ID: id, Name: id}
	switch id {
	case "Normal":
		return style, true
	case "Title":
		style.RunAttrs = RunAttrs{Bold: true, FontSize: 26}
		return style, true
	case "Subtitle":
		style.RunAttrs = RunAttrs{FontSize: 15}
		return style, true
	}
	if len(id) == 8 && id[:7] == "Heading" {
		level, err := strconv.Atoi(id[7:])
		if err == nil && level >= 1 && level <= 6 {
			style.Name = "Heading " + strconv.Itoa(level)
			style.RunAttrs = RunAttrs{Bold: true, FontSize: []float64{20, 18, 16, 14, 12, 11}[level-1]}
			style.ParaAttrs.SetOutlineLevel(level - 1)
			style.ParaAttrs.SetKeepWithNext(true)
			return style, true
		}
	}
	return Style{}, false
}
