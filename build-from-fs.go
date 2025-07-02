package goopenapibuilder

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type openAPIDoc map[string]any

var rootFields = map[string]any{
	"openapi":           nil,
	"info":              nil,
	"jsonSchemaDialect": nil,
	"servers":           nil,
	"components":        nil,
	"security":          nil,
	"tags":              nil,
	"externalDocs":      nil,
}

var pathLikeFields = map[string]any{
	"paths":    nil,
	"webhooks": nil,
}

var componentFields = map[string]any{
	"schemas":         nil,
	"responses":       nil,
	"parameters":      nil,
	"examples":        nil,
	"requestBodies":   nil,
	"headers":         nil,
	"securitySchemes": nil,
	"links":           nil,
	"callbacks":       nil,
	"pathItems":       nil,
}

type fileType string

const (
	fileTypeYaml fileType = "yaml"
	fileTypeJson fileType = "json"
)

var knownExtensions = map[string]fileType{
	".yml":  fileTypeYaml,
	".yaml": fileTypeYaml,
	".json": fileTypeJson,
}

func createLoader(doc openAPIDoc, fsys fs.FS) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		// no business with directories
		if d.IsDir() {
			return nil
		}

		// parse separate parts of path
		ext := filepath.Ext(path)
		baseFilename, _ := strings.CutSuffix(filepath.Base(path), ext)
		ext = strings.ToLower(ext)
		parentDir := filepath.Base(filepath.Dir(path))
		isComponentObject := false
		isPathOrWebhook := false

		// check if file is a components field
		if _, ok := componentFields[parentDir]; ok {
			isComponentObject = true
		}

		// check if file is a path or webhook item
		if _, ok := pathLikeFields[parentDir]; ok {
			isPathOrWebhook = true
		}

		// check file extension, skip if not known
		ft, ok := knownExtensions[ext]
		if !ok {
			return nil
		}

		// read file
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return errors.New("could not read file " + path + " . error: " + err.Error())
		}

		// initialize parsed data
		partial := map[string]any{}

		// decode data into partial
		switch ft {
		case fileTypeYaml:
			if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&partial); err != nil {
				return errors.New("could not parse yaml file " + path + " . error: " + err.Error())
			}

		case fileTypeJson:
			if err := json.NewDecoder(bytes.NewReader(data)).Decode(&partial); err != nil {
				return errors.New("could not parse yaml file " + path + " . error: " + err.Error())
			}
		}

		// if it's a component, insert into components field of doc
		if isComponentObject {
			var compObj, compFieldObj map[string]any
			var ok bool
			compObj, ok = doc["components"].(map[string]any)

			if !ok {
				compObj = map[string]any{}
				doc["components"] = compObj
			}

			compFieldObj, ok = compObj[parentDir].(map[string]any)

			if !ok {
				compFieldObj = map[string]any{}
				compObj[parentDir] = compFieldObj
			}

			compFieldObj[baseFilename] = partial

			return nil
		}

		// if it's a path or webhook item, insert it into paths/webhooks field
		if isPathOrWebhook {
			var obj map[string]any
			var ok bool

			obj, ok = doc[parentDir].(map[string]any)

			if !ok {
				obj = map[string]any{}
				doc[parentDir] = obj
			}

			maps.Copy(obj, partial)

			return nil
		}

		// not a component, so set it in root document
		maps.Copy(doc, partial)

		return nil
	}
}

// BuildFromFS builds OpenAPI documentation from files
// found under fss.
func BuildFromFS(fss ...fs.FS) (io.Reader, error) {
	// no FS, no document to build
	if len(fss) == 0 {
		return nil, errors.New("no FS provided")
	}

	doc := openAPIDoc{}

	for _, fsys := range fss {
		fs.WalkDir(fsys, ".", createLoader(doc, fsys))
	}

	buf, err := json.Marshal(doc)
	if err != nil {
		return nil, errors.New("could not marshal root document. error: " + err.Error())
	}

	return bytes.NewReader(buf), nil
}
