package guard

import (
	"luma/internal/parser"
	"luma/pkg/jid"
)

type AdminGuard struct {
	adminJIDs []string
}

func NewAdminGuard(admins []string) *AdminGuard {
	return &AdminGuard{adminJIDs: admins}
}

func (g *AdminGuard) Check(msg *parser.ParsedMessage) (bool, string) {
	for _, adminStr := range g.adminJIDs {
		adminJID := jid.Parse(adminStr)
		if jid.Compare(msg.From, adminJID) {
			return true, ""
		}
	}
	return false, "⛔ This command is for admins only."
}
