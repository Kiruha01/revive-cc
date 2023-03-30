package rule

import (
	"fmt"
	"github.com/mgechev/revive/lint"
	"go/ast"
)

type FieldDeclaration struct{}

// Apply applies the rule to given file.
func (r *FieldDeclaration) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure
	var contractNames []string

	fileAst := file.AST
	walkerFindContractStruct := contractStruct{
		file:    file,
		fileAst: fileAst,
		addContractName: func(name string) {
			contractNames = append(contractNames, name)
		},
	}

	ast.Walk(walkerFindContractStruct, fileAst)

	walkerFieldDeclaration := lintFieldDeclaration{
		file:    file,
		fileAst: fileAst,
		onFailure: func(failure lint.Failure) {
			failures = append(failures, failure)
		},
		contractNames: contractNames,
	}

	ast.Walk(walkerFieldDeclaration, fileAst)

	return failures
}

// Name returns the rule name.
func (r *FieldDeclaration) Name() string {
	return "field-declaration"
}

type contractStruct struct {
	file            *lint.File
	fileAst         *ast.File
	addContractName func(string)
}

type lintFieldDeclaration struct {
	file          *lint.File
	fileAst       *ast.File
	onFailure     func(lint.Failure)
	contractNames []string
}

// AST Traversal
func (w contractStruct) Visit(node ast.Node) ast.Visitor {
	// Check if the node is a function declaration
	fn, ok := node.(*ast.FuncDecl)
	if !ok {
		return w
	}

	// Check if the function is the main function
	if fn.Name.Name != "main" {
		return w
	}

	// Traverse the function body
	ast.Inspect(fn, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			if typ, ok := cl.Type.(*ast.Ident); ok {
				w.addContractName(typ.Name)
			}
		}
		// Look for new() calls
		if ce, ok := n.(*ast.CallExpr); ok {
			if fn, ok := ce.Fun.(*ast.Ident); ok && fn.Name == "new" {
				if typ, ok := ce.Args[0].(*ast.Ident); ok {
					w.addContractName(typ.Name)
				}
			}
		}
		return true
	})
	return w
}

func (w lintFieldDeclaration) Visit(node ast.Node) ast.Visitor {
	if typeSpec, ok := node.(*ast.TypeSpec); ok {
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			for _, name := range w.contractNames {
				if typeSpec.Name.Name == name {
					for _, field := range structType.Fields.List {
						// Check if this is a public field
						if field.Names != nil {
							w.onFailure(lint.Failure{
								Confidence: 1,
								Failure:    fmt.Sprintf("field dectalation detected: %s; there should be no field declarations in the chaincode structure", field.Names[0].Name),
								Node:       field,
								Category:   "variables",
							})
						}
					}
				}
			}
		}
	}

	return w
}
