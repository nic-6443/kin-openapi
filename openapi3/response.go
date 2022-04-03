package openapi3

import (
	"bytes"
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
	// ExtensionProps
	Extensions map[string]json.RawMessage `json:"-" yaml:"-"`

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
	var illegals []string
	for k := range response.Extensions {
		if !strings.HasPrefix("x-", k) {
			illegals = append(illegals, k)
		}
	}
	if len(illegals) != 0 {
		sort.Strings(illegals)
		return nil, fmt.Errorf(`expected "x-" prefixes, got: %+v`, illegals)
	}
	type _Response Response
	return json.Marshal(_Response(*response))
}

// UnmarshalJSON sets Response to a copy of data.
func (response *Response) UnmarshalJSON(data []byte) error {
	// return jsoninfo.UnmarshalStrictStruct(data, response)

	// d := json.NewDecoder(bytes.NewReader(data))
	// d.DisallowUnknownFields()
	// return d.Decode(&response)

	// var x interface{}
	// if err := json.Unmarshal(data, &x); err != nil {
	// 	return err
	// }
	// y, ok := x.(map[string]interface{})
	// if !ok {
	// 	return fmt.Errorf("expected a mapping, got: %s", data)
	// }
	// if z:=y["description"];z!=nil{
	// 	var w *string
	// 	if err:=json.Unmarshal()
	// }

	// Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// Headers     Headers `json:"headers,omitempty" yaml:"headers,omitempty"`
	// Content     Content `json:"content,omitempty" yaml:"content,omitempty"`
	// Links       Links   `json:"links,omitempty" yaml:"links,omitempty"`

	// var x struct {
	// 	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// 	Headers     Headers `json:"headers,omitempty" yaml:"headers,omitempty"`
	// 	Content     Content `json:"content,omitempty" yaml:"content,omitempty"`
	// 	Links       Links   `json:"links,omitempty" yaml:"links,omitempty"`
	// 	// Extensions map[string]json.RawMessage `json:"-"`
	// 	Extensions map[string]interface{} `json:"-" yaml:"-"`
	// }
	// if err := json.Unmarshal(data, &x); err != nil {
	// 	return err
	// }
	// response.Description = x.Description
	// response.Headers = x.Headers
	// response.Content = x.Content
	// response.Links = x.Links
	// fmt.Printf(">>> %+v\n", response)
	// fmt.Printf(">>> %+v\n", x.Extensions)
	// fmt.Printf(">>> %#v\n", x.Extensions)
	// return nil

	if false { // THIS WORKS
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		type _Response Response
		var x _Response
		if err := d.Decode(&x); err != nil {
			return err
		}
		*response = Response(x)
		return nil
	}

	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	// var x struct {
	// 	Description *string                    `json:"description,omitempty" yaml:"description,omitempty"`
	// 	Headers     Headers                    `json:"headers,omitempty" yaml:"headers,omitempty"`
	// 	Content     Content                    `json:"content,omitempty" yaml:"content,omitempty"`
	// 	Links       Links                      `json:"links,omitempty" yaml:"links,omitempty"`
	// 	Extensions  map[string]json.RawMessage `json:"-" yaml:"-"`
	// }
	type _Response Response
	var x _Response
	if err := d.Decode(&x); err != nil {
		return err
	}
	var m map[string]json.RawMessage
	_ = json.Unmarshal(data, &m)
	delete(m, "description")
	delete(m, "headers")
	delete(m, "content")
	delete(m, "links")
	x.Extensions = make(map[string]json.RawMessage, len(m))
	var unknowns []string
	for k, v := range m {
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			delete(m, k)
			continue
		}
		unknowns = append(unknowns, k)
	}
	if len(unknowns) != 0 {
		sort.Strings(unknowns)
		return fmt.Errorf("unknown fields: %+v", unknowns)
	}
	*response = Response(x)
	return nil
}

// Validate returns an error if Response does not comply with the OpenAPI spec.
func (response *Response) Validate(ctx context.Context) error {
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
