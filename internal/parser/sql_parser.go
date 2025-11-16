package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/rifqiakrm/starter-cli/internal/types"
)

// FindMigration finds the migration file for given schema + table
func FindMigration(root, schema, table string) (string, error) {
	var found string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// ensure it's inside the schema folder
		if strings.Contains(path, string(filepath.Separator)+schema+string(filepath.Separator)) &&
			strings.Contains(path, "_"+table) &&
			strings.HasSuffix(path, ".up.sql") {
			found = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("no migration for %s.%s", schema, table)
	}
	return found, nil
}

// ParseSQL parses a CREATE TABLE SQL file into a Table struct
func ParseSQL(path string) (*types.Table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := string(data)

	// find "create table" (case-insensitive)
	lower := strings.ToLower(s)
	ctIdx := strings.Index(lower, "create table")
	if ctIdx == -1 {
		return nil, fmt.Errorf("could not find CREATE TABLE in file")
	}

	// Find '(' that starts the column block after the CREATE TABLE
	openIdx := strings.Index(s[ctIdx:], "(")
	if openIdx == -1 {
		return nil, fmt.Errorf("could not find opening '(' after CREATE TABLE")
	}
	openIdx += ctIdx // make absolute index

	// Find matching closing ')' by scanning and counting parentheses
	level := 0
	closeIdx := -1
	for i := openIdx; i < len(s); i++ {
		ch := s[i]
		if ch == '(' {
			level++
		} else if ch == ')' {
			level--
			if level == 0 {
				closeIdx = i
				break
			}
		}
	}
	if closeIdx == -1 {
		return nil, fmt.Errorf("could not find matching ')' for CREATE TABLE")
	}

	// header contains everything between "CREATE TABLE" and '('
	header := s[ctIdx:openIdx]
	body := s[openIdx+1 : closeIdx] // column definitions and constraints

	// Extract name token from header
	headerLower := strings.ToLower(header)
	after := header
	if idx := strings.Index(headerLower, "create table"); idx != -1 {
		after = header[idx+len("create table"):]
	}
	after = strings.TrimSpace(after)
	if strings.HasPrefix(strings.ToLower(after), "if not exists") {
		after = strings.TrimSpace(after[len("if not exists"):])
	}

	// header name should now be something like:
	// - auth.users
	// - "auth"."users"
	// - users
	after = strings.TrimSpace(strings.ReplaceAll(after, "\n", " "))
	tokens := strings.Fields(after)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("could not parse table name from header: %q", header)
	}
	nameToken := tokens[0]

	// handle quoted identifiers and schema.table
	var schemaName, tableName string
	if strings.Contains(nameToken, ".") {
		parts := strings.SplitN(nameToken, ".", 2)
		schemaName = trimQuotes(parts[0])
		tableName = trimQuotes(parts[1])
	} else {
		tableName = trimQuotes(nameToken)
	}

	// split body into column/constraint lines safely
	colLines := splitColumns(body)

	var cols []types.Column
	for _, raw := range colLines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		upper := strings.ToUpper(line)

		// skip constraint lines
		if strings.HasPrefix(upper, "PRIMARY KEY") ||
			strings.HasPrefix(upper, "CONSTRAINT") ||
			strings.HasPrefix(upper, "UNIQUE") ||
			strings.HasPrefix(upper, "FOREIGN KEY") ||
			strings.HasPrefix(upper, "CHECK") {
			continue
		}

		// parse column name
		colName, rest, err := splitNameAndRest(line)
		if err != nil {
			// if can't parse, skip (safe)
			continue
		}

		// determine type: take everything until one of stop tokens
		stopTokens := []string{"not null", "null", "default", "primary", "unique", "references", "check", "constraint"}
		idx := indexOfAny(strings.ToLower(rest), stopTokens)
		var typePart string
		if idx >= 0 {
			typePart = strings.TrimSpace(rest[:idx])
		} else {
			typePart = strings.TrimSpace(rest)
		}

		nullability := true
		if strings.Contains(strings.ToUpper(rest), "NOT NULL") {
			nullability = false
		}
		isPrimary := strings.Contains(strings.ToUpper(rest), "PRIMARY KEY")

		cols = append(cols, types.Column{
			Name:       colName,
			Type:       strings.ToUpper(typePart),
			Nullable:   nullability,
			PrimaryKey: isPrimary,
		})
	}

	// Table name conversions for generator
	entityName := toCamel(naiveSingular(tableName))

	return &types.Table{
		Schema:    schemaName,
		Name:      tableName,
		NameUpper: entityName,
		NameLower: strings.ToLower(tableName),
		Columns:   cols,
		IsView:    false,
	}, nil
}

