package gogroup

import (
	"go/parser"
	"go/token"
	"io"
	"strconv"
)

// An import statement with a group.
type groupedImport struct {
	// The zero-based starting and ending lines in the file.
	// The endLine is the last line of this statement, not the line after.
	startLine, endLine int

	// The import package path.
	path string

	// The import group.
	group int

	named bool
	name  string
}

// Allow sorting grouped imports.
type groupedImports []*groupedImport

func (gs groupedImports) Len() int {
	return len(gs)
}
func (gs groupedImports) Swap(i, j int) {
	gs[i], gs[j] = gs[j], gs[i]
}

//func (gs groupedImports) Less(i, j int) bool {
//	if gs[i].named == false && gs[j].named == true {
//		return true
//	}
//	if gs[i].named == true && gs[j].named == false {
//		return false
//	}
//	if gs[i].named == true && gs[j].named == true {
//		if gs[i].name < gs[j].name {
//			return true
//		}
//		if gs[i].name > gs[j].name {
//			return false
//		}
//	}
//	if gs[i].group < gs[j].group {
//		return true
//	}
//	if gs[i].group == gs[j].group && gs[i].path < gs[j].path {
//		return true
//	}
//	return false
//}

// Read import statements from a file, and assign them groups.
func (p *Processor) readImports(fileName string, r io.Reader) (groupedImports, error) {
	fset := token.NewFileSet()
	tree, err := parser.ParseFile(fset, fileName, r, parser.ImportsOnly|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	gs := groupedImports{}
	for _, ispec := range tree.Imports {
		named := false
		var name string
		if ispec.Name != nil {
			named = true
			name = ispec.Name.Name
		}
		var path string
		path, err = strconv.Unquote(ispec.Path.Value)
		if err != nil {
			return nil, err
		}

		startPos, endPos := ispec.Pos(), ispec.End()
		if ispec.Doc != nil {
			// Comments go with the following import statement.
			startPos = ispec.Doc.Pos()
		}

		file := fset.File(startPos)
		groupPath := path
		if named {
			groupPath = name + " " + groupPath
		}
		gs = append(gs, &groupedImport{
			path: path,
			// Line numbers are one-based in token.File.
			startLine: file.Line(startPos) - 1,
			endLine:   file.Line(endPos) - 1,
			group:     p.grouper.Group(groupPath),
			named:     named,
			name:      name,
		})
	}

	return gs, nil
}
