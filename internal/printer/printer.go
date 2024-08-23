/*
Copyright 2020 Cornelius Weig

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package printer

import (
	"fmt"
	"io"
	"sync"

	"github.com/corneliusweig/tabwriter"
)

type color int

const (
	red    = color(31)
	green  = color(32)
	purple = color(35)
	none   = color(0)
)

var (
	isTerminal = isTerminalImpl
	once       sync.Once
)

type Outcome uint8

const (
	None Outcome = iota
	Up
	Down
	Err
)

func (o Outcome) Render(e Env) string {
	conv := humanreadableAccessCode
	if e.IsTerminal {
		conv = colored(conv)
	}
	if e.OutputFormat == "ascii-table" {
		conv = asciiAccessCode
	}
	return conv(o)
}

type Env struct {
	IsTerminal   bool
	OutputFormat string
}

type Renderable interface {
	Render(Env) string
}

type Text string

func (t Text) Render(_ Env) string {
	return string(t)
}

type Row struct {
	Intro   []Renderable
	Entries []Outcome
}
type Table struct {
	Headers []Renderable
	Rows    []Row
}

func TableWithHeaders(headers []Renderable) *Table {
	return &Table{
		Headers: headers,
	}
}

func (p *Table) AddRow(intro []Renderable, outcomes ...Outcome) {
	row := Row{
		Intro:   intro,
		Entries: outcomes,
	}
	p.Rows = append(p.Rows, row)
}

func (p *Table) Render(out io.Writer, outputFormat string) {
	once.Do(func() { initTerminal(out) })

	w := tabwriter.NewWriter(out, 4, 8, 2, ' ', tabwriter.SmashEscape|tabwriter.StripEscape)
	defer w.Flush()

	env := Env{
		IsTerminal:   isTerminal(out),
		OutputFormat: outputFormat,
	}

	// table header
	for i, h := range p.Headers {
		if i == 0 {
			fmt.Fprint(w, h.Render(env))
		} else {
			fmt.Fprintf(w, "\t%s", h.Render(env))
		}
	}
	fmt.Fprint(w, "\n")

	// table body
	for _, row := range p.Rows {
		for i, e := range row.Intro {
			if i != 0 {
				fmt.Fprintf(w, "\t")
			}
			fmt.Fprintf(w, "%s", e.Render(env)) // FIXME
		}
		//fmt.Fprintf(w, "%s", strings.Join(row.Intro, "\t"))
		for _, e := range row.Entries {
			fmt.Fprintf(w, "\t%s", e.Render(env)) // FIXME
		}
		fmt.Fprint(w, "\n")
	}
}

func humanreadableAccessCode(o Outcome) string {
	switch o {
	case None:
		return ""
	case Up:
		return "✔" // ✓
	case Down:
		return "✖" // ✕
	case Err:
		return "ERR"
	default:
		panic("unknown access code")
	}
}

func colored(wrap func(Outcome) string) func(Outcome) string {
	return func(o Outcome) string {
		c := none
		switch o {
		case Up:
			c = green
		case Down:
			c = red
		case Err:
			c = purple
		}
		return fmt.Sprintf("\xff\033[%dm\xff%s\xff\033[0m\xff", c, wrap(o))
	}
}

func asciiAccessCode(o Outcome) string {
	switch o {
	case None:
		return "n/a"
	case Up:
		return "yes"
	case Down:
		return "no"
	case Err:
		return "ERR"
	default:
		panic("unknown access code")
	}
}

func Bold(in Renderable) Renderable {
	return RenderableFunc(func(e Env) string {
		inner := in.Render(e)
		if !e.IsTerminal {
			return inner
		}
		return fmt.Sprintf("\xff\033[1m%s\033[0m\xff", inner)
	})
}

func BoldText(str string) Renderable {
	return Bold(Text(str))
}

type RenderableFunc func(Env) string

func (f RenderableFunc) Render(e Env) string {
	return f(e)
}

func TextList(strs ...string) []Renderable {
	list := make([]Renderable, 0, len(strs))
	for _, str := range strs {
		list = append(list, Text(str))
	}
	return list
}
