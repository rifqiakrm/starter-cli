package generator

import (
	"strings"

	"github.com/rifqiakrm/starter-cli/internal/types"
)

// TemplateData holds data passed to templates
type TemplateData struct {
	Schema          string
	Version         string
	EntityCamelCase string
	EntityLower     string
	EntityUpper     string
}

// createTemplateData creates template data from entity name
func (g *Generator) createTemplateData(schema, entity, version string) TemplateData {
	singular := singularize(entity)
	return TemplateData{
		Schema:          schema,
		Version:         version,
		EntityCamelCase: toCamelCase(singular),
		EntityLower:     strings.ToLower(singular),
		EntityUpper:     toPascalCase(singular),
	}
}

// getActions returns specific actions or all if empty
func getActions(action string) []string {
	if action == "" {
		return []string{"creator", "finder", "updater", "deleter"}
	}
	return []string{action}
}

// toCamelCase converts snake_case to camelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			if i == 0 {
				parts[i] = strings.ToLower(p)
			} else {
				parts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
	}
	return strings.Join(parts, "")
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			// Handle special cases like id, ip, url, etc.
			word := strings.ToLower(p)
			switch word {
			case "id":
				parts[i] = "ID"
			case "ip":
				parts[i] = "IP"
			case "url":
				parts[i] = "URL"
			case "api":
				parts[i] = "API"
			case "http":
				parts[i] = "HTTP"
			case "html":
				parts[i] = "HTML"
			case "json":
				parts[i] = "JSON"
			case "xml":
				parts[i] = "XML"
			default:
				parts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
	}
	return strings.Join(parts, "")
}

func toLowerCamelCase(s string) string {
	camel := toCamelCase(s)
	if len(camel) > 0 {
		camel = strings.ToLower(camel[:1]) + camel[1:]
	}
	return camel
}

// singularize converts plural nouns to singular form with common rules.
// It’s not perfect like inflection libraries, but good enough for DB tables.
func singularize(word string) string {
	word = strings.ToLower(word)

	// Handle common special cases
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
		"quizzes":  "quiz",
	}
	if val, ok := irregulars[word]; ok {
		return val
	}

	// Handle endings
	switch {
	case strings.HasSuffix(word, "ies"):
		// companies -> company
		return word[:len(word)-3] + "y"
	case strings.HasSuffix(word, "ves"):
		// wolves -> wolf, knives -> knife
		base := word[:len(word)-3]
		if strings.HasSuffix(base, "f") {
			return base
		}
		return base + "fe"
	case strings.HasSuffix(word, "xes"),
		strings.HasSuffix(word, "ches"),
		strings.HasSuffix(word, "shes"):
		// boxes -> box, addresses -> address
		return word[:len(word)-2]
	case strings.HasSuffix(word, "zes"):
		// quizzes -> quiz
		return word[:len(word)-3]
	case strings.HasSuffix(word, "oes"):
		// tomatoes -> tomato, heroes -> hero
		return word[:len(word)-2]
	case strings.HasSuffix(word, "s") && len(word) > 1:
		// generic fallback: users -> user, tables -> table
		return word[:len(word)-1]
	}

	return word
}

func goType(col types.Column) string {
	sqlType := strings.ToUpper(col.Type)
	nullable := col.Nullable
	switch {
	case strings.HasPrefix(sqlType, "UUID"):
		return "uuid.UUID"
	case strings.HasPrefix(sqlType, "VARCHAR"), strings.HasPrefix(sqlType, "TEXT"):
		if nullable {
			return "sql.NullString"
		}
		return "string"
	case strings.HasPrefix(sqlType, "DATE"), strings.HasPrefix(sqlType, "TIMESTAMPTZ"):
		if nullable {
			return "sql.NullTime"
		}
		return "time.Time"
	case strings.HasPrefix(sqlType, "INT"):
		if nullable {
			return "sql.NullInt64"
		}
		return "int64"
	case strings.HasPrefix(sqlType, "SERIAL"):
		if nullable {
			return "sql.NullInt64"
		}
		return "int64"
	default:
		return "string"
	}
}

