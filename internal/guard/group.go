package guard

import (
	"luma/internal/parser"
)

type GroupGuard struct{}

func NewGroupGuard() *GroupGuard {
	return &GroupGuard{}
}

func (g *GroupGuard) Check(msg *parser.ParsedMessage) (bool, string) {
	if msg.IsGroup {
		return true, ""
	}
	return false, "⛔ This command only works in group chats."
}
