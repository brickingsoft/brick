package configs_test

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/brickingsoft/brick/pkg/configs"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

var (
	content1 = `
a: "sss"
hello:
 name: world
 age: 11`

	content2 = `x: "xxx"`

	content = `defaults: &defaults #Anchor *ast.AnchorNode
  adapter:  postgres
  host:     localhost

development:
  database: myapp_development
  <<: *defaults # Alias *ast.AliasNode

test:
  database: myapp_test
  <<: *defaults`
)

func TestMerge(t *testing.T) {
	var (
		tb = `hello:
 name: "b"`
	)

	target, targetErr := parser.ParseBytes([]byte(tb), 0)
	if targetErr != nil {
		t.Fatal(targetErr)
	}
	targetNode := target.Docs[0].Body.(*ast.MappingNode)
	file, parseErr := parser.ParseBytes([]byte(content2), 0)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	doc := file.Docs[0]
	node := doc.Body.(*ast.MappingNode)
	iter := node.MapRange()
	for iter.Next() {
		fmt.Println(iter.Key(), iter.KeyValue().Type(), iter.Value().Type())
		if iter.Key().String() == "hello" {
			vn := iter.Value().(*ast.MappingNode)
			for _, value := range targetNode.Values {
				vnn := value.Value.(*ast.MappingNode)
				vn.Merge(vnn)
			}
		}
	}
	//node.Merge(targetNode)

	p, pErr := io.ReadAll(node)
	if pErr != nil {
		t.Fatal(pErr)
	}
	t.Log(string(p))
}

func TestUnm(t *testing.T) {
	b := []byte("ssss")

	s := ""

	err := yaml.Unmarshal(b, &s)
	t.Log(string(b), s, err)

	_, _ = yaml.Marshal(b)

}

func TestParse(t *testing.T) {
	file, parseErr := parser.ParseBytes([]byte(content), 0)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	t.Log("name:", file.Name, "docs:", len(file.Docs))

	for _, doc := range file.Docs {
		t.Log("path:", doc.GetPath())
		t.Log("path:", doc.Type())

		t.Log("body:", reflect.TypeOf(doc.Body), doc.Body.GetPath())
		node := doc.Body.(*ast.MappingNode)

		t.Log("mapping:", node.Path)
		iter := node.MapRange()
		for iter.Next() {
			key := iter.Key()
			kb, _ := io.ReadAll(key)
			vb, _ := io.ReadAll(iter.Value())
			value := iter.Value()
			t.Log("key:", key.Type(), reflect.TypeOf(key), key, string(kb))
			t.Log("value:", value.Type(), reflect.TypeOf(value), value)
			t.Log("---")
			t.Log(string(vb))
			t.Log("---")

			if value.Type() == ast.MappingType {
				mv := value.(*ast.MappingNode)
				for _, valueNode := range mv.Values {
					t.Log("->value:", valueNode.Value.Type(), reflect.TypeOf(valueNode.Value), valueNode)

				}
			}
			if value.Type() == ast.AnchorType {
				av := value.(*ast.AnchorNode)
				t.Log("anchor:", av.Value.Type(), reflect.TypeOf(av.Value), av.Value)
			}
		}
		//node.Merge()
		//yaml.NodeToValue()
		//node := doc.Body.GetPath()
	}

}

type Target struct {
	Hello configs.Raw `yaml:"hello"`
}

func TestDecode(t *testing.T) {
	target := &Target{}
	err := yaml.Unmarshal([]byte(content), target)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("target:", string(target.Hello))

	target.Hello = nil
	b, encodeErr := yaml.Marshal(target)
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	t.Log("target:", string(b))
}

func TestMergeAnchor(t *testing.T) {
	var (
		anchorBytes = []byte(`defaults: &defaults #Anchor *ast.AnchorNode
  adapter:  postgres
  host:     localhost`)
		dstBytes = []byte(`development:
  database: myapp_development
  <<: *defaults # Alias *ast.AliasNode
  a: "abc"
h: 11"`)
	)

	type DB struct {
		Adapter string `yaml:"adapter"`
		Host    string `yaml:"host"`
	}
	type Ac struct {
		Defaults DB `yaml:"defaults"`
	}

	ac := &Ac{}
	decodeErr := yaml.Unmarshal(anchorBytes, ac)
	if decodeErr != nil {
		t.Fatal(decodeErr)
	}
	t.Log(*ac)

	dstFile, dstFileErr := parser.ParseBytes(dstBytes, 0)
	if dstFileErr != nil {
		t.Fatal(dstFileErr)
	}
	dst := dstFile.Docs[0].Body.(*ast.MappingNode)

	srcFile, srcFileErr := parser.ParseBytes(anchorBytes, 0)
	if srcFileErr != nil {
		t.Fatal(srcFileErr)
	}
	src := srcFile.Docs[0].Body.(*ast.MappingNode)

	sr := src.MapRange()
	for sr.Next() {
		snk := sr.Key()
		snv := sr.Value()
		t.Log("snk:", snk, snk.Type(), reflect.TypeOf(snk))
		t.Log("snv:", snv.Type(), reflect.TypeOf(snv))
		if snv.Type() == ast.AnchorType {
			anchor := snv.(*ast.AnchorNode)
			t.Log("anchor:", anchor.Name, anchor.Name.Type())
			anchorValue := anchor.Value.(*ast.MappingNode)
			dr := dst.MapRange()
			for dr.Next() {
				dnk := dr.Key()
				dnv := dr.Value()
				t.Log("dnk:", dnk, dnk.Type(), reflect.TypeOf(dnk))
				t.Log("dnv:", dnv.Type(), reflect.TypeOf(dnv))

				if dnv.Type() == ast.MappingType {
					dnn := dnv.(*ast.MappingNode)
					dnr := dnn.MapRange()
					for dnr.Next() {
						dnnk := dnr.Key()
						dnnv := dnr.Value()
						t.Log("dnnk:", dnnk, dnnk.Type(), dnnk.IsMergeKey())
						if dnnk.Type() == ast.MergeKeyType {
							mk := dnnk.(*ast.MergeKeyNode)
							mv := dnnv.(*ast.AliasNode)
							t.Log("-mk:", dnnk.String(), dnnv.String(), dnr.KeyValue().String())
							t.Log("dmk:", mk.String(), mk.Type(), reflect.TypeOf(mk), mk.Token.Position.Column)
							t.Log("dmv:", mv, mv.Value.String(), mv.Value.Type(), reflect.TypeOf(mv.Value))
							dnn.Values = append(dnn.Values[:1], dnn.Values[2:]...)
							dnn.Merge(anchorValue)
						}
					}
				}

			}
		}
	}

	//t.Log("dst:", dst)
	fmt.Println(string(dstBytes))
	fmt.Println(dst)

}
