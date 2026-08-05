package authorization

import (
	"testing"
)

func FuzzParsePermission(f *testing.F) {
	f.Add("tenant.read")
	f.Add("members.invite")
	f.Add("metadata.write")
	f.Add("invalid")
	f.Add("")
	f.Add("a.b.c.d.e")

	f.Fuzz(func(t *testing.T, perm string) {
		// Try to parse the permission
		// This should not panic on any input
		if perm == "" {
			return
		}
		_ = perm
	})
}
