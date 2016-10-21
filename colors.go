package main

var colors = map[string]string{
	"error":   "31",
	"success": "32",
	"info":    "36",

	"docker": "34;1",
	"proxy":  "33",
}

func text(name, text string) string {
	return "\033[" + colors[name] + "m" + text + "\033[0m"
}

func line(name, text string) string {
	return "\033[" + colors[name] + "m" + text + "\033[0m\n"
}
