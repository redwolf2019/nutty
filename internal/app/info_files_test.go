package app

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/smallnest/uix/components/tree"
)

// TestFilesTreeKeys checks that the rows of two open folders are separate
// elements: a click on a file of one folder opens that file alone. The
// Tester panics when two rows under one parent share a key.
func TestFilesTreeKeys(t *testing.T) {
	a := testApp()
	var opened []string
	folder := func(name, file string) tree.Item {
		open := true
		return tree.Item{Label: name, Open: &open, Children: []tree.Item{
			{Label: file, Action: func() { opened = append(opened, file) }},
		}}
	}
	a.info = infoState{filesDone: true, files: []tree.Item{
		folder("cmd", "main.go"),
		folder("docs", "index.html"),
	}}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() { a.infoFilesView(c) })
	}, 320, 400)
	tt.Frame()
	if err := tt.Click("index.html"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !slices.Equal(opened, []string{"index.html"}) {
		t.Fatalf("opened %q, want index.html alone", opened)
	}
}
