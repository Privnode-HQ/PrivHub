package model

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// TokenGroups is an ordered list of groups for a token.
// It is stored in DB as a JSON array.
type TokenGroups []string

func (g TokenGroups) Value() (driver.Value, error) {
	return common.Marshal(&g)
}

func (g *TokenGroups) Scan(value interface{}) error {
	if value == nil {
		*g = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*g = nil
			return nil
		}
		return common.Unmarshal(v, g)
	case string:
		if v == "" {
			*g = nil
			return nil
		}
		return common.Unmarshal([]byte(v), g)
	default:
		return fmt.Errorf("unsupported TokenGroups Scan type: %T", value)
	}
}

func normalizeGroupName(group string) string {
	return strings.TrimSpace(group)
}

func (token *Token) HasGroupOverride() bool {
	if token == nil {
		return false
	}
	return token.Group != "" || len(token.Groups) > 0
}

func (token *Token) PrimaryGroup() string {
	if token == nil {
		return ""
	}
	if len(token.Groups) > 0 {
		return normalizeGroupName(token.Groups[0])
	}
	return normalizeGroupName(token.Group)
}

// GetOrderedGroups returns the token's ordered group candidates.
// - If token.Groups is set, it is used.
// - Otherwise, token.Group is treated as a single group.
func (token *Token) GetOrderedGroups() []string {
	if token == nil {
		return nil
	}
	if len(token.Groups) > 0 {
		out := make([]string, 0, len(token.Groups))
		for _, g := range token.Groups {
			g = normalizeGroupName(g)
			if g == "" {
				continue
			}
			out = append(out, g)
		}
		return out
	}
	if token.Group == "" {
		return nil
	}
	return []string{normalizeGroupName(token.Group)}
}
