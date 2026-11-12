package main

import (
	"encoding/csv"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestRunFlagsChangedFrame(t *testing.T) {
	dir := t.TempDir()
	for i, c := range []color.NRGBA{{A: 255}, {R: 255, G: 255, B: 255, A: 255}} {
		img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		img.SetNRGBA(0, 0, c)
		path := filepath.Join(dir, []string{"01.png", "02.png"}[i])
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(t.TempDir(), "report.csv")
	if err := run(dir, "0,0,1,1", 0.5, out); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][2] != "1.000000" || rows[1][3] != "true" {
		t.Fatalf("unexpected CSV: %v", rows)
	}
}
