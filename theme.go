package main

import (
	//"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	B16_DIR      = os.ExpandEnv("$HOME/.themes/shell/scripts")
	RE_b16name   = regexp.MustCompile(`base16-(.+)\.sh`)
	RE_shellAssn = regexp.MustCompile(`(?m)^color_?\w+=\S+`)
	RE_colorkey  = regexp.MustCompile(`\bcolor_?(\d+|foreground|background)\b`)
)

type ThemeMap map[string]Theme

func allThemes() ThemeMap {
	ents, err := os.ReadDir(B16_DIR)
	if err != nil {
		log.Fatalln(err)
	}

	tm := make(ThemeMap)
	for _, ent := range ents {
		if match := RE_b16name.FindStringSubmatch(ent.Name()); len(match) > 1 {
			theme := match[1]
			tm[theme] = parseTheme(theme)
		}
	}
	return tm
}

func shellKeyIndex(k string) (int, bool) {
	if match := RE_colorkey.FindStringSubmatch(k); len(match) > 1 {
		if index, err := strconv.ParseInt(match[1], 10, 32); err == nil {
			return int(index), true
		}
	}
	return 0, false
}

func parseTheme(theme string) Theme {
	file, err := os.ReadFile(B16_DIR + "/base16-" + theme + ".sh")
	if err != nil {
		log.Fatalln("couldn't read theme file", theme)
	}

	var t []string
	var fg, bg string

	for _, assn := range RE_shellAssn.FindAll(file, -1) {
		key, val, _ := strings.Cut(string(assn), "=")
		if index, found := shellKeyIndex(val); found {
			val = t[index]
		}
		if index, found := shellKeyIndex(key); found {
			if index != len(t) {
				panic("theme keys out of order")
			}
			t = append(t, val)
		} else if key == "color_foreground" {
			fg = val
		} else if key == "color_background" {
			bg = val
		} else {
			log.Fatalln("couldn't parse key", key)
		}
	}

	return make(Theme).From(t, fg, bg)
}

func (tm ThemeMap) Apply(theme string) Theme {
	t, found := tm[theme]
	if !found {
		log.Fatalln("nonexistant theme", theme)
	}
	return t.Apply()
}
