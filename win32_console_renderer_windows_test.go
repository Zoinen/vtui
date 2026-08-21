//go:build windows

package vtui

import "testing"

func TestWin32ConsoleRendererTracksOnlyChangedCellBounds(t *testing.T) {
	const width, height = 80, 25
	buf := make([]CharInfo, width*height)
	shadow := make([]CharInfo, len(buf))
	renderer := &Win32ConsoleRenderer{}

	// Allocating the native mirror requires one initial full publish.
	renderer.Render(buf, shadow, width, height, false)
	initial, ok := renderer.damage.take()
	if !ok || initial != (SmallRect{Left: 0, Top: 0, Right: 79, Bottom: 24}) {
		t.Fatalf("initial damage = %+v, %v; want full frame", initial, ok)
	}

	buf[7*width+19] = CharInfo{Char: 'A', Attributes: SetIndexBoth(0, 7, 0)}
	buf[8*width+23] = CharInfo{Char: 'B', Attributes: SetIndexBoth(0, 7, 0)}
	renderer.Render(buf, shadow, width, height, false)
	changed, ok := renderer.damage.take()
	want := SmallRect{Left: 19, Top: 7, Right: 23, Bottom: 8}
	if !ok || changed != want {
		t.Fatalf("changed-cell damage = %+v, %v; want %+v", changed, ok, want)
	}
}

func TestWin32ConsoleRendererSkipsUnchangedFramePublish(t *testing.T) {
	const width, height = 12, 4
	buf := make([]CharInfo, width*height)
	shadow := make([]CharInfo, len(buf))
	renderer := &Win32ConsoleRenderer{}

	renderer.Render(buf, shadow, width, height, false)
	_, _ = renderer.damage.take()
	renderer.Render(buf, shadow, width, height, false)
	if _, ok := renderer.damage.take(); ok {
		t.Fatal("unchanged frame unexpectedly scheduled a console-buffer write")
	}
}

func TestWin32ConsoleRendererPaletteChangeForcesFullPublish(t *testing.T) {
	const width, height = 12, 4
	buf := make([]CharInfo, width*height)
	shadow := make([]CharInfo, len(buf))
	renderer := &Win32ConsoleRenderer{}

	renderer.Render(buf, shadow, width, height, false)
	_, _ = renderer.damage.take()

	palette := XTerm256Palette
	renderer.SetPalette(&palette)
	renderer.Render(buf, shadow, width, height, false)
	full, ok := renderer.damage.take()
	want := SmallRect{Left: 0, Top: 0, Right: 11, Bottom: 3}
	if !ok || full != want {
		t.Fatalf("new-palette damage = %+v, %v; want %+v", full, ok, want)
	}

	// Reusing an unchanged palette must not turn a cursor-only frame into a
	// full screen-buffer update.
	renderer.SetPalette(&palette)
	renderer.Render(buf, shadow, width, height, false)
	if _, ok := renderer.damage.take(); ok {
		t.Fatal("unchanged palette unexpectedly forced a frame publish")
	}

	palette[17] ^= 0x010101
	renderer.SetPalette(&palette)
	renderer.Render(buf, shadow, width, height, false)
	full, ok = renderer.damage.take()
	if !ok || full != want {
		t.Fatalf("changed-palette damage = %+v, %v; want %+v", full, ok, want)
	}
}
