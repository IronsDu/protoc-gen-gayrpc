package generator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/plugin"
	"gopkg.in/ahmetb/go-linq.v3"

	"github.com/IronsDu/protoc-gen-gayrpc/protoc-plugin"
)

type Method struct {
	name       string
	inputType  string
	outputType string
	messageID  int32
}

func (fd *Method) Name() string {
	return strings.Title(fd.name)
}

func (fd *Method) MethodName() string {
	return strings.ToLower(fd.name)
}

func (fd *Method) EnumName() string {
	return strings.ToLower(fd.name)
}

func (fd *Method) InputType() string {
	split := strings.Split(fd.inputType, ".")
	if len(split) == 0 {
		return ""
	}
	return split[len(split)-1]
}

func (fd *Method) OutputType() string {
	split := strings.Split(fd.outputType, ".")
	if len(split) == 0 {
		return ""
	}
	return split[len(split)-1]
}

func (fd *Method) MethodID() int32 {
	return fd.messageID
}

type Service struct {
	name string

	Methods []*Method
}

func (fd *Service) Name() string {
	return strings.Title(fd.name)
}

func (fd *Service) MethodsEnumName() string {
	return fmt.Sprintf("%sMsgID", fd.Name())
}

type FileDescriptor struct {
	fileName    string
	packageName string

	Services []*Service
}

func (fd *FileDescriptor) fileNameWithoutExt() string {
	return strings.TrimSuffix(strings.ToLower(fd.fileName), ".proto")
}

func (fd *FileDescriptor) GeneratedFilename() string {
	return fmt.Sprintf("%s.gayrpc.h", fd.fileNameWithoutExt())
}

func (fd *FileDescriptor) MacroName() string {
	return fmt.Sprintf("_%s_H", strings.ToUpper(fd.fileNameWithoutExt()))
}

func (fd *FileDescriptor) ModelFileName() string {
	return fmt.Sprintf("%s.pb.h", fd.fileNameWithoutExt())
}

func (fd *FileDescriptor) Namespace() string {
	return fd.fileNameWithoutExt()
}

func (fd *FileDescriptor) ContainerNamespace() string {
	return fmt.Sprintf("%s::", strings.Join(strings.Split(fd.packageName, "."), "::"))
}

// ---------------------------------------------------------------------------------------------------------------------

type Generator struct {
	*plugin.BaseGenerator
}

func (g *Generator) GenerateAllFiles() {
	for _, fd := range g.Files {
		if !linq.From(g.Request.FileToGenerate).Contains(fd.Descriptor.GetName()) {
			continue
		}
		d := &FileDescriptor{
			fileName:    fd.Descriptor.GetName(),
			packageName: fd.Descriptor.GetPackage(),
		}
		linq.From(fd.Descriptor.GetService()).SelectT(func(fd *descriptor.ServiceDescriptorProto) *Service {
			d := &Service{
				name: fd.GetName(),
			}
			linq.From(fd.GetMethod()).SelectT(func(fd *descriptor.MethodDescriptorProto) *Method {
				var options = make(map[string]string)
				linq.From(strings.Fields(proto.CompactTextString(fd.GetOptions()))).
					WhereT(func(s string) bool {
						return len(strings.TrimSpace(s)) > 0
					}).
					SelectT(func(s string) linq.KeyValue {
						lines := strings.Split(strings.TrimSpace(s), ":")
						if len(lines) != 2 {
							g.Fail("miss option in:", s)
						}
						return linq.KeyValue{
							Key:   lines[0],
							Value: lines[1],
						}
					}).ToMap(&options)

				var id int64
				if op, ok := options["51002"]; ok && op != "" {
					var err error
					id, err = strconv.ParseInt(op, 10, 32)
					if err != nil {
						g.Error(err, "method id illegal")
					}
				} else {
					g.Fail("miss message id of service function name:", *fd.Name)
				}

				return &Method{
					name:       fd.GetName(),
					inputType:  fd.GetInputType(),
					outputType: fd.GetOutputType(),
					messageID:  int32(id),
				}
			}).ToSlice(&d.Methods)
			for i, v := range d.Methods {
				if v.messageID == 0 {
					v.messageID = int32(i) + 1
				}
			}
			return d
		}).ToSlice(&d.Services)
		content, name := g.printFile(d)
		g.Response.File = append(g.Response.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String(name),
			Content: proto.String(content),
		})
	}
}

func (g *Generator) printFile(model *FileDescriptor) (content string, name string) {
	t, err := template.New("protoc-gen-gayrpc").Parse(CppTemplate)
	if err != nil {
		g.Error(err, "template parse failed")
	}

	w := bytes.NewBuffer(make([]byte, 0, 1024))
	err = t.Execute(w, model)
	if err != nil {
		g.Error(err, "execute template")
	}

	return w.String(), model.GeneratedFilename()
}

func NewGenerator(name string) *Generator {
	g := new(Generator)
	g.BaseGenerator = plugin.NewBaseGenerator(name)
	return g
}
