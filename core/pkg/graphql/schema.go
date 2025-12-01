package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// FieldType represents a GraphQL field type
type FieldType string

const (
	TypeString   FieldType = "String"
	TypeInt      FieldType = "Int"
	TypeFloat    FieldType = "Float"
	TypeBoolean  FieldType = "Boolean"
	TypeID       FieldType = "ID"
	TypeList     FieldType = "List"
	TypeNonNull  FieldType = "NonNull"
	TypeObject   FieldType = "Object"
	TypeEnum     FieldType = "Enum"
	TypeInterface FieldType = "Interface"
	TypeUnion    FieldType = "Union"
)

// Field represents a GraphQL field
type Field struct {
	Name        string
	Type        FieldType
	ElementType string // For lists and non-null
	Description string
	Args        []*Argument
	Resolver    ResolverFunc
	Deprecated  bool
	DeprecationReason string
}

// Argument represents a field argument
type Argument struct {
	Name         string
	Type         FieldType
	ElementType  string
	Description  string
	DefaultValue interface{}
	Required     bool
}

// ObjectType represents a GraphQL object type
type ObjectType struct {
	Name        string
	Description string
	Fields      []*Field
	Interfaces  []string
}

// InputType represents a GraphQL input type
type InputType struct {
	Name        string
	Description string
	Fields      []*InputField
}

// InputField represents an input field
type InputField struct {
	Name         string
	Type         FieldType
	ElementType  string
	Description  string
	DefaultValue interface{}
	Required     bool
}

// EnumType represents a GraphQL enum type
type EnumType struct {
	Name        string
	Description string
	Values      []*EnumValue
}

// EnumValue represents an enum value
type EnumValue struct {
	Name              string
	Description       string
	Deprecated        bool
	DeprecationReason string
}

// InterfaceType represents a GraphQL interface type
type InterfaceType struct {
	Name        string
	Description string
	Fields      []*Field
}

// UnionType represents a GraphQL union type
type UnionType struct {
	Name        string
	Description string
	Types       []string
}

// ResolverFunc is the function signature for field resolvers
type ResolverFunc func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error)

// Schema represents a GraphQL schema
type Schema struct {
	QueryType        *ObjectType
	MutationType     *ObjectType
	SubscriptionType *ObjectType
	Types            map[string]*ObjectType
	Inputs           map[string]*InputType
	Enums            map[string]*EnumType
	Interfaces       map[string]*InterfaceType
	Unions           map[string]*UnionType
	Directives       []*Directive
}

// Directive represents a GraphQL directive
type Directive struct {
	Name        string
	Description string
	Locations   []string
	Args        []*Argument
}

// NewSchema creates a new GraphQL schema
func NewSchema() *Schema {
	return &Schema{
		Types:      make(map[string]*ObjectType),
		Inputs:     make(map[string]*InputType),
		Enums:      make(map[string]*EnumType),
		Interfaces: make(map[string]*InterfaceType),
		Unions:     make(map[string]*UnionType),
		Directives: []*Directive{},
	}
}

// AddType adds an object type to the schema
func (s *Schema) AddType(t *ObjectType) {
	s.Types[t.Name] = t
}

// AddInput adds an input type to the schema
func (s *Schema) AddInput(i *InputType) {
	s.Inputs[i.Name] = i
}

// AddEnum adds an enum type to the schema
func (s *Schema) AddEnum(e *EnumType) {
	s.Enums[e.Name] = e
}

// AddInterface adds an interface type to the schema
func (s *Schema) AddInterface(i *InterfaceType) {
	s.Interfaces[i.Name] = i
}

// AddUnion adds a union type to the schema
func (s *Schema) AddUnion(u *UnionType) {
	s.Unions[u.Name] = u
}

// SetQuery sets the query root type
func (s *Schema) SetQuery(q *ObjectType) {
	s.QueryType = q
}

// SetMutation sets the mutation root type
func (s *Schema) SetMutation(m *ObjectType) {
	s.MutationType = m
}

// SetSubscription sets the subscription root type
func (s *Schema) SetSubscription(sub *ObjectType) {
	s.SubscriptionType = sub
}

