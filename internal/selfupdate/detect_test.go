package selfupdate

import "testing"

func TestMethod(t *testing.T) {
	cases := []struct {
		name string
		path string
		want Install
	}{
		{"npm global", "/usr/lib/node_modules/@runaic/aic/bin/aic", Managed},
		{"npm nested", "/home/u/proj/node_modules/@runaic/aic/bin/aic", Managed},
		{"usr local", "/usr/local/bin/aic", Direct},
		{"home local", "/home/u/.local/bin/aic", Direct},
		{"substring not segment", "/opt/node_modules_backup/bin/aic", Direct},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Method(c.path); got != c.want {
				t.Errorf("Method(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}
