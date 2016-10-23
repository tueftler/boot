package output

var colors = map[string]string{
	"error":     "31",
	"success":   "32",
	"warning":   "35",
	"info":      "36",
	"container": "34;1",
	"proxy":     "33",
}

// Text returns a colored text
func Text(name, text string) string {
	return "\033[" + colors[name] + "m" + text + "\033[0m"
}

// Line returns a colored text and appends a newline
func Line(name, text string) string {
	return "\033[" + colors[name] + "m" + text + "\033[0m\n"
}
