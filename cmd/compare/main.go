package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
		path     = fs.String("models-path", "", "Models directory path")
		samples  = fs.String("samples-path", "", "Image samples path")
		_        = fs.String("output-path", "", "Face detection output path")
		passport = fs.String("passport-image-name", "", "Verified passport image name")
		input    = fs.String("input-image-name", "", "Input image name")
		_        = fs.String("config", "", "Config file (optional)")
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

	if err := checkExtension(*passport); err != nil {
		exit(1, err)
	}

	faces, err := rec.RecognizeFile(filepath.Join(*samples, *passport))
	if err != nil {
		exit(1, err)
	}
	if len(faces) < 1 {
		exit(1, errors.New("wrong number of faces"))
	}

	var (
		known      []face.Descriptor
		categories []int32
		labels     = []string{"Obama"}
	)

	for i, f := range faces {
		known = append(known, f.Descriptor)
		categories = append(categories, int32(i))
	}

	rec.SetSamples(known, categories)

	inputFace, err := recognizeByPath(rec, filepath.Join(*samples, *input))
	if err != nil {
		exit(1, err)
	}

	if id := rec.ClassifyThreshold(inputFace.Descriptor, 0.5); id < 0 {
		exit(1, errors.New("cannot classify input image"))
	} else {
		fmt.Println("OK.", labels[id])
	}
}
