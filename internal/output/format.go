// Package output provides functions for formatting and printing member data.
package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cli/go-gh/v2/pkg/jq"
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/template"
	"github.com/rmuraix/gh-member/internal/gh"
)

// PrintTable prints members as a table.
// When isTTY is true, output is column-aligned with headers; when false, output is TSV (no headers).
// maxWidth is the terminal width used only in TTY mode.
// Column headers: NAME / ID / ROLE / PROFILE
// When a member's Name is empty, Login is shown instead.
func PrintTable(w io.Writer, isTTY bool, maxWidth int, members []gh.Member) error {
	tp := tableprinter.New(w, isTTY, maxWidth)
	tp.AddHeader([]string{"NAME", "ID", "ROLE", "PROFILE"})
	for _, m := range members {
		name := m.Name
		if name == "" {
			name = m.Login
		}
		tp.AddField(name)
		tp.AddField(fmt.Sprintf("%d", m.DatabaseID))
		tp.AddField(m.Role)
		tp.AddField(m.URL)
		tp.EndRow()
	}
	return tp.Render()
}

// PrintJSON writes members as a JSON array to w.
// If fields is nil or empty (len==0), all fields are included.
// If fields is non-empty, only those fields are included.
// Valid field names: "name", "login", "role", "databaseId", "url".
// Invalid field names are silently ignored.
func PrintJSON(w io.Writer, members []gh.Member, fields []string) error {
	allFields := len(fields) == 0
	var result []any
	for _, m := range members {
		if allFields {
			result = append(result, map[string]any{
				"name":       m.Name,
				"login":      m.Login,
				"role":       m.Role,
				"databaseId": m.DatabaseID,
				"url":        m.URL,
			})
		} else {
			item := make(map[string]any, len(fields))
			for _, f := range fields {
				switch f {
				case "name":
					item["name"] = m.Name
				case "login":
					item["login"] = m.Login
				case "role":
					item["role"] = m.Role
				case "databaseId":
					item["databaseId"] = m.DatabaseID
				case "url":
					item["url"] = m.URL
				// unknown fields are silently ignored
				}
			}
			result = append(result, item)
		}
	}
	if result == nil {
		result = []any{}
	}
	enc := json.NewEncoder(w)
	// Disable HTML escaping so characters like '<', '>', '&' appear as-is in output.
	enc.SetEscapeHTML(false)
	return enc.Encode(result)
}

// ApplyJQ applies a jq filter to jsonInput and writes the result to w.
func ApplyJQ(w io.Writer, jsonInput io.Reader, jqFilter string) error {
	return jq.Evaluate(jsonInput, w, jqFilter)
}

// ApplyTemplate applies a Go template to jsonInput and writes the result to w.
func ApplyTemplate(w io.Writer, jsonInput io.Reader, goTemplate string) error {
	tmpl := template.New(w, 0, false)
	if err := tmpl.Parse(goTemplate); err != nil {
		return err
	}
	return tmpl.Execute(jsonInput)
}
