package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
	"github.com/smallnest/uix/components/tree"
)

// infoTab is the section of the info panel that is open.
type infoTab int

const (
	infoTabInfo infoTab = iota
	infoTabOutline
	infoTabGit
	infoTabFiles
)

// infoState is the info panel's data: the open section and each section's
// loaded state, one section at a time, as otty shows them.
type infoState struct {
	tab         infoTab
	ports       []string
	portsDone   bool
	outline     []string
	outlineDone bool
	git         gitState
	files       []tree.Item
	filesDone   bool
}

// gitState is the working directory's repository, as git reports it, and
// the inline diff the Git section shows for one changed file.
type gitState struct {
	branch    string
	changes   []gitChange
	done      bool
	diffOpen  string // the path whose diff is open, or ""
	diffLines []string
}

// gitChange is one changed file: its porcelain status and its path. The
// status keeps the raw two-letter code, so a staged "M " and an unstaged
// " M" stay distinct for the diff.
type gitChange struct {
	status string // e.g. " M", "M ", "??", "A "
	path   string
}

// infoView is the panel on the right with what otty shows there. Its
// toolbar lives in the top strip; here are only the sections, the ones
// of the open tab. The Info tab holds the working directory, the running
// process and the ports; the Outline tab the shell's command marks; the
// Git tab the repository's status; the Files tab the working tree.
func (a *app) infoView(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(a.detailsPanelWidth()).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		a.infoBody(c)
	})
}

// infoColumn is the info panel of the Tabs layouts: a full-height column
// that reaches the top of the window, with the mode switch at the strip's
// height and the sections below, as otty draws the panel beside the
// horizontal tab bar.
func (a *app) infoColumn(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Width(a.detailsPanelWidth()).FillHeight().Shrink(0).Background(t.Surface).
		BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		ui.Box(c).FillWidth().Shrink(0).Height(a.stripBar(c)).Children(func() {
			a.infoToolbarRow(c)
		})
		a.infoBody(c)
	})
}

// infoBody is the panel's sections, the ones of the open tab.
func (a *app) infoBody(c *ui.Context) {
	ui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).FillWidth().Padding(14, 18).Gap(20).Children(func() {
			switch a.info.tab {
			case infoTabOutline:
				a.infoOutlineView(c)
			case infoTabGit:
				a.infoGitView(c)
			case infoTabFiles:
				a.infoFilesView(c)
			default:
				a.infoDirectory(c)
				a.infoProcess(c)
				a.infoPorts(c)
			}
		})
	})
}

// setInfoTab opens the info panel on a section; the section's data loads
// as the panel opens.
func (a *app) setInfoTab(t infoTab) {
	a.info.tab = t
	a.setPanel(panelInfo)
}

// infoTabID is the details view id of an info tab, the id the settings'
// Details Views list holds.
func infoTabID(t infoTab) string {
	switch t {
	case infoTabOutline:
		return "outline"
	case infoTabGit:
		return "git"
	case infoTabFiles:
		return "files"
	}
	return "info"
}

// detailsViewEnabled reports whether a details view shows, per the View
// page's Details Panel settings; an id without a setting shows.
func (a *app) detailsViewEnabled(id string) bool {
	if a.settings == nil {
		return true
	}
	for _, v := range a.settings.DetailsViews {
		if v.ID == id {
			return v.Enabled
		}
	}
	return true
}

// loadInfoTab loads the open section's data. A section the settings hide
// falls back to Info. The Info section rescans the ports each time it
// opens; the others load once and keep their state.
func (a *app) loadInfoTab() {
	if !a.detailsViewEnabled(infoTabID(a.info.tab)) {
		a.info.tab = infoTabInfo
	}
	switch a.info.tab {
	case infoTabInfo:
		a.loadPorts()
	case infoTabOutline:
		if !a.info.outlineDone {
			a.loadOutline()
		}
	case infoTabGit:
		if !a.info.git.done {
			a.loadGit()
		}
	case infoTabFiles:
		if !a.info.filesDone {
			a.loadFiles()
		}
	}
}

// infoDirChanged marks the info panel's directory-dependent sections stale
// after the working directory changes, and reloads the one that is open,
// so the Git and Files panels follow the shell into the new directory
// instead of keeping the previous one's state. The other sections (Info,
// Outline) read live state and need no invalidation.
func (a *app) infoDirChanged() {
	a.info.git.done = false
	a.info.git.branch = ""
	a.info.git.changes = nil
	a.info.git.diffOpen = ""
	a.info.git.diffLines = nil
	a.info.filesDone = false
	a.info.files = nil
	if a.info.tab == infoTabGit || a.info.tab == infoTabFiles {
		a.loadInfoTab()
	}
}

