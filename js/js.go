// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js // import "github.com/tdewolff/minify/js"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse/js"
)

var (
	spaceBytes     = []byte(" ")
	newlineBytes   = []byte("\n")
	semicolonBytes = []byte(";")
)

////////////////////////////////////////////////////////////////

// Minifier is a JS minifier.
type Minifier struct{}

// Minify minifies JS data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return (&Minifier{}).Minify(m, w, r, params)
}

// Minify minifies JS data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	prev := js.LineTerminatorToken
	prevLast := byte(' ')
	lineTerminatorQueued := false
	whitespaceQueued := false
	semicolonQueued := false

	l := js.NewLexer(r)
	for {
		tt, data := l.Next()
		if tt == js.ErrorToken {
			if l.Err() != io.EOF {
				return l.Err()
			}
			return nil
		} else if tt == js.LineTerminatorToken {
			lineTerminatorQueued = true
		} else if tt == js.WhitespaceToken {
			whitespaceQueued = true
		} else if tt != js.CommentToken {
			first := data[0]
			if tt == js.PunctuatorToken && first == ';' {
				if semicolonQueued {
					// when two semicolons follow, print them (such as in `for (var i = 0;;)`)
					if _, err := w.Write(data); err != nil {
						return err
					}
				}
				semicolonQueued = true
			} else {
				if semicolonQueued && (tt != js.PunctuatorToken || first != '}') {
					if _, err := w.Write(semicolonBytes); err != nil {
						return err
					}
				}
				if (prev == js.IdentifierToken || prev == js.NumericToken || prev == js.PunctuatorToken || prev == js.StringToken || prev == js.RegexpToken) && (tt == js.IdentifierToken || tt == js.NumericToken || tt == js.PunctuatorToken || tt == js.RegexpToken) {
					if lineTerminatorQueued && (prev != js.PunctuatorToken || prevLast == '}' || prevLast == ']' || prevLast == ')' || prevLast == '+' || prevLast == '-' || prevLast == '"' || prevLast == '\'') && (tt != js.PunctuatorToken || first == '{' || first == '[' || first == '(' || first == '+' || first == '-') {
						if _, err := w.Write(newlineBytes); err != nil {
							return err
						}
					} else if whitespaceQueued && (prev != js.StringToken && prev != js.PunctuatorToken && tt != js.PunctuatorToken || (prevLast == '+' || prevLast == '-') && first == prevLast) {
						if _, err := w.Write(spaceBytes); err != nil {
							return err
						}
					}
				}
				if _, err := w.Write(data); err != nil {
					return err
				}
				lineTerminatorQueued = false
				whitespaceQueued = false
				semicolonQueued = false
			}
			prev = tt
			prevLast = data[len(data)-1]
		}
		l.Free(len(data))
	}
	return nil
}
