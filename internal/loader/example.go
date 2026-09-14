package loader

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/printer"
	"go/token"
	"strings"
)

// formatExample renders the example function body as Go source, without the
// outer braces.
func formatExample(fset *token.FileSet, ex *doc.Example) string {
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	node := &printer.CommentedNode{Node: ex.Code, Comments: withoutOutput(ex.Comments)}
	if err := cfg.Fprint(&buf, fset, node); err != nil {
		return ""
	}
	return unwrapBody(buf.String(), ex.Code)
}

// withoutOutput drops the comment group that declares the expected output.
func withoutOutput(groups []*ast.CommentGroup) []*ast.CommentGroup {
	var result []*ast.CommentGroup
	for _, g := range groups {
		text := strings.ToLower(strings.TrimSpace(g.Text()))
		if strings.HasPrefix(text, "output:") || strings.HasPrefix(text, "unordered output:") {
			continue
		}
		result = append(result, g)
	}
	return result
}

// unwrapBody strips the braces of a printed block statement and dedents the
// remaining lines by one tab.
func unwrapBody(src string, node ast.Node) string {
	if _, ok := node.(*ast.BlockStmt); !ok {
		return src
	}
	src = strings.TrimSpace(src)
	src = strings.TrimPrefix(src, "{")
	src = strings.TrimSuffix(src, "}")
	lines := strings.Split(strings.Trim(src, "\n"), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, "\t")
	}
	return strings.Join(lines, "\n") + "\n"
}