// infoSection is one titled block of the info panel. The title is plain
// text, small and muted, as otty draws its section headers.
func (a *app) infoSection(c *ui.Context, title string, fn func()) {
	t := c.Theme()
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		ui.Text(c, title).FontSize(11).TextColor(t.TextMuted)
		fn()
	})
}

func (a *app) infoDirectory(c *ui.Context) {
	t := c.Theme()
	s := a.currentSession()
	a.infoSection(c, tr("Working Directory"), func() {
		if s == nil {
			ui.Text(c, "—").TextColor(t.TextMuted)
			return
		}
		ui.Column(c).FillWidth().Gap(12).Children(func() {
			ui.Text(c, homeShown(s.dir)).SingleLine().FontSize(12).TextColor(t.Text)
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				a.infoAction(c, tr("Copy Path"), lucide.Copy(), func() {
					mygo.Clipboard.WriteText(s.dir)
				})
				a.infoAction(c, tr("Reveal in Finder"), lucide.FolderOpen(), func() {
					_ = exec.Command("open", s.dir).Start()
				})
				a.infoAction(c, tr("Open in Zed"), lucide.ArrowUpRight(), func() {
					openZed(s.dir)
				})
			})
		})
	})
}

func (a *app) infoProcess(c *ui.Context) {
	t := c.Theme()
	a.infoSection(c, tr("Process"), func() {
		// Otty lists the children of the session's shell here. MyGo's
		// terminal plugin exposes no child process, so show the idle
		// state, as otty does when nothing runs.
		ui.Text(c, tr("No child processes")).FontSize(11).TextColor(t.TextMuted)
	})
}

func (a *app) infoPorts(c *ui.Context) {
	t := c.Theme()
	a.infoSection(c, tr("Ports"), func() {
		switch {
		case !a.info.portsDone:
			ui.Text(c, tr("Scanning…")).TextColor(t.TextMuted)
		case len(a.info.ports) == 0:
			ui.Text(c, tr("No listening ports")).TextColor(t.TextMuted)
		default:
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				for _, p := range a.info.ports {
					ui.Text(c, p).FontSize(11).TextColor(t.Text)
				}
			})
		}
	})
}

// infoAction is one row of the working directory's actions: an icon and a
// green label, as otty draws them.
func (a *app) infoAction(c *ui.Context, label string, icon *ui.SVG, onClick func()) {
	t := c.Theme()
	b := ui.ButtonBase(c.Key("info-"+strings.ReplaceAll(label, " ", "-"))).
		FillWidth().Padding(5, 0).Radius(t.Radius).FocusRing(false)
	b.Children(func() {
		// The row fills the button so the label starts at the panel's left
		// edge; a shrink-to-fit row would center on the button.
		ui.Row(c).FillWidth().Gap(7).Children(func() {
			ui.Image(c, icon).Size(14, 14)
			ui.Text(c, label).FontSize(12).TextColor(actionGreen(t))
		})
	})
	if b.Clicked() {
		onClick()
	}
}

// actionGreen is the green of the info panel's action rows: a muted green
// on light, a brighter one on dark, as otty colors them.
func actionGreen(t *ui.Theme) ui.Color {
	if t.Dark {
		return ui.Hex("#7fbf8a")
	}
	return ui.Hex("#78877C")
}

// loadPorts scans the ports the window's processes listen on, in the
// background, and puts them on the panel when the scan ends.
func (a *app) loadPorts() {
	if a.win == nil {
		return
	}
	a.info.ports = nil
	a.info.portsDone = false
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN").Output()
		var ports []string
		if err == nil {
			seen := map[string]bool{}
			for _, line := range strings.Split(string(out), "\n") {
				f := strings.Fields(line)
				if len(f) < 9 {
					continue
				}
				if p := portOf(f[8]); p != "" && !seen[p] {
					seen[p] = true
					ports = append(ports, p)
				}
			}
			sort.Strings(ports)
		}
		a.win.Update(func() {
			a.info.ports = ports
			a.info.portsDone = true
		})
	}()
}

// portOf extracts the port from an lsof name like "*:8000 (LISTEN)" or
// "[::1]:8000 (LISTEN)", or "" when there is none.
func portOf(name string) string {
	name = strings.TrimSuffix(strings.TrimSpace(name), " (LISTEN)")
	if i := strings.LastIndex(name, "]:"); i >= 0 { // IPv6
		return name[i+2:]
	}
	if i := strings.LastIndex(name, ":"); i >= 0 {
		p := name[i+1:]
		if p != "" && p != "*" {
			return p
		}
	}
	return ""
}

