package plugin

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/pkg/errors"
)

type BaseDescriptor struct {
	File *descriptor.FileDescriptorProto
}

type Descriptor struct {
	BaseDescriptor
	Descriptor *descriptor.DescriptorProto
	Parent     *Descriptor
	Nested     []*Descriptor
	Enums      []*EnumDescriptor
	Name       []string
	Index      int
	Path       string
	Location   *descriptor.SourceCodeInfo_Location
}

type EnumDescriptor struct {
	BaseDescriptor
	Descriptor *descriptor.EnumDescriptorProto
	Parent     *Descriptor
	Name       []string
	Index      int
	Path       string
	Location   *descriptor.SourceCodeInfo_Location
}

type FileDescriptor struct {
	Descriptor  *descriptor.FileDescriptorProto
	MessageType []*Descriptor
	EnumType    []*EnumDescriptor
}

type BaseGenerator struct {
	name string

	Request  *plugin.CodeGeneratorRequest
	Response *plugin.CodeGeneratorResponse

	Parameters map[string]string

	Files []*FileDescriptor
}

func NewBaseGenerator(name string) *BaseGenerator {
	g := new(BaseGenerator)
	g.name = name
	g.Request = new(plugin.CodeGeneratorRequest)
	g.Response = new(plugin.CodeGeneratorResponse)
	return g
}

func (g *BaseGenerator) Error(err error, messages ...string) {
	s := strings.Join(messages, " ") + ":" + err.Error()
	log.Print(fmt.Sprintf("%s: error:", g.name), s)
	os.Exit(1)
}

func (g *BaseGenerator) Fail(messages ...string) {
	s := strings.Join(messages, " ")
	log.Print(fmt.Sprintf("%s: error:", g.name), s)
	os.Exit(1)
}

func (g *BaseGenerator) WrapTypes() {
	g.Files = make([]*FileDescriptor, 0, len(g.Request.ProtoFile))
	for _, f := range g.Request.ProtoFile {
		fd, err := wrapFileDescriptor(f)
		if err != nil {
			g.Fail(err.Error())
		}
		g.Files = append(g.Files, fd)
	}
}

func (g *BaseGenerator) CommandLineParameters(parameter string) {
	g.Parameters = make(map[string]string)
	for _, p := range strings.Split(parameter, ",") {
		if i := strings.Index(p, "="); i < 0 {
			g.Parameters[p] = ""
		} else {
			g.Parameters[p[0:i]] = p[i+1:]
		}
	}
}

func wrapFileDescriptor(f *descriptor.FileDescriptorProto) (*FileDescriptor, error) {
	comments := wrapComments(f)
	descriptors := wrapDescriptors(f, comments)
	if err := buildNestedDescriptors(descriptors); err != nil {
		return nil, err
	}
	enums := wrapEnumDescriptors(f, descriptors, comments)
	if err := buildNestedEnums(descriptors, enums); err != nil {
		return nil, err
	}
	fd := &FileDescriptor{
		Descriptor:  f,
		MessageType: descriptors,
		EnumType:    enums,
	}
	return fd, nil
}

func wrapComments(f *descriptor.FileDescriptorProto) map[string]*descriptor.SourceCodeInfo_Location {
	comments := make(map[string]*descriptor.SourceCodeInfo_Location)
	for _, location := range f.GetSourceCodeInfo().GetLocation() {
		var paths []string
		for _, n := range location.Path {
			paths = append(paths, strconv.Itoa(int(n)))
		}
		comments[strings.Join(paths, ",")] = location
	}
	return comments
}

func wrapDescriptors(f *descriptor.FileDescriptorProto, comments map[string]*descriptor.SourceCodeInfo_Location) []*Descriptor {
	descriptors := make([]*Descriptor, 0, len(f.MessageType)+10)
	for i, d := range f.MessageType {
		descriptors = wrapThisDescriptor(descriptors, d, nil, f, i, comments)
	}
	return descriptors
}

