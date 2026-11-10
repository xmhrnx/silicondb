package main

import (
	"strings"

	"golang.org/x/net/html"
)

// text flattens an HTML document to visible text. Block-level elements end
// a line and inline elements are joined with spaces, so a label and its
// value in a spec table stay adjacent without touching.
func text(doc []byte) string {
	z := html.NewTokenizer(strings.NewReader(string(doc)))
	var sb strings.Builder
	skip := 0
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}
		switch tt {
		case html.StartTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "script", "style", "noscript", "svg":
				skip++
			case "br", "div", "p", "li", "tr", "td", "th", "dd", "dt", "h1", "h2", "h3", "h4", "section", "table":
				sb.WriteByte('|')
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			switch string(name) {
			case "script", "style", "noscript", "svg":
				if skip > 0 {
					skip--
				}
			case "div", "p", "li", "tr", "td", "th", "dd", "dt", "h1", "h2", "h3", "h4", "section", "table":
				sb.WriteByte('|')
			}
		case html.TextToken:
			if skip == 0 {
				sb.Write(z.Text())
			}
		}
	}
	return tidy(sb.String())
}

// tidy collapses whitespace and separators into a single line of
// "label|value|label|value|..." runs.
func tidy(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\u2009", " ")
	s = strings.ReplaceAll(s, "\u202f", " ")
	s = strings.Join(strings.Fields(s), " ")
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '|' {
			// Collapse runs of separators, dropping surrounding spaces.
			for sb.Len() > 0 && sb.String()[sb.Len()-1] == ' ' {
				t := sb.String()
				sb.Reset()
				sb.WriteString(t[:len(t)-1])
			}
			if sb.Len() == 0 || sb.String()[sb.Len()-1] != '|' {
				sb.WriteByte('|')
			}
			for i+1 < len(s) && (s[i+1] == '|' || s[i+1] == ' ') {
				i++
			}
			continue
		}
		sb.WriteByte(s[i])
	}
	return strings.Trim(sb.String(), "| ")
}

// table is an HTML table: its rows, each a list of cell texts. Cells with
// rowspan or colspan are repeated so column indexes line up with headers.
type table struct {
	rows [][]string
}

// tables extracts every <table> in the document.
func tables(doc []byte) []*table {
	root, err := html.Parse(strings.NewReader(string(doc)))
	if err != nil {
		return nil
	}
	var out []*table
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			// Wikipedia nests data tables inside layout tables, so keep
			// descending after collecting this one.
			if t := buildTable(n); t != nil {
				out = append(out, t)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return out
}

func buildTable(n *html.Node) *table {
	t := &table{}
	// pending holds cells that extend into lower rows from a rowspan.
	var pending []map[int]string
	var walkRows func(*html.Node)
	walkRows = func(node *html.Node) {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode {
				continue
			}
			switch c.Data {
			case "thead", "tbody", "tfoot":
				walkRows(c)
			case "tr":
				row := map[int]string{}
				if len(pending) > 0 {
					for col, v := range pending[0] {
						row[col] = v
					}
					pending = pending[1:]
				}
				col := 0
				for cell := c.FirstChild; cell != nil; cell = cell.NextSibling {
					if cell.Type != html.ElementNode || (cell.Data != "td" && cell.Data != "th") {
						continue
					}
					for row[col] != "" || occupiedBySpan(pending, col) {
						col++
					}
					v := cellText(cell)
					rowspan, colspan := 1, 1
					for _, a := range cell.Attr {
						if a.Key == "rowspan" {
							rowspan = atoiDefault(a.Val, 1)
						} else if a.Key == "colspan" {
							colspan = atoiDefault(a.Val, 1)
						}
					}
					for i := 0; i < colspan; i++ {
						row[col+i] = v
						if rowspan > 1 {
							for r := 0; r < rowspan-1; r++ {
								for len(pending) <= r {
									pending = append(pending, map[int]string{})
								}
								pending[r][col+i] = v
							}
						}
					}
					col += colspan
				}
				width := len(row)
				for _, m := range pending {
					for k := range m {
						if k+1 > width {
							width = k + 1
						}
					}
				}
				cells := make([]string, width)
				for i := 0; i < width; i++ {
					cells[i] = row[i]
				}
				t.rows = append(t.rows, cells)
			}
		}
	}
	walkRows(n)
	if len(t.rows) == 0 {
		return nil
	}
	return t
}

func occupiedBySpan(pending []map[int]string, col int) bool {
	for _, m := range pending {
		if _, ok := m[col]; ok {
			return true
		}
	}
	return false
}

// cellText returns a cell's text with tags flattened to single spaces.
func cellText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		if n.Type == html.ElementNode && n.Data == "br" {
			sb.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(strings.ReplaceAll(sb.String(), "\u00a0", " ")), " ")
}

func atoiDefault(s string, def int) int {
	n := 0
	seen := false
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
		seen = true
	}
	if !seen {
		return def
	}
	return n
}
