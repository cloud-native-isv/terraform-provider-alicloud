package alicloud

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

func TestOssBucketCreateAppliesConfiguredTagsBeforeFinalRead(t *testing.T) {
	positions := ossBucketCreateCallPositions(t)

	setID := positions["d.SetId"]
	getTags := positions[`d.GetOk("tags")`]
	setTags := positions["resourceAliCloudOssBucketTaggingUpdate"]
	read := positions["resourceAliCloudOssBucketRead"]
	if setID == token.NoPos || getTags == token.NoPos || setTags == token.NoPos || read == token.NoPos {
		t.Fatalf("create lifecycle positions = %#v; want SetId, GetOk(tags), tagging update, and final Read", positions)
	}
	if !(setID < getTags && getTags < setTags && setTags < read) {
		t.Fatalf("create lifecycle positions = %#v; want SetId < GetOk(tags) < tagging update < final Read", positions)
	}
}

func ossBucketCreateCallPositions(t *testing.T) map[string]token.Pos {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "resource_alicloud_oss_bucket.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	positions := map[string]token.Pos{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != "resourceAliCloudOssBucketCreate" {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch called := call.Fun.(type) {
			case *ast.Ident:
				if called.Name == "resourceAliCloudOssBucketTaggingUpdate" || called.Name == "resourceAliCloudOssBucketRead" {
					positions[called.Name] = call.Pos()
				}
			case *ast.SelectorExpr:
				receiver, ok := called.X.(*ast.Ident)
				if !ok || receiver.Name != "d" {
					return true
				}
				if called.Sel.Name == "SetId" {
					positions["d.SetId"] = call.Pos()
				}
				if called.Sel.Name == "GetOk" && len(call.Args) == 1 {
					literal, ok := call.Args[0].(*ast.BasicLit)
					if !ok || literal.Kind != token.STRING {
						return true
					}
					value, err := strconv.Unquote(literal.Value)
					if err == nil && value == "tags" {
						positions[`d.GetOk("tags")`] = call.Pos()
					}
				}
			}
			return true
		})
		return positions
	}
	t.Fatal("resourceAliCloudOssBucketCreate not found")
	return nil
}
