package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue495WithCustom(t *testing.T) {
	spec := []byte(`
openapi: 3.0.1
servers:
- url: http://localhost:5000
info:
  version: v1
  title: Products api
  contact:
    name: me
    email: me@github.com
  description: This is a sample
paths:
  /categories:
    get:
      summary: Provides the available categories for the store
      operationId: list-categories
      responses:
        '200':
          description: this is a desc
          content:
            application/json:
              schema:
                $ref: http://schemas.sentex.io/store/categories.json
`[1:])

	// When that site fails to respond:

	// http://schemas.sentex.io/store/categories.json
	// {
	//   "$id": "http://schemas.sentex.io/store/categories.json",
	//   "$schema": "http://json-schema.org/draft-07/schema#",
	//   "description": "array of category strings",
	//   "type": "array",
	//   "items": {
	//     "allOf": [
	//       {
	//         "$ref": "http://schemas.sentex.io/store/category.json"
	//       }
	//     ]
	//   }
	// }

	// http://schemas.sentex.io/store/category.json
	// {
	//   "$id": "http://schemas.sentex.io/store/category.json",
	//   "$schema": "http://json-schema.org/draft-07/schema#",
	//   "description": "category name for products",
	//   "type": "string",
	//   "pattern": "^[A-Za-z0-9\\-]+$",
	//   "minimum": 1,
	//   "maximum": 30
	// }

	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}

func TestIssue495WithDraft04(t *testing.T) {
	spec := []byte(`
openapi: 3.0.1
servers:
- url: http://localhost:5000
info:
  version: v1
  title: Products api
  contact:
    name: me
    email: me@github.com
  description: This is a sample
paths:
  /categories:
    get:
      summary: Provides the available categories for the store
      operationId: list-categories
      responses:
        '200':
          description: this is a desc
          content:
            application/json:
              schema:
                $ref: http://json-schema.org/draft-04/schema
`[1:])

	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromData(spec)
	t.Skip("TODO: fix dereferencing of schemaArray.items.$ref:'#'")
	// In:
	// definitions:
	//   schemaArray:
	//     type: array
	//     minItems: 1
	//     items:
	//       "$ref": "#"
	// properties:
	//   allOf:
	//     "$ref": "#/definitions/schemaArray"
	// ...it seems schemaArray's '#' reference doesn't get expanded at the right leve:
	// => bad data in "#"
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}

func TestIssue495ObjectInsteadOfObjects(t *testing.T) {
	spec := []byte(`
openapi: 3.0.1
info:
  version: v1
  title: Products api
components:
  schemas:
    schemaArray:
      type: array
      minItems: 1
      items:
        $ref: '#'
paths:
  /categories:
    get:
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                allOf:
                  $ref: '#/components/schemas/schemaArray'
`[1:])

	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromData(spec)
	require.Contains(t, err.Error(), `failed to unmarshal property "allOf" (*openapi3.SchemaRefs): json: cannot unmarshal object into Go value of type openapi3.SchemaRefs`)
	require.Nil(t, doc)
}

func TestIssue495Bis(t *testing.T) {
	spec := []byte(`
openapi: 3.0.1
info:
  version: v1
  title: Products api
components:
  schemas:
    someSchema:
      type: object
    schemaArray:
      type: array
      minItems: 1
      items:
        $ref: '#/components/schemas/someSchema'
paths:
  /categories:
    get:
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                properties:
                  allOf:
                    $ref: '#/components/schemas/schemaArray'
`[1:])

	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
