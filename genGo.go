// Copyright 2020 - 2021 The xgen Authors. All rights reserved. Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.
//
// Package xgen written in pure Go providing a set of functions that allow you
// to parse XSD (XML schema files). This library needs Go version 1.10 or
// later.

package xgen

import (
	"fmt"
	"go/format"
	"io"
	"os"
	"reflect"
	"strings"
)

const (
	interfaceType = " interface{}\n"
	omitEmpty     = `,omitempty`
	timePackage   = "time.Time"
)

// CodeGenerator holds code generator overrides and runtime data that are used
// when generate code from proto tree.
type CodeGenerator struct {
	Lang              string
	File              string
	Field             string
	Package           string
	ImportTime        bool // For Go language
	ImportEncodingXML bool // For Go language
	ProtoTree         []interface{}
	StructAST         map[string]string
}

var goBuiltinType = map[string]struct{}{
	"[]bool":        {},
	"[]byte":        {},
	"[]interface{}": {},
	"[]string":      {},
	"bool":          {},
	"byte":          {},
	"complex128":    {},
	"complex64":     {},
	"float32":       {},
	"float64":       {},
	"int":           {},
	"int16":         {},
	"int32":         {},
	"int64":         {},
	"int8":          {},
	"interface":     {},
	"string":        {},
	"time.Time":     {},
	"uint":          {},
	"uint16":        {},
	"uint32":        {},
	"uint64":        {},
	"uint8":         {},
	"xml.Name":      {},
}

// GenGo generate Go programming language source code for XML schema
// definition files.
func (gen *CodeGenerator) GenGo() error {
	fieldNameCount = make(map[string]int)
	for _, ele := range gen.ProtoTree {
		if ele == nil {
			continue
		}

		funcName := fmt.Sprintf("Go%s", reflect.TypeOf(ele).String()[6:])

		if funcName != "GoSimpleType" {
			if err := callFuncByName(gen, funcName, []reflect.Value{reflect.ValueOf(ele)}); err != nil {
				panic(err)
			}
		}
	}

	f, err := os.Create(gen.File + ".go")
	if err != nil {
		return err
	}

	defer func(c io.Closer) { _ = c.Close() }(f)

	var importPackage, packages string
	if gen.ImportTime {
		packages += "\t\"time\"\n"
	}

	if gen.ImportEncodingXML {
		packages += "\t\"encoding/xml\"\n"
	}

	if packages != "" {
		importPackage = fmt.Sprintf("import (\n%s)", packages)
	}

	packageName := gen.Package
	if packageName == "" {
		packageName = "schema"
	}

	source, err := format.Source(
		[]byte(fmt.Sprintf(
			"%s\n\npackage %s\n%s%s",
			copyright,
			packageName,
			importPackage,
			gen.Field,
		)),
	)
	if err != nil {
		return err
	}

	_, err = f.Write(source)

	return err
}

func genGoFieldName(name string, unique bool) (fieldName string) {
	for _, str := range strings.Split(name, ":") {
		fieldName += MakeFirstUpperCase(str)
	}

	var tmp string
	for _, str := range strings.Split(fieldName, ".") {
		tmp += MakeFirstUpperCase(str)
	}

	fieldName = strings.NewReplacer("-", "", "_", "").Replace(tmp)

	if unique {
		fieldNameCount[fieldName]++

		if count := fieldNameCount[fieldName]; count != 1 {
			fieldName = fmt.Sprintf("%s%d", fieldName, count)
		}
	}

	return
}

func genGoFieldType(name string) string {
	if _, ok := goBuiltinType[name]; ok {
		return name
	}

	var fieldType string

	for _, str := range strings.Split(name, ".") {
		fieldType += MakeFirstUpperCase(str)
	}

	fieldType = strings.ReplaceAll(MakeFirstUpperCase(strings.ReplaceAll(fieldType, "-", "")), "_", "")
	if fieldType != "" {
		return fieldType
	}

	return "interface{}"
}

