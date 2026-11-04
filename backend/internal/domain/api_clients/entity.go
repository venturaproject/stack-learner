package apiclients

import "time"

const (
	ScopeTrabajadoresRead = "trabajadores:read"
	ScopeTelefonosRead    = "telefonos:read"
	ScopeDispositivosRead = "dispositivos:read"
	ScopeAll              = "*"
)

var AllScopes = []string{
	ScopeTrabajadoresRead,
	ScopeTelefonosRead,
	ScopeDispositivosRead,
}

type ApiClient struct {
	ID         string
	Name       string
	ClientID   string
	Scopes     []string
	Active     bool
	LastUsedAt *time.Time
	CreatedAt  time.Time
}

func (c *ApiClient) HasScope(scope string) bool {
	for _, s := range c.Scopes {
		if s == ScopeAll || s == scope {
			return true
		}
	}
	return false
}

type ApiClientToken struct {
	ID               string
	ClientID         string
	Scopes           []string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	CreatedAt        time.Time
}

// ExpandScopes replaces ["*"] with the full scope list.
func ExpandScopes(scopes []string) []string {
	for _, s := range scopes {
		if s == ScopeAll {
			out := make([]string, len(AllScopes))
			copy(out, AllScopes)
			return out
		}
	}
	return scopes
}
