package openapi3

import (
	// "bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	// "github.com/getkin/kin-openapi/jsoninfo"
	"github.com/go-openapi/jsonpointer"
)

// Responses is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#responsesObject
type Responses map[string]*ResponseRef

var _ jsonpointer.JSONPointable = (*Responses)(nil)

func NewResponses() Responses {
	r := make(Responses)
	r["default"] = &ResponseRef{Value: NewResponse().WithDescription("")}
	return r
}

func (responses Responses) Default() *ResponseRef {
	return responses["default"]
}

func (responses Responses) Get(status int) *ResponseRef {
	return responses[strconv.FormatInt(int64(status), 10)]
}

// Validate returns an error if Responses does not comply with the OpenAPI spec.
func (responses Responses) Validate(ctx context.Context) error {
	if len(responses) == 0 {
		return errors.New("the responses object MUST contain at least one response code")
	}
	for _, v := range responses {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// JSONLookup implements github.com/go-openapi/jsonpointer#JSONPointable
func (responses Responses) JSONLookup(token string) (interface{}, error) {
	ref, ok := responses[token]
	if ok == false {
		return nil, fmt.Errorf("invalid token reference: %q", token)
	}

	if ref != nil && ref.Ref != "" {
		return &Ref{Ref: ref.Ref}, nil
	}
	return ref.Value, nil
}

// Response is specified by OpenAPI/Swagger 3.0 standard.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#responseObject
type Response struct {
	Extensions map[string]json.RawMessage `json:"-" yaml:"-"` // x-... fields

	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Headers     Headers `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     Content `json:"content,omitempty" yaml:"content,omitempty"`
	Links       Links   `json:"links,omitempty" yaml:"links,omitempty"`
}

func NewResponse() *Response {
	return &Response{}
}

func (response *Response) WithDescription(value string) *Response {
	response.Description = &value
	return response
}

func (response *Response) WithContent(content Content) *Response {
	response.Content = content
	return response
}

func (response *Response) WithJSONSchema(schema *Schema) *Response {
	response.Content = NewContentWithJSONSchema(schema)
	return response
}

func (response *Response) WithJSONSchemaRef(schema *SchemaRef) *Response {
	response.Content = NewContentWithJSONSchemaRef(schema)
	return response
}

// MarshalJSON returns the JSON encoding of Response.
func (response *Response) MarshalJSON() ([]byte, error) {
	// return jsoninfo.MarshalStrictStruct(response)

	// var illegals []string
	// for k := range response.Extensions {
	// 	if !strings.HasPrefix(k, "x-") {
	// 		illegals = append(illegals, k)
	// 	}
	// }
	// if len(illegals) != 0 {
	// 	sort.Strings(illegals)
	// 	return nil, fmt.Errorf(`expected "x-" prefixes, got: %+v`, illegals) // move to Validate()
	// }
	// type _Response Response
	// return json.Marshal(_Response(*response))

	m := make(map[string]interface{}, 4+len(response.Extensions))
	if x := response.Description; x != nil {
		m["description"] = response.Description
	}
	if x := response.Headers; len(x) != 0 {
		m["headers"] = x
	}
	if x := response.Content; len(x) != 0 {
		m["content"] = x
	}
	if x := response.Links; len(x) != 0 {
		m["links"] = x
	}
	for k, v := range response.Extensions {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets Response to a copy of data.
func (response *Response) UnmarshalJSON(data []byte) error {
	// return jsoninfo.UnmarshalStrictStruct(data, response)

	type ResponseBis Response
	var x ResponseBis
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "description")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "content")
	delete(x.Extensions, "links")
	*response = Response(x)
	return nil
}

// type ExtensionProps map[string]json.RawMessage ?
func validateExtensions(extensions map[string]json.RawMessage) error {
	var unknowns []string
	for k := range extensions {
		if !strings.HasPrefix(k, "x-") {
			unknowns = append(unknowns, k)
		}
	}
	if len(unknowns) != 0 {
		sort.Strings(unknowns)
		return fmt.Errorf("unknown fields: %+v", unknowns)
	}
	return nil
}

// Validate returns an error if Response does not comply with the OpenAPI spec.
func (response *Response) Validate(ctx context.Context) error {
	if err := validateExtensions(response.Extensions); err != nil {
		return err
	}

	if response.Description == nil {
		return errors.New("a short description of the response is required")
	}

	if content := response.Content; content != nil {
		if err := content.Validate(ctx); err != nil {
			return err
		}
	}
	for _, header := range response.Headers {
		if err := header.Validate(ctx); err != nil {
			return err
		}
	}

	for _, link := range response.Links {
		if err := link.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}
