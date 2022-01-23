package utils

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

func isAlpha(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '-' || ch == '!'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlnum(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

func matchPart(b byte) func(byte) bool {
	return func(c byte) bool {
		return c != b && c != '/'
	}
}

func match(s string, f func(byte) bool, i int) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) {
		j++
	}
	if j < len(s) {
		next = s[j]
	}
	return s[i:j], next, j
}

// ParamsMatch match string by param
func ParamsMatch(source string, pth string) (url.Values, string, bool) {
	var (
		i, j int
		p    = make(url.Values)
		ext  = path.Ext(pth)
	)

	pth = strings.TrimSuffix(pth, ext)

	if ext != "" {
		p.Add(":format", strings.TrimPrefix(ext, "."))
	}

	for i < len(pth) {
		switch {
		case j >= len(source):

			if source != "/" && len(source) > 0 && source[len(source)-1] == '/' {
				return p, pth[:i], true
			}

			if source == "" && pth == "/" {
				return p, pth, true
			}
			return p, pth[:i], false
		case source[j] == ':':
			var name, val string
			var nextc byte

			name, nextc, j = match(source, isAlnum, j+1)
			val, _, i = match(pth, matchPart(nextc), i)

			if (j < len(source)) && source[j] == '[' {
				var index int
				if idx := strings.Index(source[j:], "]/"); idx > 0 {
					index = idx
				} else if source[len(source)-1] == ']' {
					index = len(source) - j - 1
				}

				if index > 0 {
					match := strings.TrimSuffix(strings.TrimPrefix(source[j:j+index+1], "["), "]")
					if reg, err := regexp.Compile("^" + match + "$"); err == nil && reg.MatchString(val) {
						j = j + index + 1
					} else {
						return nil, "", false
					}
				}
			}

			p.Add(":"+name, val)
		case pth[i] == source[j]:
			i++
			j++
		default:
			return nil, "", false
		}
	}

	if j != len(source) {
		if (len(source) == j+1) && source[j] == '/' {
			return p, pth, true
		}

		return nil, "", false
	}
	return p, pth, true
}
