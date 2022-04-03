package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue513OKWithExtension(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    delete:
      summary: Delete something
      responses:
        200:
          description: Success
        default:
          description: '* **400** - Bad Request'
          x-my-extension: {val: ue}
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      description: An error response body.
      properties:
        message:
          description: A detailed message describing the error.
          type: string
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	data, err := json.Marshal(doc)
	require.NoError(t, err)
	require.Contains(t, string(data), `x-my-extension`)
}

func TestIssue513KOHasExtraFieldSchema(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    delete:
      summary: Delete something
      responses:
        200:
          description: Success
        default:
          description: '* **400** - Bad Request'
          x-my-extension: {val: ue}
          # Notice here schema is invalid. It should instead be:
          # content:
          #   application/json:
          #     schema:
          #       $ref: '#/components/schemas/Error'
          schema:
            $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      description: An error response body.
      properties:
        message:
          description: A detailed message describing the error.
          type: string
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context) // FIXME unmarshal or validation error
	// TODO: merge unmarshal + validation ?
	// but still allow Validate so one can modify value then validate without marshaling
	require.Error(t, err)
}

func TestIssue513KOMixesRefAlongWithOtherFields(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    delete:
      summary: Delete something
      responses:
        200:
          description: Success
          $ref: '#/components/responseBodies/SomeResponseBody'
components:
  responseBodies:
    SomeResponseBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
  schemas:
    Error:
      type: object
      description: An error response body.
      properties:
        message:
          description: A detailed message describing the error.
          type: string
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.Error(t, err)
	require.Nil(t, doc)
}