// GoSimpleType generates code for simple type XML schema in Go language
// syntax.
func (gen *CodeGenerator) GoSimpleType(v *SimpleType) {
	if v.List {
		if _, ok := gen.StructAST[v.Name]; !ok {
			fieldType := genGoFieldType(getBaseFromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree))
			if fieldType == timePackage {
				gen.ImportTime = true
			}
			content := fmt.Sprintf(" []%s\n", genGoFieldType(fieldType))
			gen.StructAST[v.Name] = content
			fieldName := genGoFieldName(v.Name, true)
			gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])
			return
		}
	}

	if v.Union && len(v.MemberTypes) > 0 {
		if _, ok := gen.StructAST[v.Name]; ok {
			return
		}

		var fields []string
		fieldName := genGoFieldName(v.Name, true)
		if fieldName != v.Name {
			gen.ImportEncodingXML = true
			fields = append(fields, fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`", v.Name))
		}

		for _, member := range toSortedPairs(v.MemberTypes) {
			memberName := member.key
			memberType := member.value

			if memberType == "" { // fix order issue
				memberType = getBaseFromSimpleType(memberName, gen.ProtoTree)
			}
			fields = append(fields, fmt.Sprintf("\t%s\t%s", genGoFieldName(memberName, false), genGoFieldType(memberType)))
		}

		if len(fields) > 0 {
			gen.StructAST[v.Name] = " struct {\n" + strings.Join(fields, "\n") + "\n}\n"
		} else {
			gen.StructAST[v.Name] = interfaceType
		}

		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])

		return
	}

	if _, ok := gen.StructAST[v.Name]; !ok {
		gen.StructAST[v.Name] = " " + genGoFieldType(getBaseFromSimpleType(trimNSPrefix(v.Base), gen.ProtoTree)) + "\n"
		fieldName := genGoFieldName(v.Name, true)
		gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])
	}
}

// GoComplexType generates code for complex type XML schema in Go language
// syntax.
func (gen *CodeGenerator) GoComplexType(v *ComplexType) {
	if gen.seen(v.Name) {
		return
	}

	var fields []string
	fieldName := genGoFieldName(v.Name, true)

	if fieldName != v.Name {
		gen.ImportEncodingXML = true
		fields = append(fields, fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`", v.Name))
	}

	for _, attrGroup := range v.AttributeGroup {
		fieldType := getBaseFromSimpleType(trimNSPrefix(attrGroup.Ref), gen.ProtoTree)
		if fieldType == timePackage {
			gen.ImportTime = true
		}
		fields = append(fields, fmt.Sprintf("\t%s\t%s", genGoFieldName(attrGroup.Name, false), genGoFieldType(fieldType)))
	}

	for _, attribute := range v.Attributes {
		var optional string
		if attribute.Optional {
			optional = omitEmpty
		}
		fieldType := genGoFieldType(getBaseFromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree))
		if fieldType == timePackage {
			gen.ImportTime = true
		}
		fields = append(fields, fmt.Sprintf(
			"\t%s\t%s\t`xml:\"%s,attr%s\"`",
			genGoFieldName(attribute.Name, false),
			fieldType,
			attribute.Name,
			optional,
		))
	}

	for _, group := range v.Groups {
		fields = append(fields, fmt.Sprintf(
			"\t%s\t%s%s",
			genGoFieldName(group.Name, false),
			plural(group.Plural),
			genGoFieldType(getBaseFromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree))),
		)
	}

	for _, element := range v.Elements {
		var (
			typePrefix string
			tagSuffix  string
		)

		if element.Nillable {
			typePrefix = "*"
			tagSuffix = ",omitempty"
		}

		fieldType := genGoFieldType(getBaseFromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree))
		if fieldType == timePackage {
			gen.ImportTime = true
		}

		fields = append(fields, fmt.Sprintf(
			"\t%s\t%s%s%s\t`xml:\"%s%s\"`",
			genGoFieldName(element.Name, false),
			plural(element.Plural),
			typePrefix,
			fieldType,
			element.Name,
			tagSuffix,
		))
	}

	if len(v.Base) > 0 {
		// If the type is a built-in type, generate a Value field as chardata.
		// If it's not built-in one, embed the base type in the struct for the child type
		// to effectively inherit all of the base type's fields
		if isGoBuiltInType(v.Base) {
			fields = append(fields, fmt.Sprintf("\tValue\t%s\t`xml:\",chardata\"`", genGoFieldType(v.Base)))
		} else {
			fields = append(fields, fmt.Sprintf("\t%s", "*"+genGoFieldType(v.Base)))
		}
	}

	if len(fields) == 0 {
		gen.StructAST[v.Name] = interfaceType
	} else {
		gen.StructAST[v.Name] = " struct {\n" + strings.Join(fields, "\n") + "}\n"
	}

	gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])
}

