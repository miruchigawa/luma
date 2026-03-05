package guard

import (
	"luma/internal/parser"
)

type PrivateGuard struct{}

func NewPrivateGuard() *PrivateGuard {
	return &PrivateGuard{}
}

func (g *PrivateGuard) Check(msg *parser.ParsedMessage) (bool, string) {
	if !msg.IsGroup {
		return true, ""
	}
	return false, "⛔ This command only works in private chats."
}
