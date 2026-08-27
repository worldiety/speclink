package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // registers the JPEG decoder
	_ "image/png"  // registers the PNG decoder
	"os"
	"path/filepath"
	"strings"
)

// ManifestSuffix names the sidecar that declares the regions of an image.
//
// The manifest sits next to the image rather than inside the requirement tree,
// because it describes the source and belongs to whoever owns the source. It is
// itself a document under fingerprint: adding a region is a change to the
// source and creates a new coverage obligation, which is exactly what should
// happen when somebody draws a new control into a mockup.
const ManifestSuffix = ".speclink.json"

// ManifestVersion is the schema version of the sidecar. A reader that does not
// know the version refuses rather than guesses, on the same grounds as the
// baseline: a half understood manifest would silently change which parts of an
// image carry an obligation.
const ManifestVersion = 1

// Manifest is the on disk form of an image region declaration.
type Manifest struct {
	Version int      `json:"version"`
	Regions []Region `json:"regions"`
}

// Region is one declared part of an image.
//
// Rect is in pixels of the image as stored, origin top left. Coordinates rather
// than names alone are what make drift specific: the fingerprint covers the
// pixels inside the rectangle, so changing one control reports the requirements
// of that control instead of every requirement on the screen. The cost is that
// a layout shift moves everything below it and reports drift across the board —
// visible, resolved by one freeze, and preferable to a report so coarse that it
// is ignored.
type Region struct {
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Rect        Rect   `json:"rect"`
	Informative bool   `json:"informative,omitempty"`
}

// Rect is a pixel rectangle, origin top left.
type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// ManifestPath returns the sidecar path belonging to an image.
func ManifestPath(image string) string { return image + ManifestSuffix }

// segmentImage reads the sidecar and fingerprints each declared region.
func segmentImage(doc string, abs string) ([]Segment, []error) {
	manifestAbs := ManifestPath(abs)
	manifestDoc := ManifestPath(doc)

	data, err := os.ReadFile(manifestAbs)
	if err != nil {
		return nil, []error{&SegmentError{
			Doc: doc,
			Msg: "image has no region manifest at " + filepath.Base(manifestDoc),
			Why: "An image cannot be decomposed by any deterministic rule, so its addressable parts have to be declared. Without a manifest the image contributes nothing to the forward coverage and looks exactly like one that is fully covered.",
			How: "Create " + filepath.Base(manifestDoc) + ` with {"version": 1, "regions": [{"name": "...", "rect": {"x":0,"y":0,"w":0,"h":0}}]}.`,
		}}
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, []error{&SegmentError{
			Doc: manifestDoc,
			Msg: "region manifest is not valid JSON: " + err.Error(),
			Why: "The manifest decides which parts of the image carry a requirement obligation. An unreadable one cannot be treated as empty, because empty means no obligation.",
			How: "Fix the JSON syntax.",
		}}
	}
	if m.Version != ManifestVersion {
		return nil, []error{&SegmentError{
			Doc: manifestDoc,
			Msg: fmt.Sprintf("region manifest has version %d, this speclink reads version %d", m.Version, ManifestVersion),
			Why: "Reading a manifest of an unknown version would mean guessing at its meaning, and a wrong guess changes which parts of the image are required to be covered.",
			How: fmt.Sprintf("Set \"version\": %d and adjust the regions to that schema.", ManifestVersion),
		}}
	}

	img, err := decode(abs)
	if err != nil {
		return nil, []error{&SegmentError{
			Doc: doc,
			Msg: "image cannot be decoded: " + err.Error(),
			Why: "Region fingerprints are taken over decoded pixels, so the image has to be readable.",
			How: "Re-export the image as PNG or JPEG.",
		}}
	}
	bounds := img.Bounds()

	var (
		out  []Segment
		errs []error
		seen = map[string]bool{}
	)
	for _, r := range m.Regions {
		switch {
		case r.Name == "":
			errs = append(errs, &SegmentError{
				Doc: manifestDoc,
				Msg: "a region has no name",
				Why: "The name is the address a requirement cites. An unnamed region cannot be referenced and therefore can never be covered.",
				How: "Give every region a name.",
			})
			continue
		case seen[r.Name]:
			errs = append(errs, &SegmentError{
				Doc: manifestDoc,
				Msg: "region " + quote(r.Name) + " is declared twice",
				Why: "The name is the address of a region. Two regions with one address make every citation of it ambiguous.",
				How: "Rename one of them.",
			})
			continue
		}
		seen[r.Name] = true

		rect := image.Rect(r.Rect.X, r.Rect.Y, r.Rect.X+r.Rect.W, r.Rect.Y+r.Rect.H)
		if r.Rect.W <= 0 || r.Rect.H <= 0 || !rect.In(bounds) {
			errs = append(errs, &SegmentError{
				Doc: manifestDoc,
				Msg: fmt.Sprintf("region %s does not lie inside the image", quote(r.Name)),
				Why: fmt.Sprintf("The rectangle is %dx%d at (%d,%d); the image is %dx%d. A region outside the image has no pixels to fingerprint, so its drift could never be detected.", r.Rect.W, r.Rect.H, r.Rect.X, r.Rect.Y, bounds.Dx(), bounds.Dy()),
				How: "Correct the rectangle.",
			})
			continue
		}

		title := r.Title
		if title == "" {
			title = r.Name
		}
		out = append(out, Segment{
			Doc:         doc,
			ID:          r.Name,
			Kind:        KindImage,
			Title:       title,
			Fingerprint: fingerprint(pixels(img, rect)),
			Informative: r.Informative,
			Pos:         Pos{File: manifestDoc},
		})
	}
	return out, errs
}

// decode reads an image file into memory.
func decode(abs string) (image.Image, error) {
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// pixels renders a region into raw RGBA bytes.
//
// The fingerprint has to be taken over decoded pixels, never over the file.
// Re-exporting the same mockup writes different bytes — metadata, compression
// parameters, a timestamp — while the picture is unchanged. A file hash would
// therefore report drift on every export, and a rule that fires when nothing
// changed is one that gets waived by habit within a month.
//
// The hash is exact rather than perceptual. A perceptual hash tolerates small
// differences by design, and a tolerance is a threshold below which a change to
// a requirement source goes unreported. There is no defensible value for that
// threshold, so there is none.
func pixels(img image.Image, rect image.Rectangle) []byte {
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst.Pix
}

func quote(s string) string { return `"` + s + `"` }

func itoa(n int) string { return fmt.Sprintf("%d", n) }

// IsManifest reports whether a path is a region sidecar rather than a document
// in its own right.
func IsManifest(path string) bool { return strings.HasSuffix(path, ManifestSuffix) }
