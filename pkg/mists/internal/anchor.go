package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml/ast"
)

func AliasReplaceAnchor(ctx context.Context, node ast.Node, anchors map[string]*ast.AnchorNode) (err error) {
	if node == nil || len(anchors) == 0 {
		return
	}
	if node.Type() == ast.AnchorType {
		anchor := node.(*ast.AnchorNode)
		if ctx.Value(anchor.Name) != nil {
			err = fmt.Errorf("replace aliases failed, anchor is circular referenced of %s", anchor.Name)
			return
		}
		ctx = context.WithValue(ctx, anchor.Name, anchor)
		err = AliasReplaceAnchor(ctx, anchor.Value, anchors)
		return
	}
	if node.Type() == ast.MappingType {
		mapping := node.(*ast.MappingNode)
		iter := mapping.MapRange()
		idx := -1
		for iter.Next() {
			idx++
			key := iter.Key()
			if key.IsMergeKey() {
				val := iter.Value()
				alias, isAlias := val.(*ast.AliasNode)
				if !isAlias {
					err = fmt.Errorf("cannot replace aliases, value of '%s' is not alias type", strings.TrimSpace(iter.KeyValue().String()))
					return
				}
				aliasName := alias.Value.String()
				anchor, exists := anchors[aliasName]
				if !exists {
					err = fmt.Errorf("cannot replace aliases, anchor of '%s' is missing", strings.TrimSpace(iter.KeyValue().String()))
					return
				}
				if err = AliasReplaceAnchor(ctx, anchor, anchors); err != nil {
					return
				}
				anchorValue, isMapping := anchor.Value.(*ast.MappingNode)
				if !isMapping {
					err = fmt.Errorf("cannot replace aliases, alias of '%s' has anchor but anchor is not mapping type", strings.TrimSpace(iter.KeyValue().String()))
					return
				}
				mapping.Values = append(mapping.Values[:idx], mapping.Values[idx+1:]...)
				mapping.Merge(anchorValue)
				continue
			}
			if iter.Value().Type() == ast.MappingType {
				value := iter.Value().(*ast.MappingNode)
				err = AliasReplaceAnchor(ctx, value, anchors)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func Anchors(node ast.Node) (nodes map[string]*ast.AnchorNode, err error) {
	nodes, err = anchors(node)
	if err != nil {
		return
	}
	if len(nodes) == 0 {
		return
	}
	ctx := context.TODO()
	for key, value := range nodes {
		err = AliasReplaceAnchor(ctx, value, nodes)
		if err != nil {
			return
		}
		nodes[key] = value
	}
	return
}

func anchors(node ast.Node) (nodes map[string]*ast.AnchorNode, err error) {
	if node == nil {
		return
	}
	if node.Type() == ast.AnchorType {
		anchor := node.(*ast.AnchorNode)
		if anchor.Value == nil {
			return
		}
		anchorNode, isMapping := anchor.Value.(*ast.MappingNode)
		if !isMapping {
			err = fmt.Errorf("get anchors from node failed, anchor is not mapping type of %s", anchor.Name.String())
			return
		}
		subs, subsErr := anchors(anchorNode)
		if subsErr != nil {
			err = subsErr
			return
		}
		nodes = make(map[string]*ast.AnchorNode)
		if len(subs) > 0 {
			if _, has := subs[anchor.Name.String()]; has {
				err = fmt.Errorf("get anchors from node failed, anchor is duplicated of %s", anchor.Name.String())
				return
			}
			for sk, sv := range subs {
				nodes[sk] = sv
			}
		}
		nodes[anchor.Name.String()] = anchor
		return
	}
	if node.Type() != ast.MappingType {
		return
	}
	mapping := node.(*ast.MappingNode)
	iter := mapping.MapRange()
	for iter.Next() {
		value := iter.Value()
		subs, subsErr := anchors(value)
		if subsErr != nil {
			err = subsErr
			return
		}
		if len(subs) == 0 {
			continue
		}
		if nodes == nil {
			nodes = make(map[string]*ast.AnchorNode)
		}
		for sk, sv := range subs {
			if _, has := nodes[sk]; has {
				err = fmt.Errorf("get anchors from node failed, anchor is duplicated of %s", sk)
				return
			}
			nodes[sk] = sv
		}
	}
	return
}
