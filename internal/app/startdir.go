package app

import (
	"os"
	"path/filepath"
	"strings"
)

// startDir is where a tab opened without a directory starts: the Shell
// page's Working Directory, a leading ~ standing for the home directory,
// as the page says ("Where new sessions start. Empty uses your home
// directory."). A setting that names no directory starts at home too: a
// shell cannot start in a directory that is not there, and the tab would
// open with an error in place of a prompt.
func (a *app) startDir() string {
	dir := strings.TrimSpace(a.settings.WorkingDirectory)
	if dir == "~" || strings.HasPrefix(dir, "~/") {
		dir = filepath.Join(homeDir(), dir[1:])
	}
	if st, err := os.Stat(dir); dir == "" || err != nil || !st.IsDir() {
		return homeDir()
	}
	return dir
}
