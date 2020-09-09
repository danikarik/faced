package main

import (
	"errors"
	"flag"
	"fmt"
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

func checkExtension(path string) error {
	if filepath.Ext(path) != ".jpg" && filepath.Ext(path) != ".jpeg" {
		return errors.New("image must be in jpeg extension")
	}
	return nil
}

func recognizeByPath(rec *face.Recognizer, path string) (*face.Face, error) {
	face, err := rec.RecognizeSingleFile(path)
	if err != nil {
		return nil, err
	}
	if face == nil {
		return nil, fmt.Errorf("no faces are found for %q", path)
	}
	return face, nil
}

func main() {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)

	var (
		path    = fs.String("models-path", "", "Models directory path")
		samples = fs.String("samples-path", "", "Image samples path")
		_       = fs.String("output-path", "", "Face detection output path")
		_       = fs.String("passport-image-name", "", "Verified passport image name")
		input   = fs.String("input-image-name", "", "Input image name")
		_       = fs.String("config", "", "Config file (optional)")
	)

	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("COMPARE"),
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

	var (
		known      []face.Descriptor
		categories []int32
		labels     []string
		index      int32
	)

	if err := filepath.Walk(*samples, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() || strings.Contains(fpath, "multiple") {
			return nil
		}

		if err := checkExtension(fpath); err != nil {
			return err
		}

		face, err := recognizeByPath(rec, fpath)
		if err != nil {
			return err
		}

		known = append(known, face.Descriptor)
		categories = append(categories, index)
		labels = append(labels, fpath)

		index++
		return nil
	}); err != nil {
		exit(1, err)
	}

	rec.SetSamples(known, categories)

	inputFace, err := recognizeByPath(rec, *input)
	if err != nil {
		exit(1, err)
	}

	if id := rec.ClassifyThreshold(inputFace.Descriptor, 0.5); id < 0 {
		exit(1, errors.New("cannot classify input image"))
	} else {
		fmt.Println("OK.", labels[id], id)
	}
}
