package ui

import "strings"

func fixedWidth(s string, width int) string {
	s = strings.ReplaceAll(s, "\t", " ")
	if width <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) > width {
		r = r[:width]
	}
	out := string(r)
	if len([]rune(out)) < width {
		out = padRight(out, width)
	}
	return out
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(r))
}

func trimTrailingEmpty(lines []string) []string {
	i := len(lines) - 1
	for i >= 0 && strings.TrimSpace(lines[i]) == "" {
		i--
	}
	return lines[:i+1]
}

func maxStringLen(ss []string) int {
	m := 0
	for _, s := range ss {
		if len([]rune(s)) > m {
			m = len([]rune(s))
		}
	}
	return m
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