// openZed opens dir in Zed, when it is installed.
func openZed(dir string) {
	if _, err := os.Stat("/Applications/Zed.app"); err != nil {
		return
	}
	_ = exec.Command("open", "-a", "Zed", dir).Start()
}

// loadOutline reads the shell's history for the command marks the Outline
// section lists, as otty lists a pane's command marks. zsh writes each
// command as ": <epoch>:<duration>;<command>"; the last 200 commands are
// shown, newest first.
func (a *app) loadOutline() {
	if a.win == nil {
		return
	}
	a.info.outline = nil
	go func() {
		var cmds []string
		hf := os.Getenv("HISTFILE")
		if hf == "" {
			hf = filepath.Join(homeDir(), ".zsh_history")
		}
		if b, err := os.ReadFile(hf); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				// Strip the ": <epoch>:<duration>;" stamp when there is one.
				if i := strings.LastIndex(line, ";"); strings.HasPrefix(line, ": ") && i >= 0 {
					line = line[i+1:]
				}
				if line = strings.TrimSpace(line); line != "" {
					cmds = append(cmds, line)
				}
			}
			if len(cmds) > 200 {
				cmds = cmds[len(cmds)-200:]
			}
			slices.Reverse(cmds)
		}
		a.win.Update(func() {
			a.info.outline = cmds
			a.info.outlineDone = true
		})
	}()
}

// infoOutlineView is the Outline section: the shell's command marks, one
// row each, newest first, as otty lists them for a terminal pane.
func (a *app) infoOutlineView(c *ui.Context) {
	t := c.Theme()
	if !a.info.outlineDone {
		ui.Text(c, tr("Reading…")).TextColor(t.TextMuted)
		return
	}
	a.infoSection(c, tr("Command Marks"), func() {
		if len(a.info.outline) == 0 {
			ui.Text(c, tr("No commands yet")).TextColor(t.TextMuted)
			return
		}
		ui.Column(c).FillWidth().Gap(4).Children(func() {
			for _, cmd := range a.info.outline {
				ui.Text(c, cmd).SingleLine().FontSize(11).TextColor(t.Text)
			}
		})
	})
}

// loadGit runs git status in the working directory for the Git section:
// the branch and the changed files, as otty's Git tab shows them. An open
// inline diff stays open when the status refreshes.
func (a *app) loadGit() {
	if a.win == nil {
		return
	}
	a.info.git.done = false
	go func() {
		s := a.currentSession()
		dir := ""
		var branch string
		var changes []gitChange
		if s != nil {
			dir = s.dir
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, "git", "-C", dir,
				"status", "--porcelain=v1", "-b").Output()
			if err == nil {
				branch, changes = parseGitStatus(string(out))
			}
		}
		a.win.Update(func() {
			// A directory changed mid-load: the newer load owns the panel,
			// and this stale one must not commit over it.
			if staleInfoCommit(a.currentSession(), s, dir) {
				return
			}
			keep := a.info.git.diffOpen
			a.info.git.branch = branch
			a.info.git.changes = changes
			a.info.git.done = true
			a.info.git.diffOpen = keep
		})
	}()
}

// staleInfoCommit reports whether a loader's commit must be dropped: the
// active session changed while the loader ran, or the session's directory
// moved on, so a newer load owns the panel's data.
func staleInfoCommit(cur, loaded *session, dir string) bool {
	if cur != loaded {
		return true
	}
	return loaded != nil && loaded.dir != dir
}

// parseGitStatus reads git status --porcelain=v1 -b output: the branch
// from the "## " header, and one change per other line, keeping the raw
// two-letter status. A detached HEAD reports the branch as "HEAD".
func parseGitStatus(out string) (branch string, changes []gitChange) {
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			branch = strings.TrimPrefix(strings.Fields(line)[1], "#")
			if i := strings.Index(branch, "..."); i >= 0 {
				branch = branch[:i]
			}
			continue
		}
		changes = append(changes, gitChange{
			status: line[:2],
			path:   strings.TrimSpace(line[3:]),
		})
	}
	return branch, changes
}

// loadGitDiff loads the diff of one changed file into the Git section. A
// staged change diffs against the index; an untracked file has no diff.
func (a *app) loadGitDiff(path string, staged bool) {
	if a.win == nil {
		return
	}
	a.info.git.diffOpen = path
	a.info.git.diffLines = nil
	go func() {
		s := a.currentSession()
		var lines []string
		if s != nil {
			args := []string{"-C", s.dir, "diff"}
			if staged {
				args = append(args, "--cached")
			}
			args = append(args, "--", path)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if out, err := exec.CommandContext(ctx, "git", args...).Output(); err == nil {
				lines = strings.Split(strings.TrimRight(string(out), "\n"), "\n")
			}
		}
		a.win.Update(func() {
			a.info.git.diffLines = lines
		})
	}()
}

