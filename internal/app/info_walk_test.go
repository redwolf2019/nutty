package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/tree"
)

// treeLabels is the labels of a level of the tree, in order.
func treeLabels(items []tree.Item) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.Label)
	}
	return out
}

// treeChild is the item of a level with the label, or nil.
func treeChild(items []tree.Item, label string) *tree.Item {
	for i := range items {
		if items[i].Label == label {
			return &items[i]
		}
	}
	return nil
}

// projectDir makes a working directory as the Files section meets them: a
// folder with more files than a folder lists, files and a folder beside
// it, a chain of folders deeper than any walk went, and what the section
// hides.
func projectDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(path string) {
		t.Helper()
		path = filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < fileTreeMaxNodes+100; i++ {
		write(fmt.Sprintf("node_modules/f%04d", i))
	}
	for _, path := range []string{
		"README.md", "package.json", ".env", ".DS_Store", ".git/config",
		"src/main.go", "src/d1/d2/d3/d4/d5/d6/d7/d8/d9/deep.txt",
	} {
		write(path)
	}
	return dir
}

// TestListFolder checks what one listing holds: the folder's own entries
// and no deeper ones, the folders closed and not yet listed, the files
// with their click, and a count of what a long folder leaves out.
func TestListFolder(t *testing.T) {
	dir := projectDir(t)
	items, more := listFolder(dir)
	if got, want := fmt.Sprint(treeLabels(items)), "[.env README.md node_modules package.json src]"; got != want || more != 0 {
		t.Fatalf("the directory lists %s and %d more, want %s and none", got, more, want)
	}
	src := treeChild(items, "src")
	if src.Open == nil || *src.Open || src.Children != nil {
		t.Fatalf("a folder should start closed and unlisted: %+v", src)
	}
	if f := treeChild(items, "README.md"); f.Open != nil || f.Action == nil {
		t.Fatal("a file should have no open state and should open on a click")
	}
	items, more = listFolder(filepath.Join(dir, "node_modules"))
	if len(items) != fileTreeMaxNodes || more != 100 || items[0].Label != "f0000" {
		t.Fatalf("the long folder lists %d entries from %s and %d more, want %d from f0000 and 100",
			len(items), items[0].Label, more, fileTreeMaxNodes)
	}
	// A folder that is not there lists nothing, and is listed all the same.
	if items, more := listFolder(filepath.Join(dir, "gone")); items == nil || len(items) != 0 || more != 0 {
		t.Fatalf("a missing folder lists %v and %d more", items, more)
	}
}

// TestFilesOpenFolder checks that a folder of the Files section lists its
// entries when it is opened, whatever is beside or above it: the folder
// after one that is too long to list in full, and a folder deeper than
// the eight levels the section used to walk ahead, where the walk's node
// budget and depth cap left them empty.
func TestFilesOpenFolder(t *testing.T) {
	dir := projectDir(t)
	a := testApp()
	files, _ := listFolder(dir)
	a.info = infoState{filesDone: true, files: files, filesDir: dir}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() { a.infoFilesView(c) })
	}, 320, 800)
	tt.Frame()
	if tt.HasText("main.go") {
		t.Fatal("a closed folder's entries are drawn")
	}
	open := func(folder, want string) {
		t.Helper()
		if err := tt.Click(folder); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if !tt.HasText(want) {
			t.Fatalf("%s is open and does not list %s, texts: %v", folder, want, tt.Texts())
		}
	}
	open("src", "main.go")
	for i := 1; i <= 8; i++ {
		open(fmt.Sprintf("d%d", i), fmt.Sprintf("d%d", i+1))
	}
	open("d9", "deep.txt")

	// Closing and opening again keeps what was open inside.
	open("src", "README.md")
	if tt.HasText("main.go") {
		t.Fatal("a closed folder's entries are drawn")
	}
	open("src", "deep.txt")

	// The long folder lists its first entries and counts the rest.
	open("node_modules", "… +100")
	if !tt.HasText(fmt.Sprintf("f%04d", fileTreeMaxNodes-1)) || tt.HasText(fmt.Sprintf("f%04d", fileTreeMaxNodes)) {
		t.Fatalf("node_modules should list its first %d entries", fileTreeMaxNodes)
	}
}
