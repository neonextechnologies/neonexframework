package graphql

import (
	"context"
	"fmt"
	"reflect"
)

// Builder helps build GraphQL schemas fluently
type Builder struct {
	schema *Schema
}

// NewBuilder creates a new schema builder
func NewBuilder() *Builder {
	return &Builder{
		schema: NewSchema(),
	}
}

// Query defines the Query root type
func (b *Builder) Query(fields ...*Field) *Builder {
	b.schema.QueryType = &ObjectType{
		Name:   "Query",
		Fields: fields,
	}
	return b
}

// Mutation defines the Mutation root type
func (b *Builder) Mutation(fields ...*Field) *Builder {
	b.schema.MutationType = &ObjectType{
		Name:   "Mutation",
		Fields: fields,
	}
	return b
}

// Subscription defines the Subscription root type
func (b *Builder) Subscription(fields ...*Field) *Builder {
	b.schema.SubscriptionType = &ObjectType{
		Name:   "Subscription",
		Fields: fields,
	}
	return b
}

// Type adds a type to the schema
func (b *Builder) Type(name string, fields ...*Field) *Builder {
	b.schema.AddType(&ObjectType{
		Name:   name,
		Fields: fields,
	})
	return b
}

// TypeFromStruct adds a type from a Go struct
func (b *Builder) TypeFromStruct(name string, v interface{}, description ...string) *Builder {
	objType := FromStruct(name, v)
	if len(description) > 0 {
		objType.Description = description[0]
	}
	b.schema.AddType(objType)
	return b
}

// Input adds an input type
func (b *Builder) Input(name string, fields ...*InputField) *Builder {
	b.schema.AddInput(&InputType{
		Name:   name,
		Fields: fields,
	})
	return b
}

// Enum adds an enum type
func (b *Builder) Enum(name string, values ...*EnumValue) *Builder {
	b.schema.AddEnum(&EnumType{
		Name:   name,
		Values: values,
	})
	return b
}

// Interface adds an interface type
func (b *Builder) Interface(name string, fields ...*Field) *Builder {
	b.schema.AddInterface(&InterfaceType{
		Name:   name,
		Fields: fields,
	})
	return b
}

// Union adds a union type
func (b *Builder) Union(name string, types ...string) *Builder {
	b.schema.AddUnion(&UnionType{
		Name:  name,
		Types: types,
	})
	return b
}

// Directive adds a directive
func (b *Builder) Directive(name string, locations []string, args ...*Argument) *Builder {
	b.schema.Directives = append(b.schema.Directives, &Directive{
		Name:      name,
		Locations: locations,
		Args:      args,
	})
	return b
}

// Build returns the built schema
func (b *Builder) Build() *Schema {
	return b.schema
}

// Field helpers

// F creates a new field
func F(name string, fieldType FieldType, resolver ResolverFunc, options ...FieldOption) *Field {
	field := &Field{
		Name:     name,
		Type:     fieldType,
		Resolver: resolver,
		Args:     []*Argument{},
	}

	for _, opt := range options {
		opt(field)
	}

	return field
}

// FieldOption is a functional option for fields
type FieldOption func(*Field)

// WithDescription sets the field description
func WithDescription(desc string) FieldOption {
	return func(f *Field) {
		f.Description = desc
	}
}

// WithArgs sets the field arguments
func WithArgs(args ...*Argument) FieldOption {
	return func(f *Field) {
		f.Args = args
	}
}

// WithDeprecated marks the field as deprecated
func WithDeprecated(reason string) FieldOption {
	return func(f *Field) {
		f.Deprecated = true
		f.DeprecationReason = reason
	}
}

// WithElementType sets the element type for lists/non-null
func WithElementType(elementType string) FieldOption {
	return func(f *Field) {
		f.ElementType = elementType
	}
}

// List creates a list field
func List(elementType string) FieldOption {
	return func(f *Field) {
		f.Type = TypeList
		f.ElementType = elementType
	}
}