// infoGitView is the Git section: the branch, then the changed files with
// an inline diff that opens under a file when its row is clicked, as
// otty's Git tab shows them.
func (a *app) infoGitView(c *ui.Context) {
	t := c.Theme()
	if !a.info.git.done {
		ui.Text(c, tr("Reading…")).TextColor(t.TextMuted)
		return
	}
	a.infoSection(c, tr("Repository"), func() {
		if a.info.git.branch == "" {
			ui.Text(c, tr("Not a repository")).TextColor(t.TextMuted)
			return
		}
		ui.Text(c, a.info.git.branch).FontSize(12).TextColor(t.Text)
	})
	a.infoSection(c, tr("Changes"), func() {
		if len(a.info.git.changes) == 0 {
			ui.Text(c, tr("No changes")).TextColor(t.TextMuted)
			return
		}
		ui.Column(c).FillWidth().Gap(4).Children(func() {
			for _, ch := range a.info.git.changes {
				a.gitChangeRow(c, ch)
			}
		})
	})
}

// gitChangeRow is one changed file of the Git section: its status letter
// and path, and the inline diff below it while it is open. Clicking the
// row toggles the diff, as otty does in its Git tab.
func (a *app) gitChangeRow(c *ui.Context, ch gitChange) {
	t := c.Theme()
	open := a.info.git.diffOpen == ch.path
	row := ui.ButtonBase(c.Key("git-"+ch.path)).FillWidth().Padding(4, 0).
		Radius(t.Radius).FocusRing(false)
	row.Children(func() {
		ui.Row(c).FillWidth().Gap(6).Children(func() {
			ui.Text(c, gitStatusLabel(ch.status)).SingleLine().FontSize(11).TextColor(gitStatusColor(t, ch.status))
			ui.Text(c, ch.path).SingleLine().FontSize(11).TextColor(t.Text).Grow(1).MinWidth(0)
		})
	})
	if row.Clicked() {
		if open {
			a.info.git.diffOpen = ""
			a.info.git.diffLines = nil
		} else if ch.status == "??" {
			// An untracked file has no committed diff to show.
			a.info.git.diffOpen = ch.path
			a.info.git.diffLines = []string{tr("untracked — no diff yet")}
		} else if len(ch.status) > 0 && ch.status[0] != ' ' && ch.status[0] != '?' {
			a.loadGitDiff(ch.path, true)
		} else {
			a.loadGitDiff(ch.path, false)
		}
	}
	if open {
		ui.Column(c).FillWidth().Gap(2).Padding(0, 6, 0, 0).Children(func() {
			if a.info.git.diffLines == nil {
				ui.Text(c, tr("Reading…")).FontSize(11).TextColor(t.TextMuted)
			}
			for _, line := range a.info.git.diffLines {
				ui.Text(c, line).FontSize(11).TextColor(diffColor(t, line))
			}
		})
	}
}

// gitStatusLabel is the letter a change shows: the staged letter when
// there is one, the unstaged one otherwise, "?" for an untracked file.
func gitStatusLabel(status string) string {
	if status[0] == '?' {
		return "?"
	}
	if status[0] != ' ' {
		return string(status[0])
	}
	return string(status[1])
}

// gitStatusColor colors the status of a change: green when added, red
// when deleted, muted when untracked, the text color otherwise.
func gitStatusColor(t *ui.Theme, status string) ui.Color {
	switch {
	case strings.Contains(status, "A"):
		return actionGreen(t)
	case strings.Contains(status, "D"):
		return diffRed(t)
	case strings.Contains(status, "?"):
		return t.TextMuted
	}
	return t.Text
}

// diffColor colors one line of an inline diff by its leading marker, as
// terminal diff viewers do.
func diffColor(t *ui.Theme, line string) ui.Color {
	switch {
	case strings.HasPrefix(line, "+"):
		return actionGreen(t)
	case strings.HasPrefix(line, "-"):
		return diffRed(t)
	case strings.HasPrefix(line, "@@"):
		return t.TextMuted
	}
	return t.Text
}

// diffRed is the red of deleted lines, muted on light, brighter on dark.
func diffRed(t *ui.Theme) ui.Color {
	if t.Dark {
		return ui.Hex("#e08a8a")
	}
	return ui.Hex("#b35959")
}