// String generates the GraphQL SDL (Schema Definition Language)
func (s *Schema) String() string {
	var sb strings.Builder

	// Write schema definition
	sb.WriteString("schema {\n")
	if s.QueryType != nil {
		sb.WriteString(fmt.Sprintf("  query: %s\n", s.QueryType.Name))
	}
	if s.MutationType != nil {
		sb.WriteString(fmt.Sprintf("  mutation: %s\n", s.MutationType.Name))
	}
	if s.SubscriptionType != nil {
		sb.WriteString(fmt.Sprintf("  subscription: %s\n", s.SubscriptionType.Name))
	}
	sb.WriteString("}\n\n")

	// Write directives
	for _, dir := range s.Directives {
		sb.WriteString(s.directiveToSDL(dir))
		sb.WriteString("\n\n")
	}

	// Write enums
	for _, enum := range s.Enums {
		sb.WriteString(s.enumToSDL(enum))
		sb.WriteString("\n\n")
	}

	// Write interfaces
	for _, iface := range s.Interfaces {
		sb.WriteString(s.interfaceToSDL(iface))
		sb.WriteString("\n\n")
	}

	// Write input types
	for _, input := range s.Inputs {
		sb.WriteString(s.inputToSDL(input))
		sb.WriteString("\n\n")
	}

	// Write types
	for _, t := range s.Types {
		sb.WriteString(s.typeToSDL(t))
		sb.WriteString("\n\n")
	}

	// Write unions
	for _, union := range s.Unions {
		sb.WriteString(s.unionToSDL(union))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// typeToSDL converts an object type to SDL
func (s *Schema) typeToSDL(t *ObjectType) string {
	var sb strings.Builder

	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", t.Description))
	}

	sb.WriteString(fmt.Sprintf("type %s", t.Name))

	if len(t.Interfaces) > 0 {
		sb.WriteString(" implements ")
		sb.WriteString(strings.Join(t.Interfaces, " & "))
	}

	sb.WriteString(" {\n")

	for _, field := range t.Fields {
		sb.WriteString(s.fieldToSDL(field, "  "))
	}

	sb.WriteString("}")

	return sb.String()
}

// inputToSDL converts an input type to SDL
func (s *Schema) inputToSDL(i *InputType) string {
	var sb strings.Builder

	if i.Description != "" {
		sb.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", i.Description))
	}

	sb.WriteString(fmt.Sprintf("input %s {\n", i.Name))

	for _, field := range i.Fields {
		if field.Description != "" {
			sb.WriteString(fmt.Sprintf("  \"\"\"%s\"\"\"\n", field.Description))
		}

		fieldType := s.getFieldTypeString(field.Type, field.ElementType, field.Required)
		sb.WriteString(fmt.Sprintf("  %s: %s", field.Name, fieldType))

		if field.DefaultValue != nil {
			defaultJSON, _ := json.Marshal(field.DefaultValue)
			sb.WriteString(fmt.Sprintf(" = %s", string(defaultJSON)))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}")

	return sb.String()
}

// enumToSDL converts an enum type to SDL
func (s *Schema) enumToSDL(e *EnumType) string {
	var sb strings.Builder

	if e.Description != "" {
		sb.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", e.Description))
	}

	sb.WriteString(fmt.Sprintf("enum %s {\n", e.Name))

	for _, value := range e.Values {
		if value.Description != "" {
			sb.WriteString(fmt.Sprintf("  \"\"\"%s\"\"\"\n", value.Description))
		}

		sb.WriteString(fmt.Sprintf("  %s", value.Name))

		if value.Deprecated {
			sb.WriteString(fmt.Sprintf(" @deprecated(reason: \"%s\")", value.DeprecationReason))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}")

	return sb.String()
}

// interfaceToSDL converts an interface type to SDL
func (s *Schema) interfaceToSDL(i *InterfaceType) string {
	var sb strings.Builder

	if i.Description != "" {
		sb.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", i.Description))
	}

	sb.WriteString(fmt.Sprintf("interface %s {\n", i.Name))

	for _, field := range i.Fields {
		sb.WriteString(s.fieldToSDL(field, "  "))
	}

	sb.WriteString("}")

	return sb.String()
}

// unionToSDL converts a union type to SDL
func (s *Schema) unionToSDL(u *UnionType) string {
	var sb strings.Builder

	if u.Description != "" {
		sb.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", u.Description))
	}

	sb.WriteString(fmt.Sprintf("union %s = %s", u.Name, strings.Join(u.Types, " | ")))

	return sb.String()
}

// directiveToSDL converts a directive to SDL
func (s *Schema) directiveToSDL(d *Directive) string {
	var sb strings.Builder

	if d.Description != "" {
		sb.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", d.Description))
	}

	sb.WriteString(fmt.Sprintf("directive @%s", d.Name))

	if len(d.Args) > 0 {
		sb.WriteString("(")
		for i, arg := range d.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			argType := s.getFieldTypeString(arg.Type, arg.ElementType, arg.Required)
			sb.WriteString(fmt.Sprintf("%s: %s", arg.Name, argType))
		}
		sb.WriteString(")")
	}

	sb.WriteString(" on ")
	sb.WriteString(strings.Join(d.Locations, " | "))

	return sb.String()
}

// fieldToSDL converts a field to SDL
func (s *Schema) fieldToSDL(f *Field, indent string) string {
	var sb strings.Builder

	if f.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\"\"\"%s\"\"\"\n", indent, f.Description))
	}

	sb.WriteString(fmt.Sprintf("%s%s", indent, f.Name))

	if len(f.Args) > 0 {
		sb.WriteString("(")
		for i, arg := range f.Args {
			if i > 0 {
				sb.WriteString(", ")
			}
			argType := s.getFieldTypeString(arg.Type, arg.ElementType, arg.Required)
			sb.WriteString(fmt.Sprintf("%s: %s", arg.Name, argType))

			if arg.DefaultValue != nil {
				defaultJSON, _ := json.Marshal(arg.DefaultValue)
				sb.WriteString(fmt.Sprintf(" = %s", string(defaultJSON)))
			}
		}
		sb.WriteString(")")
	}

	fieldType := s.getFieldTypeString(f.Type, f.ElementType, false)
	sb.WriteString(fmt.Sprintf(": %s", fieldType))

	if f.Deprecated {
		sb.WriteString(fmt.Sprintf(" @deprecated(reason: \"%s\")", f.DeprecationReason))
	}

	sb.WriteString("\n")

	return sb.String()
}

