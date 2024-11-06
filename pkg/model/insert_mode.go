package model

import "strings"

// InsertMode defines the type of insert that will be performed.
type InsertMode string

const (
	InsertModeInsert   InsertMode = "insert"
	InsertModeConflict InsertMode = "conflict"
	InsertModeUpsert   InsertMode = "upsert"
	InsertModeInvalid  InsertMode = "INVALID"
)

// ParseInsertMode takes a string and returns the corresponding InsertMode
// or invalid.
func ParseInsertMode(raw string) InsertMode {
	switch strings.ToLower(raw) {
	case "insert":
		return InsertModeInsert
	case "conflict":
		return InsertModeConflict
	case "upsert":
		return InsertModeUpsert
	default:
		return InsertModeInvalid
	}
}