// loadFiles walks the working directory into the tree the Files section
// shows, as otty's quick directory browser. The walk is capped so a large
// tree cannot stall the panel: 600 nodes and 8 levels. Dot directories
// and .DS_Store are skipped; a directory opens in Finder, a file in Zed.
func (a *app) loadFiles() {
	if a.win == nil {
		return
	}
	a.info.files = nil
	a.info.filesDone = false
	go func() {
		s := a.currentSession()
		dir := ""
		var items []tree.Item
		if s != nil {
			dir = s.dir
			nodes := 0
			items = buildFileTree(dir, 0, &nodes)
		}
		a.win.Update(func() {
			// A directory changed mid-walk: the newer walk owns the panel,
			// and this stale one must not commit over it.
			if staleInfoCommit(a.currentSession(), s, dir) {
				return
			}
			a.info.files = items
			a.info.filesDone = true
		})
	}()
}

// fileTreeMaxNodes and fileTreeMaxDepth cap the Files walk.
const (
	fileTreeMaxNodes = 600
	fileTreeMaxDepth = 8
)

// buildFileTree builds the uix tree items of dir, depth levels deep. The
// count caps the total nodes, so a deep tree ends with an elided level
// rather than a slow walk.
func buildFileTree(dir string, depth int, nodes *int) []tree.Item {
	if depth >= fileTreeMaxDepth || *nodes >= fileTreeMaxNodes {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var items []tree.Item
	for _, e := range entries {
		name := e.Name()
		if name == ".DS_Store" || (e.IsDir() && strings.HasPrefix(name, ".")) {
			continue
		}
		if *nodes >= fileTreeMaxNodes {
			break
		}
		*nodes++
		it := tree.Item{Label: name}
		path := filepath.Join(dir, name)
		if e.IsDir() {
			open := false
			it.Open = &open
			it.Children = buildFileTree(path, depth+1, nodes)
			it.Action = func() { _ = exec.Command("open", path).Start() }
		} else {
			it.Action = func() { openZed(path) }
		}
		items = append(items, it)
	}
	return items
}

// infoFilesView is the Files section: the working directory as a tree,
// with folders that open in Finder and files that open in Zed, as otty's
// Files tab browses the project. The tree draws a folder icon on each
// directory and a file icon on each file, as otty does.
func (a *app) infoFilesView(c *ui.Context) {
	t := c.Theme()
	if !a.info.filesDone {
		ui.Text(c, tr("Reading…")).TextColor(t.TextMuted)
		return
	}
	a.infoSection(c, tr("Files"), func() {
		if len(a.info.files) == 0 {
			ui.Text(c, tr("No files")).TextColor(t.TextMuted)
			return
		}
		a.fileTree(c, a.info.files, 0)
	})
}

// fileTree draws one level of the Files section's tree. A directory shows
// a chevron and a folder icon, opens and closes when clicked, and nests
// its children one indent in; a file shows a file icon and opens on a
// click. The open state lives on the items' Open pointers, as the tree
// component's does.
func (a *app) fileTree(c *ui.Context, items []tree.Item, depth int) {
	t := c.Theme()
	for i, it := range items {
		dir := it.Open != nil
		b := ui.ButtonBase(c.Key(fmt.Sprintf("file-%d-%d", depth, i))).
			FillWidth().Padding(4, 0, 4, 4+float32(depth)*12).Radius(t.Radius).FocusRing(false)
		if b.Hovered() {
			b.Background(t.SurfaceHover)
		}
		b.Children(func() {
			// The row fills the button so the items of a level line up at
			// the same x; a shrink-to-fit row would center on the button.
			ui.Row(c).FillWidth().Gap(6).Children(func() {
				if dir {
					if *it.Open {
						ui.Image(c, lucide.ChevronDown()).Size(13, 13).TextColor(faint(t))
					} else {
						ui.Image(c, lucide.ChevronRight()).Size(13, 13).TextColor(faint(t))
					}
				} else {
					ui.Box(c).Width(13)
				}
				if dir {
					ui.Image(c, lucide.Folder()).Size(14, 14).TextColor(t.TextMuted)
				} else {
					ui.Image(c, lucide.File()).Size(14, 14).TextColor(t.TextMuted)
				}
				ui.Text(c, it.Label).SingleLine().FontSize(12).TextColor(t.Text)
			})
		})
		if b.Clicked() {
			if dir {
				*it.Open = !*it.Open
			} else if it.Action != nil {
				it.Action()
			}
		}
		if dir && *it.Open {
			a.fileTree(c, it.Children, depth+1)
		}
	}
}
