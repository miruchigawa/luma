package jid

import (
	"go.mau.fi/whatsmeow/types"
)

// Parse converts a string to a types.JID, ignoring errors.
func Parse(s string) types.JID {
	j, _ := types.ParseJID(s)
	return j
}

// Compare checks if two JIDs represent the same entity (ignoring devices).
func Compare(a, b types.JID) bool {
	return a.ToNonAD().String() == b.ToNonAD().String()
}