// getFieldTypeString returns the GraphQL type string
func (s *Schema) getFieldTypeString(fieldType FieldType, elementType string, required bool) string {
	var typeStr string

	switch fieldType {
	case TypeList:
		typeStr = fmt.Sprintf("[%s]", elementType)
	case TypeNonNull:
		typeStr = fmt.Sprintf("%s!", elementType)
	default:
		typeStr = string(fieldType)
		if elementType != "" {
			typeStr = elementType
		}
	}

	if required && fieldType != TypeNonNull {
		typeStr = fmt.Sprintf("%s!", typeStr)
	}

	return typeStr
}

// FromStruct generates a GraphQL type from a Go struct
func FromStruct(name string, v interface{}) *ObjectType {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	objType := &ObjectType{
		Name:   name,
		Fields: []*Field{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from json tag or field name
		fieldName := field.Tag.Get("json")
		if fieldName == "" || fieldName == "-" {
			fieldName = field.Name
		} else {
			// Remove omitempty
			fieldName = strings.Split(fieldName, ",")[0]
		}

		// Get GraphQL type
		gqlType := goTypeToGraphQLType(field.Type)

		// Get description from graphql tag
		description := field.Tag.Get("graphql")

		objType.Fields = append(objType.Fields, &Field{
			Name:        fieldName,
			Type:        gqlType,
			Description: description,
		})
	}

	return objType
}

// goTypeToGraphQLType converts Go type to GraphQL type
func goTypeToGraphQLType(t reflect.Type) FieldType {
	switch t.Kind() {
	case reflect.String:
		return TypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return TypeInt
	case reflect.Float32, reflect.Float64:
		return TypeFloat
	case reflect.Bool:
		return TypeBoolean
	case reflect.Slice, reflect.Array:
		return TypeList
	case reflect.Struct, reflect.Ptr:
		return TypeObject
	default:
		return TypeString
	}
}
