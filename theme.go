package main

import (
	"fmt"
	"image/color"
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
	RE_hexdigit  = regexp.MustCompile(`[[:xdigit:]]`)
)

type Theme struct {
	palette    []color.RGBA
	foreground color.RGBA
	background color.RGBA
}

type ThemeMap map[string]Theme

func allThemes() ThemeMap {
	ents, err := os.ReadDir(B16_DIR)
	if err != nil {
		log.Fatalln(err)
	}

	themes := ThemeMap{}
	for _, ent := range ents {
		if match := RE_b16name.FindStringSubmatch(ent.Name()); len(match) > 1 {
			theme := match[1]
			themes[theme] = parseTheme(theme)
		}
	}
	return themes
}

func shellKeyIndex(k string) (int, bool) {
	if match := RE_colorkey.FindStringSubmatch(k); len(match) > 1 {
		if index, err := strconv.ParseInt(match[1], 10, 32); err == nil {
			return int(index), true
		}
	}
	return 0, false
}

func parseColor(s string) color.RGBA {
	if len(s) != 6 {
		s = strings.Join(RE_hexdigit.FindAllString(s, -1), "")
	}
	if len(s) != 6 {
		return color.RGBA{}
	}
	if val, err := strconv.ParseUint(s, 16, 32); err == nil {
		return color.RGBA{uint8(val >> 16), uint8(val >> 8), uint8(val), 0xff}
	}
	return color.RGBA{}
}

func parseTheme(theme string) Theme {
	file, err := os.ReadFile(B16_DIR + "/base16-" + theme + ".sh")
	if err != nil {
		log.Fatalln("couldn't read theme file", theme)
	}

	var t Theme
	for _, assn := range RE_shellAssn.FindAll(file, -1) {
		var cVal color.RGBA
		key, val, _ := strings.Cut(string(assn), "=")
		if index, found := shellKeyIndex(val); found {
			cVal = t.palette[index]
		} else {
			cVal = parseColor(val)
			if cVal.A != 0xff {
				log.Fatalln("bad color value parse", cVal)
			}
		}

		if index, found := shellKeyIndex(key); found {
			if len(t.palette) != index {
				log.Fatalln("colors out of order")
			}
			t.palette = append(t.palette, cVal)
		} else if key == "color_foreground" {
			t.foreground = cVal
		} else if key == "color_background" {
			t.background = cVal
		} else {
			log.Fatalln("couldn't parse key", key)
		}
	}
	return t
}

func initc(code string, color color.RGBA) {
	os.Stderr.WriteString(
		fmt.Sprintf(
			"\x1b]%s;rgb:%02x/%02x/%02x\x1b\\",
			code, color.R, color.G, color.B,
		),
	)
}

func (t Theme) validate() bool {
	for _, c := range t.palette {
		if c.A != 0xff {
			return false
		}
	}
	if 0xff != t.foreground.A || 0xff != t.background.A {
		return false
	}
	return true
}

func (t Theme) SetTheme() {
	if !t.validate() {
		log.Fatalln("invalid theme", t)
	}

	for index, color := range t.palette {
		initc(fmt.Sprintf("4;%d", index), color)
	}

	initc(fmt.Sprintf("%d", 10), t.foreground)
	initc(fmt.Sprintf("%d", 11), t.background)
}

func (tm ThemeMap) SetTheme(theme string) {
	t, found := tm[theme]
	if !found {
		log.Fatalln("nonexistant theme", theme)
	}
	t.SetTheme()
}
