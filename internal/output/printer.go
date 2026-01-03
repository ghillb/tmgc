package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Mode string

const (
	ModeHuman Mode = "human"
	ModePlain Mode = "plain"
	ModeJSON  Mode = "json"
)

type Printer struct {
	Out     io.Writer
	Err     io.Writer
	Mode    Mode
	NoColor bool
}

func NewPrinter(out io.Writer, err io.Writer, mode Mode, noColor bool) *Printer {
	return &Printer{
		Out:     out,
		Err:     err,
		Mode:    mode,
		NoColor: noColor,
	}
}

func (p *Printer) JSON(v any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (p *Printer) Plain(lines []string) {
	for _, line := range lines {
		fmt.Fprintln(p.Out, line)
	}
}

func (p *Printer) Table(rows [][]string) {
	w := tabwriter.NewWriter(p.Out, 0, 4, 2, ' ', 0)
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	_ = w.Flush()
}

func (p *Printer) Logf(format string, args ...any) {
	fmt.Fprintf(p.Err, format, args...)
}

func (p *Printer) Logln(args ...any) {
	fmt.Fprintln(p.Err, args...)
}