func isGoBuiltInType(typeName string) bool {
	_, builtIn := goBuiltinType[typeName]
	return builtIn
}

// GoGroup generates code for group XML schema in Go language syntax.
func (gen *CodeGenerator) GoGroup(v *Group) {
	if _, ok := gen.StructAST[v.Name]; ok {
		return
	}

	var fields []string

	fieldName := genGoFieldName(v.Name, true)
	if fieldName != v.Name {
		gen.ImportEncodingXML = true
		fields = append(fields, fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`", v.Name))
	}
	for _, element := range v.Elements {
		var plural string
		if element.Plural {
			plural = "[]"
		}
		fields = append(fields, fmt.Sprintf(
			"\t%s\t%s%s",
			genGoFieldName(element.Name, false),
			plural,
			genGoFieldType(getBaseFromSimpleType(trimNSPrefix(element.Type), gen.ProtoTree)),
		))
	}

	for _, group := range v.Groups {
		var plural string
		if group.Plural {
			plural = "[]"
		}
		fields = append(fields, fmt.Sprintf(
			"\t%s\t%s%s",
			genGoFieldName(group.Name, false),
			plural,
			genGoFieldType(getBaseFromSimpleType(trimNSPrefix(group.Ref), gen.ProtoTree)),
		))
	}

	if len(fields) > 0 {
		gen.StructAST[v.Name] = " struct {\n" + strings.Join(fields, "\n") + "\n}\n"
	} else {
		gen.StructAST[v.Name] = interfaceType
	}

	gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])
}

// GoAttributeGroup generates code for attribute group XML schema in Go language
// syntax.
func (gen *CodeGenerator) GoAttributeGroup(v *AttributeGroup) {
	if gen.seen(v.Name) {
		return
	}

	var fields []string

	fieldName := genGoFieldName(v.Name, true)
	if fieldName != v.Name {
		gen.ImportEncodingXML = true
		fields = append(fields, fmt.Sprintf("\tXMLName\txml.Name\t`xml:\"%s\"`", v.Name))
	}

	for _, attribute := range v.Attributes {
		var optional string
		if attribute.Optional {
			optional = omitEmpty
		}
		fields = append(fields, fmt.Sprintf(
			"\t%s\t%s\t`xml:\"%s,attr%s\"`",
			genGoFieldName(attribute.Name, false),
			genGoFieldType(getBaseFromSimpleType(trimNSPrefix(attribute.Type), gen.ProtoTree)),
			attribute.Name,
			optional,
		))
	}

	if len(fields) > 0 {
		gen.StructAST[v.Name] = " struct {\n" + strings.Join(fields, "\n") + "\n}\n"
	} else {
		gen.StructAST[v.Name] = interfaceType
	}

	gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])
}

// GoElement generates code for element XML schema in Go language syntax.
func (gen *CodeGenerator) GoElement(v *Element) {
	if gen.seen(v.Name) {
		return
	}

	if v.Name == v.Type {
		return
	}

	gen.StructAST[v.Name] = fmt.Sprintf(
		"\t%s%s\n",
		plural(v.Plural),
		genGoFieldType(getBaseFromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)),
	)

	fieldName := genGoFieldName(v.Name, false)

	gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])
}

// GoAttribute generates code for attribute XML schema in Go language syntax.
func (gen *CodeGenerator) GoAttribute(v *Attribute) {
	if gen.seen(v.Name) {
		return
	}

	gen.StructAST[v.Name] = fmt.Sprintf(
		"\t%s%s\n",
		plural(v.Plural),
		genGoFieldType(getBaseFromSimpleType(trimNSPrefix(v.Type), gen.ProtoTree)),
	)

	fieldName := genGoFieldName(v.Name, true)

	gen.Field += fmt.Sprintf("%stype %s%s", genFieldComment(fieldName, v.Doc), fieldName, gen.StructAST[v.Name])
}

func (gen *CodeGenerator) seen(name string) bool {
	_, ok := gen.StructAST[name]

	return ok
}
