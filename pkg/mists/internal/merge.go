package internal

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

func MergeNode(dst ast.Node, src ast.Node) (err error) {
	if dst.Type() != ast.MappingType {
		err = fmt.Errorf("cannot merge node into type %s", dst.Type().String())
		return
	}
	if src.Type() != ast.MappingType {
		err = fmt.Errorf("cannot merge node from type %s", src.Type().String())
		return
	}
	dstMap := dst.(*ast.MappingNode)
	srcMap := src.(*ast.MappingNode)
	err = MergeMappingNode(dstMap, srcMap)
	return
}

func MergeMappingNode(dst *ast.MappingNode, src *ast.MappingNode) (err error) {
	srcIter := src.MapRange()
	for srcIter.Next() {
		srcItem := srcIter.KeyValue()
		matched := false
		for i, dstItem := range dst.Values {
			if dstItem.Key.String() == srcItem.Key.String() {
				if dstItem.Value.Type() != srcItem.Value.Type() {
					err = fmt.Errorf("cannot merge node into %s, type was not matched", src.GetPath())
					return
				}
				matched = true
				if dstItem.Value.Type() == ast.MappingType {
					dstItemValue := dstItem.Value.(*ast.MappingNode)
					srcItemValue := srcItem.Value.(*ast.MappingNode)
					err = MergeMappingNode(dstItemValue, srcItemValue)
					if err != nil {
						return
					}
					break
				}
				dst.Values[i] = srcItem
				break
			}
		}
		if matched {
			continue
		}
		dst.Values = append(dst.Values, srcIter.KeyValue())
	}
	return
}
