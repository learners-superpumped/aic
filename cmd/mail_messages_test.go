package cmd

import "testing"

func TestSafeAttachmentName(t *testing.T) {
	// A traversal/absolute path is reduced to its basename — it lands in the cwd
	// and cannot escape, so the download still succeeds with a safe name.
	ok := []struct{ in, want string }{
		{"report.pdf", "report.pdf"},
		{"weird name.png", "weird name.png"},
		{"답장.txt", "답장.txt"},
		{"../../etc/passwd", "passwd"},
		{"../x", "x"},
		{"/etc/passwd", "passwd"},
		{"sub/dir.txt", "dir.txt"},
	}
	for _, c := range ok {
		got, err := safeAttachmentName(c.in)
		if err != nil {
			t.Errorf("safeAttachmentName(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("safeAttachmentName(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	// Names with no usable basename, or that carry a backslash (a Windows path
	// separator), are rejected rather than guessed.
	bad := []string{
		"..",
		".",
		"/",
		"",
		`..\..\x`,
		`dir\file`,
	}
	for _, in := range bad {
		if got, err := safeAttachmentName(in); err == nil {
			t.Errorf("safeAttachmentName(%q) = %q, want error", in, got)
		}
	}
}
