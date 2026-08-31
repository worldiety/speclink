package main

import (
	"encoding/binary"
	"os"
	"regexp"
	"strconv"
)

// wideRatio is where a drawing stops fitting a column and starts needing the
// page turned.
//
// A figure set at the width of the text is scaled to fit it, so its height is
// the column width divided by its aspect ratio. Past about five to two that
// height is small enough that labels stop resolving in print, and the reader is
// looking at something that appears to tell them everything while telling them
// nothing.
//
// Turning the page buys roughly half as much width again — the long side of the
// sheet instead of the short one — which is the difference between a drawing
// that can be read and one that cannot.
const wideRatio = 2.5

// wide reports whether a rendered figure is broad enough to need its own turned
// page.
//
// # Why this is measured and not declared
//
// Because nothing in the model knows. How wide a drawing comes out is decided
// by the renderer from the number of nodes and the length of their labels, and
// that is neither declared anywhere nor predictable from the graph: two
// diagrams with the same node count differ by a factor of three depending on
// how long the names are.
//
// A guess would therefore be wrong in both directions — turning the page for a
// drawing that fitted, and leaving one unreadable that did not. The file is
// right there, so it is read.
//
// A file that is missing or in a format this cannot measure is not wide. That
// is the safe answer: the ordinary layout is what every other figure gets, and
// being wrong about it costs a figure that is smaller than it might have been
// rather than a page turned for nothing.
func wide(path string) bool {
	w, h, ok := imageSize(path)
	if !ok || h <= 0 {
		return false
	}
	return float64(w)/float64(h) > wideRatio
}

// svgSize matches the dimensions in an SVG header, in either of the two forms
// the renderers here produce.
var svgSize = regexp.MustCompile(`(?:style="width:(\d+)px;height:(\d+)px|width="(\d+)px" height="(\d+)px")`)

// imageSize reads the pixel dimensions of a rendered figure.
//
// SVG and PNG only, because those are what a diagram tool writes. Both are read
// from the head of the file: an SVG carries its size in the opening tag, and a
// PNG in the first chunk after the signature.
func imageSize(path string) (int, int, bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, false
	}
	defer f.Close()

	head := make([]byte, 1024)
	n, _ := f.Read(head)
	head = head[:n]

	if n >= 24 && string(head[1:4]) == "PNG" {
		return int(binary.BigEndian.Uint32(head[16:20])), int(binary.BigEndian.Uint32(head[20:24])), true
	}

	m := svgSize.FindSubmatch(head)
	if m == nil {
		return 0, 0, false
	}
	for i := 1; i+1 < len(m); i += 2 {
		if len(m[i]) == 0 {
			continue
		}
		w, err1 := strconv.Atoi(string(m[i]))
		h, err2 := strconv.Atoi(string(m[i+1]))
		if err1 == nil && err2 == nil {
			return w, h, true
		}
	}
	return 0, 0, false
}
