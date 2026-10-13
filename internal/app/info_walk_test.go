package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/smallnest/uix/components/tree"
)

// treeLabels is the labels of a level of the tree, in order.
func labels(items []tree.Item) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.Label)
	}
	return out
}

// treeChild is the item of a level with the label, or nil.
func child(items []tree.Item, label string) *tree.Item {
	for i := range items {
		if items[i].Label == label {
			return &items[i]
		}
	}
	return nil
}

// TestFileTreeBudgetKeepsTopLevel checks that one big folder does not use
// up the walk's node budget before the entries beside it are listed: the
// budget goes level by level, so the working directory's own entries all
// show, and the big folder is the one cut short.
func TestFileTreeBudgetKeepsTopLevel(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(big, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < fileTreeMaxNodes+100; i++ {
		if err := os.WriteFile(filepath.Join(big, fmt.Sprintf("f%04d", i)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"README.md", "package.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nodes := 0
	items := buildFileTree(dir, 0, &nodes)
	want := []string{"README.md", "node_modules", "package.json", "src"}
	if got := labels(items); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("top level = %v, want %v", got, want)
	}
	if nodes != fileTreeMaxNodes {
		t.Fatalf("nodes = %d, want the whole budget of %d", nodes, fileTreeMaxNodes)
	}
	// What is left after the top level goes to the big folder.
	if got := len(child(items, "node_modules").Children); got != fileTreeMaxNodes-len(want) {
		t.Fatalf("node_modules lists %d entries, want %d", got, fileTreeMaxNodes-len(want))
	}
}

// TestFileTreeShape checks that the level-by-level walk still hangs every
// entry under its own folder, skips what the Files section hides, starts
// the folders closed, and stops at the depth cap.
func TestFileTreeShape(t *testing.T) {
	dir := t.TempDir()
	// A chain of folders two levels deeper than the cap, a file in each.
	deep := dir
	for i := 0; i < fileTreeMaxDepth+2; i++ {
		deep = filepath.Join(deep, fmt.Sprintf("d%d", i))
		if err := os.MkdirAll(deep, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(deep, fmt.Sprintf("f%d.txt", i)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{"lib", ".git"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, d, "in-"+d), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{".DS_Store", ".env", "main.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	nodes := 0
	items := buildFileTree(dir, 0, &nodes)
	if got, want := fmt.Sprint(labels(items)), "[.env d0 lib main.go]"; got != want {
		t.Fatalf("top level = %s, want %s", got, want)
	}
	lib := child(items, "lib")
	if lib.Open == nil || *lib.Open {
		t.Fatal("a folder should start closed")
	}
	if got, want := fmt.Sprint(labels(lib.Children)), "[in-lib]"; got != want {
		t.Fatalf("lib lists %s, want %s", got, want)
	}
	if f := child(items, "main.go"); f.Open != nil || f.Action == nil {
		t.Fatal("a file should have no open state and should open on a click")
	}
	// Level n of the chain holds dn+1 and fn.txt, down to the cap.
	level, depth := child(items, "d0"), 1
	for ; depth < fileTreeMaxDepth; depth++ {
		want := fmt.Sprintf("[d%d f%d.txt]", depth, depth-1)
		if got := fmt.Sprint(labels(level.Children)); got != want {
			t.Fatalf("level %d lists %s, want %s", depth, got, want)
		}
		level = child(level.Children, fmt.Sprintf("d%d", depth))
	}
	if len(level.Children) != 0 {
		t.Fatalf("the walk went below %d levels: %v", fileTreeMaxDepth, labels(level.Children))
	}
}
