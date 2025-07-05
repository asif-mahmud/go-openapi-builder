package goopenapibuilder_test

import (
	"embed"
	"encoding/json"
	"io"
	"testing"

	goopenapibuilder "github.com/asif-mahmud/go-openapi-builder"
	"github.com/stretchr/testify/assert"
)

//go:embed base/*
var baseFS embed.FS

//go:embed pets/*
var petsFS embed.FS

//go:embed components/*
var componentsFS embed.FS

//go:embed pets-put/*
var petsPutFS embed.FS

//go:embed out/all.json
var allCombinedJson []byte

func TestCombined(t *testing.T) {
	doc, err := goopenapibuilder.BuildFromFS(baseFS, petsFS)

	assert.Nil(t, err)

	docData, err := io.ReadAll(doc)

	assert.Nil(t, err)

	var actual, expected map[string]any

	err = json.Unmarshal(allCombinedJson, &expected)
	assert.Nil(t, err)

	err = json.Unmarshal(docData, &actual)
	assert.Nil(t, err)

	assert.Equal(t, expected, actual)
}

func TestDuplicateSchemaFails(t *testing.T) {
	doc, err := goopenapibuilder.BuildFromFS(petsFS, componentsFS)

	assert.Nil(t, doc)
	assert.NotNil(t, err)
	assert.Equal(
		t,
		"field Category already exists in components > schemas. file: components/schemas/Category.yaml",
		err.Error(),
	)
}

func TestDuplicateOperationFails(t *testing.T) {
	doc, err := goopenapibuilder.BuildFromFS(petsFS, petsPutFS)

	assert.Nil(t, doc)
	assert.NotNil(t, err)
	assert.Equal(
		t,
		"operation put already exists in paths > /pet. file: pets-put/paths/put-pets.yaml",
		err.Error(),
	)
}