// NonNull creates a non-null field
func NonNull(elementType string) FieldOption {
	return func(f *Field) {
		f.Type = TypeNonNull
		f.ElementType = elementType
	}
}

// Argument helpers

// Arg creates a new argument
func Arg(name string, argType FieldType, options ...ArgOption) *Argument {
	arg := &Argument{
		Name: name,
		Type: argType,
	}

	for _, opt := range options {
		opt(arg)
	}

	return arg
}

// ArgOption is a functional option for arguments
type ArgOption func(*Argument)

// ArgDescription sets the argument description
func ArgDescription(desc string) ArgOption {
	return func(a *Argument) {
		a.Description = desc
	}
}

// ArgDefault sets the default value
func ArgDefault(val interface{}) ArgOption {
	return func(a *Argument) {
		a.DefaultValue = val
	}
}

// ArgRequired marks the argument as required
func ArgRequired() ArgOption {
	return func(a *Argument) {
		a.Required = true
	}
}

// ArgElementType sets the element type
func ArgElementType(elementType string) ArgOption {
	return func(a *Argument) {
		a.ElementType = elementType
	}
}

// Input field helpers

// IF creates a new input field
func IF(name string, fieldType FieldType, options ...InputFieldOption) *InputField {
	field := &InputField{
		Name: name,
		Type: fieldType,
	}

	for _, opt := range options {
		opt(field)
	}

	return field
}

// InputFieldOption is a functional option for input fields
type InputFieldOption func(*InputField)

// IFDescription sets the input field description
func IFDescription(desc string) InputFieldOption {
	return func(f *InputField) {
		f.Description = desc
	}
}

// IFDefault sets the default value
func IFDefault(val interface{}) InputFieldOption {
	return func(f *InputField) {
		f.DefaultValue = val
	}
}

// IFRequired marks the input field as required
func IFRequired() InputFieldOption {
	return func(f *InputField) {
		f.Required = true
	}
}

// IFElementType sets the element type
func IFElementType(elementType string) InputFieldOption {
	return func(f *InputField) {
		f.ElementType = elementType
	}
}

// Enum value helpers

// EV creates a new enum value
func EV(name string, options ...EnumValueOption) *EnumValue {
	ev := &EnumValue{
		Name: name,
	}

	for _, opt := range options {
		opt(ev)
	}

	return ev
}

// EnumValueOption is a functional option for enum values
type EnumValueOption func(*EnumValue)

// EVDescription sets the enum value description
func EVDescription(desc string) EnumValueOption {
	return func(ev *EnumValue) {
		ev.Description = desc
	}
}

// EVDeprecated marks the enum value as deprecated
func EVDeprecated(reason string) EnumValueOption {
	return func(ev *EnumValue) {
		ev.Deprecated = true
		ev.DeprecationReason = reason
	}
}

// Resolver helpers

// StaticResolver creates a resolver that returns a static value
func StaticResolver(value interface{}) ResolverFunc {
	return func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
		return value, nil
	}
}

// FieldResolver creates a resolver that returns a field from the parent
func FieldResolver(fieldName string) ResolverFunc {
	return func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
		if parent == nil {
			return nil, fmt.Errorf("parent is nil")
		}

		v := reflect.ValueOf(parent)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("parent is not a struct")
		}

		field := v.FieldByName(fieldName)
		if !field.IsValid() {
			return nil, fmt.Errorf("field %s not found", fieldName)
		}

		return field.Interface(), nil
	}
}

// ArgsResolver creates a resolver that returns an argument value
func ArgsResolver(argName string) ResolverFunc {
	return func(ctx context.Context, parent interface{}, args map[string]interface{}) (interface{}, error) {
		val, ok := args[argName]
		if !ok {
			return nil, fmt.Errorf("argument %s not found", argName)
		}
		return val, nil
	}
}