func isAuditable(name string) bool {
	switch strings.ToLower(name) {
	case "created_at", "updated_at", "deleted_at",
		"created_by", "updated_by", "deleted_by":
		return true
	}
	return false
}

// Add these resource template helper functions
func isSensitive(name string) bool {
	n := strings.ToLower(name)
	switch n {
	case "password", "otp", "forgot_password_token",
		"created_by", "updated_by", "deleted_by", "deleted_at":
		return true
	default:
		return false
	}
}

func includeInCreate(col types.Column) bool {
	lower := strings.ToLower(col.Name)
	if col.PrimaryKey || lower == "created_at" || lower == "updated_at" || lower == "deleted_at" {
		return false
	}
	return !isSensitive(col.Name)
}

func includeInUpdate(col types.Column) bool {
	lower := strings.ToLower(col.Name)
	if lower == "id" || lower == "created_at" || lower == "updated_at" || lower == "deleted_at" {
		return false
	}
	return !isSensitive(col.Name)
}

func goResourceType(col types.Column) string {
	sqlType := strings.ToUpper(col.Type)
	switch {
	case strings.HasPrefix(sqlType, "UUID"):
		return "string"
	case strings.HasPrefix(sqlType, "VARCHAR"), strings.HasPrefix(sqlType, "TEXT"):
		return "string"
	case strings.HasPrefix(sqlType, "DATE"), strings.HasPrefix(sqlType, "TIMESTAMPTZ"), strings.HasPrefix(sqlType, "TIMESTAMP"):
		return "string"
	case strings.HasPrefix(sqlType, "INT"), strings.HasPrefix(sqlType, "SERIAL"):
		return "int64"
	default:
		return "string"
	}
}

func goRequestType(col types.Column, required bool) string {
	typ := "string"
	sqlType := strings.ToUpper(col.Type)
	switch {
	case strings.HasPrefix(sqlType, "UUID"):
		typ = "string"
	case strings.HasPrefix(sqlType, "VARCHAR"), strings.HasPrefix(sqlType, "TEXT"):
		typ = "string"
	case strings.HasPrefix(sqlType, "DATE"), strings.HasPrefix(sqlType, "TIMESTAMP"), strings.HasPrefix(sqlType, "TIMESTAMPTZ"):
		typ = "string"
	case strings.HasPrefix(sqlType, "INT"), strings.HasPrefix(sqlType, "SERIAL"):
		typ = "int64"
	}
	return typ
}

func mapFromEntity(col types.Column) string {
	pascal := toPascalCase(col.Name)
	sqlType := strings.ToUpper(col.Type)
	lowerName := strings.ToLower(col.Name)

	switch {
	case strings.HasPrefix(sqlType, "UUID"):
		if !col.Nullable || col.PrimaryKey {
			return "e." + pascal + ".String()"
		}
		return "e." + pascal + ".String"

	case strings.HasPrefix(sqlType, "VARCHAR"), strings.HasPrefix(sqlType, "TEXT"):
		if !col.Nullable {
			return "e." + pascal
		}
		return "e." + pascal + ".String"

	case strings.HasPrefix(sqlType, "DATE"), strings.HasPrefix(sqlType, "TIMESTAMPTZ"), strings.HasPrefix(sqlType, "TIMESTAMP"):
		if lowerName == "created_at" || lowerName == "updated_at" || !col.Nullable {
			return "e." + pascal + ".Format(constant.DefaultTimeFormat)"
		}
		return "e." + pascal + ".Time.Format(constant.DefaultTimeFormat)"

	case strings.HasPrefix(sqlType, "INT"), strings.HasPrefix(sqlType, "SERIAL"):
		return "e." + pascal + ".Int64"

	default:
		return "e." + pascal + ".String"
	}
}
