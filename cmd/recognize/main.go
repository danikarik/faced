package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kagami/go-face"
	"github.com/peterbourgon/ff/v3"
)

func exit(code int, a ...interface{}) {
	if len(a) > 0 {
		fmt.Println(a...)
	}
	os.Exit(code)
}

func main() {
	fs := flag.NewFlagSet("recognize", flag.ExitOnError)

	var (
		path    = fs.String("models-path", "", "Models directory path")
		samples = fs.String("samples-path", "", "Image samples path")
		output  = fs.String("output-path", "", "Face detection output path")
		_       = fs.String("passport-image-name", "", "Verified passport image name")
		_       = fs.String("input-image-name", "", "Input image name")
		_       = fs.String("config", "", "Config file (optional)")
	)

	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("RECOGNIZE"),
		ff.WithConfigFile("config.json"),
		ff.WithConfigFileParser(ff.JSONParser),
	); err != nil {
		exit(1, err)
	}

	rec, err := face.NewRecognizer(*path)
	if err != nil {
		exit(1, err)
	}
	defer rec.Close()

	if err := filepath.Walk(*samples, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() || strings.Contains(fpath, "multiple") {
			return nil
		}

		if filepath.Ext(fpath) != ".jpg" && filepath.Ext(fpath) != ".jpeg" {
			return errors.New("image must be in jpeg extension")
		}

		face, err := rec.RecognizeSingleFile(fpath)
		if err != nil {
			return fmt.Errorf("recognizing file %q: %w", fpath, err)
		}

		if face == nil {
			return fmt.Errorf("no faces are found for %q", fpath)
		}

		original, err := os.Open(fpath)
		if err != nil {
			return fmt.Errorf("reading original file: %w", err)
		}
		defer original.Close()

		img, err := jpeg.Decode(original)
		if err != nil {
			return fmt.Errorf("decoding original image: %w", err)
		}

		rect := image.NewRGBA(face.Rectangle)
		draw.Draw(rect, rect.Bounds(), img, rect.Rect.Min, draw.Src)

		marked, err := os.Create(filepath.Join(*output, info.Name()))
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer marked.Close()

		if err := jpeg.Encode(marked, rect, &jpeg.Options{Quality: 100}); err != nil {
			return fmt.Errorf("encoding marked image: %w", err)
		}

		return nil
	}); err != nil {
		exit(1, err)
	}

	fmt.Println("OK")
}
