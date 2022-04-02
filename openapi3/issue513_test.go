package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue513(t *testing.T) {
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
	require.Error(t, err)
	require.Nil(t, doc)
}
