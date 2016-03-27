//+build ignore

// gen.go generates the data files required to decode CEL images, which specify
// the decoding algorithms, image dimensions, palettes and colour transitions of
// each CEL image.
package main

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"

	"github.com/mewkiz/pkg/errutil"
	"github.com/sanctuary/formats/level/min"
)

func main() {
	var (
		// mpqDir specifies the path to an extracted "diabdat.mpq".
		mpqDir string
	)
	flag.StringVar(&mpqDir, "mpqdir", "diabdat/", `Path to extracted "diabdat.mpq".`)
	flag.Parse()

	// Parse MIN files.
	levelNames := []string{"l1", "l2", "l3", "l4", "town"}
	var mappings []*minMapping
	for _, levelName := range levelNames {
		mapping, err := parseMin(mpqDir, levelName)
		if err != nil {
			log.Fatal(err)
		}
		mappings = append(mappings, mapping)
	}

	// Generate "data.go".
	if err := genData(mappings); err != nil {
		log.Fatal(err)
	}
}

// A minMapping specifies the mapping between frame numbers and frame types of a
// given MIN file.
type minMapping struct {
	// Level name.
	LevelName string
	// frameTypes maps from frame number to frame type.
	FrameTypes []int
}

// parseMin parses the given MIN file and returns a mapping from frame numbers
// to frame types.
func parseMin(mpqDir, levelName string) (*minMapping, error) {
	// MIN path; e.g. "diabdat/levels/l1data/l1.cel".
	name := levelName + ".min"
	minPath := filepath.Join(mpqDir, "levels", levelName+"data", name)
	pieces, err := min.Parse(minPath)
	if err != nil {
		return nil, errutil.Err(err)
	}
	// m maps from frame numbers to frame types.
	m := make(map[int]int)
	for _, piece := range pieces {
		for _, block := range piece.Blocks {
			m[block.FrameNum] = block.FrameType
		}
	}
	mapping := &minMapping{
		LevelName:  levelName,
		FrameTypes: make([]int, len(m)),
	}
	for frameNum, frameType := range m {
		mapping.FrameTypes[frameNum] = frameType
	}
	return mapping, nil
}

// genData generates the data files required to decode CEL images, which specify
// the decoding algorithms, image dimensions, palettes and colour transitions of
// each CEL image.
func genData(mappings []*minMapping) error {
	t := template.New("data")
	if _, err := t.Parse(dataContent[1:]); err != nil {
		return errutil.Err(err)
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, mappings); err != nil {
		return errutil.Err(err)
	}
	data, err := format.Source(buf.Bytes())
	if err != nil {
		return errutil.Err(err)
	}
	if err := ioutil.WriteFile("data.go", data, 0644); err != nil {
		return errutil.Err(err)
	}
	return nil
}

const dataContent = `
// generated by gen.go using 'go generate'; DO NOT EDIT.

package cel

import (
	"image"
	"image/color"
)

// decoders maps CEL frame types to decoder functions.
var decoders = [...]func([]byte, int, int, color.Palette) image.Image{
	0: decodeType0,
	1: decodeType1,
	2: decodeType2,
	3: decodeType3,
	4: decodeType4,
	5: decodeType5,
	6: decodeType6,
}

// Mappings from frame numbers to frame types for each of the level CEL files
// "l1.cel", "l2.cel", "l3.cel", "l4.cel" and "town.cel".
var (
{{- range . }}
	{{ .LevelName }}FrameTypes = {{ printf "%#v" .FrameTypes }}
{{- end }}
)
`
