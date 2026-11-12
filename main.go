// Command frame-audit compares a sequence of still frames to flag large visual
// changes for human review. It does not identify faces or detect manipulation.
package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	dir := flag.String("dir", "", "directory containing ordered PNG or JPEG frames")
	region := flag.String("rect", "", "optional x,y,width,height crop in pixels")
	threshold := flag.Float64("threshold", 0.25, "flag a normalized mean change at or above this value (0..1)")
	output := flag.String("out", "", "CSV output path; stdout when empty")
	flag.Parse()
	if err := run(*dir, *region, *threshold, *output); err != nil {
		fmt.Fprintln(os.Stderr, "frame-audit:", err)
		os.Exit(1)
	}
}

func run(dir, region string, threshold float64, output string) error {
	if dir == "" {
		return errors.New("-dir is required")
	}
	if threshold < 0 || threshold > 1 {
		return errors.New("-threshold must be between 0 and 1")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var names []string
	for _, entry := range entries {
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !entry.IsDir() && (ext == ".png" || ext == ".jpg" || ext == ".jpeg") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	if len(names) < 2 {
		return errors.New("at least two PNG or JPEG frames are required")
	}

	var writer io.Writer = os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = file
	}
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{"previous_frame", "current_frame", "mean_abs_change", "flagged"}); err != nil {
		return err
	}
	previous, err := load(filepath.Join(dir, names[0]))
	if err != nil {
		return err
	}
	box, err := parseRect(region, previous.Bounds())
	if err != nil {
		return err
	}
	for i := 1; i < len(names); i++ {
		current, err := load(filepath.Join(dir, names[i]))
		if err != nil {
			return err
		}
		if current.Bounds() != previous.Bounds() {
			return fmt.Errorf("frame %q has different dimensions", names[i])
		}
		score := meanAbsChange(previous, current, box)
		row := []string{names[i-1], names[i], strconv.FormatFloat(score, 'f', 6, 64), strconv.FormatBool(score >= threshold)}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
		previous = current
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

func load(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode %q: %w", path, err)
	}
	return img, nil
}

func parseRect(s string, bounds image.Rectangle) (image.Rectangle, error) {
	if s == "" {
		return bounds, nil
	}
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return image.Rectangle{}, errors.New("-rect must be x,y,width,height")
	}
	var values [4]int
	for i, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return image.Rectangle{}, errors.New("-rect must contain integers")
		}
		values[i] = value
	}
	if values[2] <= 0 || values[3] <= 0 {
		return image.Rectangle{}, errors.New("-rect width and height must be positive")
	}
	box := image.Rect(values[0], values[1], values[0]+values[2], values[1]+values[3])
	if !box.In(bounds) {
		return image.Rectangle{}, errors.New("-rect must fit inside every frame")
	}
	return box, nil
}

func meanAbsChange(a, b image.Image, box image.Rectangle) float64 {
	var total uint64
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			ar, ag, ab, _ := a.At(x, y).RGBA()
			br, bg, bb, _ := b.At(x, y).RGBA()
			total += distance(ar, br) + distance(ag, bg) + distance(ab, bb)
		}
	}
	return float64(total) / (float64(box.Dx()) * float64(box.Dy()) * 3 * 65535)
}

func distance(a, b uint32) uint64 {
	if a > b {
		return uint64(a - b)
	}
	return uint64(b - a)
}
