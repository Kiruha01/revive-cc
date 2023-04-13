package rule

import (
	"github.com/mgechev/revive/lint"
	"go/ast"
)

type CrossChannelInvocation struct{}

// Apply applies the rule to given file.
func (r *CrossChannelInvocation) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST

	walker := lintCrossChannel{
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
func (r *CrossChannelInvocation) Name() string {
	return "cross-channel-invocation"
}

type lintCrossChannel struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

func (w lintCrossChannel) Visit(node ast.Node) ast.Visitor {
	if callExpr, ok := node.(*ast.CallExpr); ok {
		if ident, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if ident.Sel.Name == "InvokeChaincode" {
				if len(callExpr.Args) == 3 {
					lit, ok := callExpr.Args[2].(*ast.BasicLit)
					if ok && lit.Value != `""` || !ok {
						w.onFailure(lint.Failure{
							Confidence: 1,
							Node:       node,
							Failure:    "Potential vulnerability in cross-channel invocation detected",
							Category:   "security",
						})
					}
				}
			}
		}
	}

	return w
}