// Helper functions (copied from your original code)
func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '`' && s[len(s)-1] == '`')) {
		return s[1 : len(s)-1]
	}
	return s
}

func splitColumns(body string) []string {
	var out []string
	var sb strings.Builder
	level := 0
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c == '(' {
			level++
			sb.WriteByte(c)
			continue
		}
		if c == ')' {
			level--
			sb.WriteByte(c)
			continue
		}
		if c == ',' && level == 0 {
			out = append(out, sb.String())
			sb.Reset()
			continue
		}
		sb.WriteByte(c)
	}
	if s := strings.TrimSpace(sb.String()); s != "" {
		out = append(out, s)
	}
	return out
}

func splitNameAndRest(line string) (string, string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", fmt.Errorf("empty line")
	}
	if line[0] == '"' || line[0] == '`' {
		quote := line[0]
		// find closing quote
		j := -1
		for i := 1; i < len(line); i++ {
			if line[i] == quote {
				j = i
				break
			}
		}
		if j == -1 {
			return "", "", fmt.Errorf("unclosed quoted identifier")
		}
		name := line[1:j]
		rest := strings.TrimSpace(line[j+1:])
		return name, rest, nil
	}

	// unquoted: first token is name
	i := 0
	for i < len(line) && !unicode.IsSpace(rune(line[i])) {
		i++
	}
	if i == 0 {
		return "", "", fmt.Errorf("cannot parse column name")
	}
	name := line[:i]
	rest := strings.TrimSpace(line[i:])
	return name, rest, nil
}

func indexOfAny(s string, words []string) int {
	low := -1
	ls := strings.ToLower(s)
	for _, w := range words {
		if idx := strings.Index(ls, w); idx >= 0 {
			if low == -1 || idx < low {
				low = idx
			}
		}
	}
	return low
}

func toCamel(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, "")
}

func naiveSingular(word string) string {
	word = strings.ToLower(word)

	// Handle irregulars / special cases
	irregulars := map[string]string{
		"people":   "person",
		"men":      "man",
		"women":    "woman",
		"children": "child",
		"teeth":    "tooth",
		"feet":     "foot",
		"geese":    "goose",
		"mice":     "mouse",
		"data":     "datum",
		"indices":  "index",
		"matrices": "matrix",
		"vertices": "vertex",
		"statuses": "status",
		"courses":  "course",
		"buses":    "bus",
	}
	if val, ok := irregulars[word]; ok {
		return val
	}

	// Handle common suffixes
	switch {
	case strings.HasSuffix(word, "ies"):
		return word[:len(word)-3] + "y"
	case strings.HasSuffix(word, "ves"):
		base := word[:len(word)-3]
		if strings.HasSuffix(base, "f") {
			return base
		}
		return base + "fe"
	case strings.HasSuffix(word, "xes"),
		strings.HasSuffix(word, "ches"),
		strings.HasSuffix(word, "shes"):
		return word[:len(word)-2]
	case strings.HasSuffix(word, "zes"):
		return word[:len(word)-3]
	case strings.HasSuffix(word, "oes"):
		return word[:len(word)-2]
	case strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss"):
		return word[:len(word)-1]
	default:
		return word
	}
}