func wrapThisDescriptor(
	descriptors []*Descriptor,
	descriptor *descriptor.DescriptorProto,
	parent *Descriptor,
	f *descriptor.FileDescriptorProto,
	index int,
	comments map[string]*descriptor.SourceCodeInfo_Location) []*Descriptor {
	descriptors = append(descriptors, newDescriptor(descriptor, parent, f, index, comments))
	me := descriptors[len(descriptors)-1]
	for i, nested := range descriptor.NestedType {
		descriptors = wrapThisDescriptor(descriptors, nested, me, f, i, comments)
	}
	return descriptors
}

func newDescriptor(
	descriptor *descriptor.DescriptorProto,
	parent *Descriptor,
	f *descriptor.FileDescriptorProto,
	index int,
	comments map[string]*descriptor.SourceCodeInfo_Location) *Descriptor {
	d := &Descriptor{
		BaseDescriptor: BaseDescriptor{
			File: f,
		},
		Descriptor: descriptor,
		Parent:     parent,
		Index:      index,
	}
	if parent == nil {
		d.Path = fmt.Sprintf("%d,%d", 4, index)
	} else {
		d.Path = fmt.Sprintf("%s,%d,%d", parent.Path, 3, index)
		d.Name = append(d.Name, parent.Name...)
	}
	d.Name = append(d.Name, descriptor.GetName())
	d.Location = comments[d.Path]

	return d
}

func buildNestedDescriptors(descriptors []*Descriptor) error {
	for _, d := range descriptors {
		if len(d.Descriptor.NestedType) != 0 {
			for _, nest := range descriptors {
				if nest.Parent == d {
					d.Nested = append(d.Nested, nest)
				}
			}
			if len(d.Nested) != len(d.Descriptor.NestedType) {
				return errors.New("internal error: nesting failure for " + d.Descriptor.GetName())
			}
		}
	}
	return nil
}

func buildNestedEnums(descriptors []*Descriptor, enums []*EnumDescriptor) error {
	for _, d := range descriptors {
		if len(d.Descriptor.EnumType) != 0 {
			for _, enum := range enums {
				if enum.Parent == d {
					d.Enums = append(d.Enums, enum)
				}
			}
			if len(d.Enums) != len(d.Descriptor.EnumType) {
				return errors.New("internal error: enum nesting failure for " + d.Descriptor.GetName())
			}
		}
	}
	return nil
}

func newEnumDescriptor(
	descriptor *descriptor.EnumDescriptorProto,
	parent *Descriptor,
	f *descriptor.FileDescriptorProto,
	index int,
	comments map[string]*descriptor.SourceCodeInfo_Location) *EnumDescriptor {
	d := &EnumDescriptor{
		BaseDescriptor: BaseDescriptor{
			File: f,
		},
		Descriptor: descriptor,
		Parent:     parent,
		Index:      index,
	}
	if parent == nil {
		d.Path = fmt.Sprintf("%d,%d", 5, index)
	} else {
		d.Path = fmt.Sprintf("%s,%d,%d", parent.Path, 4, index)
		d.Name = append(d.Name, parent.Name...)
	}
	d.Name = append(d.Name, descriptor.GetName())
	d.Location = comments[d.Path]

	return d
}

func wrapEnumDescriptors(
	f *descriptor.FileDescriptorProto,
	descriptors []*Descriptor,
	comments map[string]*descriptor.SourceCodeInfo_Location) []*EnumDescriptor {
	enums := make([]*EnumDescriptor, 0, len(f.EnumType)+10)
	for i, enum := range f.EnumType {
		enums = append(enums, newEnumDescriptor(enum, nil, f, i, comments))
	}
	for _, nested := range descriptors {
		for i, enum := range nested.Descriptor.EnumType {
			enums = append(enums, newEnumDescriptor(enum, nested, f, i, comments))
		}
	}
	return enums
}
