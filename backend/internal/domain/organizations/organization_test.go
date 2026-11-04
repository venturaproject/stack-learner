package organizations

import "testing"

func TestNormalizeSlug(t *testing.T) {
	if got := NormalizeSlug("  Academia Horizonte! "); got != "academia-horizonte" {
		t.Fatalf("NormalizeSlug() = %q", got)
	}
}

func TestValidateRole(t *testing.T) {
	for _, role := range []string{RoleAdmin, RoleTrainer, RoleStudent} {
		if err := ValidateRole(role); err != nil {
			t.Errorf("ValidateRole(%q): %v", role, err)
		}
	}
	if err := ValidateRole("platform_admin"); err == nil {
		t.Fatal("expected unsupported role to be rejected")
	}
}
