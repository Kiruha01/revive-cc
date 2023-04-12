package rule

import (
	"fmt"
	"github.com/mgechev/revive/lint"
	"go/ast"
	"go/token"
)

type ReifiedObjectAddress struct{}

// Apply applies the rule to given file.
func (r *ReifiedObjectAddress) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST

	walker := walkerObjectAddress{
		file:    file,
		fileAst: fileAst,
		onFailure: func(failure lint.Failure) {
			failures = append(failures, failure)
		},
	}

	ast.Walk(walker, fileAst)

	return failures
}

// Name returns the rule name.
func (r *ReifiedObjectAddress) Name() string {
	return "reified-object-address"
}

type walkerObjectAddress struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

func (w walkerObjectAddress) Visit(node ast.Node) ast.Visitor {
	if fn, ok := node.(*ast.FuncDecl); ok {
		// Address reified allowed in main function
		if fn.Name.Name == "main" {
			return w
		}
	}

	switch n := node.(type) {
	case *ast.AssignStmt:
		// p := new(int)
		for _, rhs := range n.Rhs {
			if call, ok := rhs.(*ast.CallExpr); ok {
				if fun, ok := call.Fun.(*ast.Ident); ok {
					if fun.Name == "new" && len(call.Args) == 1 {
						w.onFailure(lint.Failure{
							Confidence: 0.8,
							Failure:    fmt.Sprint("found using reified object address with new() statement"),
							Node:       fun,
							Category:   "warning",
						})
					}
				}
			}
		}
		// p := &int{}
		for _, rhs := range n.Rhs {
			if uop, ok := rhs.(*ast.UnaryExpr); ok {
				if uop.Op == token.AND {
					w.onFailure(lint.Failure{
						Confidence: 0.8,
						Failure:    fmt.Sprintf("found using reified object address using &"),
						Node:       uop,
						Category:   "warning",
					})
				}
			}
		}
	case *ast.ValueSpec:
		// var p *int
		if star, ok := n.Type.(*ast.StarExpr); ok {
			if _, ok := star.X.(*ast.Ident); ok {
				for _, name := range n.Names {
					if name.Obj != nil && name.Obj.Kind == ast.Var {
						w.onFailure(lint.Failure{
							Confidence: 0.8,
							Failure:    fmt.Sprintf("found using reified object address using *"),
							Node:       name,
							Category:   "warning",
						})
					}
				}
			}
		}
	}

	return w
}
