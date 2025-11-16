package types

// Column metadata
type Column struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
}

// Table metadata
type Table struct {
	Schema    string
	Name      string
	NameUpper string
	NameLower string
	Columns   []Column
	IsView    bool
}

// ModulePart defines which module components to generate
type ModulePart struct {
	Component string // "handler", "service", "repository"
	Action    string // "creator", "finder", "updater", "deleter", or "" for all
}

func (mp ModulePart) String() string {
	if mp.Action != "" {
		return mp.Component + "." + mp.Action
	}
	return mp.Component
}
