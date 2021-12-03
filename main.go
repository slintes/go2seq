package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {

	path := "./test"
	startPkg := "test"
	startFunc := "start"

	//path := "../../dragonfly/poison-pill/"
	//startPkg := "controllers"
	//startFunc := "Reconcile"

	files := token.NewFileSet()
	funcs := make(map[string]map[string]*function)
	walkDirs(path, funcs, files)

	for _, pkg := range funcs {
		for _, f := range pkg {
			if f.pkg == startPkg && f.name == startFunc {
				fmt.Printf("title sequence diagram for package %s, function %s\n\n", f.pkg, f.name)
				printCall(f)
				os.Exit(0)
			}
		}
	}
}

func printCall(current *function) {
	for _, call := range current.calls {
		fmt.Printf("%s->%s: %s\n", current.pkg, call.pkg, call.name)
		printCall(call)
	}
}

func walkDirs(path string, funcs functions, files *token.FileSet) {
	visit := func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			if entry.Name() == "vendor" {
				return fs.SkipDir
			}
			parseDir(path, funcs, files)
		}
		return nil
	}

	err := filepath.WalkDir(path, visit)
	if err != nil {
		log.Fatal(err)
	}
}

func parseDir(path string, funcs functions, files *token.FileSet) {
	pkgs, err := parser.ParseDir(files, path, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	//spew.Dump(f)

	v := visitor{
		status: &status{
			funcs:      funcs,
			isFuncCall: false,
		},
	}
	for _, pkg := range pkgs {
		//fmt.Printf("pkg: %s\n", pkgName)
		for _, file := range pkg.Files {
			//fmt.Printf("  file: %s\n", fileName)
			ast.Walk(v, file)

			// end of file, close potentially dangling function call
			v.endFuncCall()

			//fmt.Printf("\n\n")
		}
	}
}

type status struct {
	funcs           functions
	currentPackage  string
	currentFunction *function
	isFuncCall      bool
	funcCallIdents  []string
}

type visitor struct {
	status *status
}

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	//fmt.Printf("--- %T\n", n)

	if _, ok := n.(*ast.Ident); !ok {
		v.endFuncCall()
	}

	switch t := n.(type) {
	case *ast.File:
		//fmt.Printf("*** package %s\n", t.Name)
		v.status.currentPackage = t.Name.Name
	case *ast.Ident:
		//fmt.Printf("*** ident name %s, obj %#v\n", t.Name, t.Obj)
		if !v.status.isFuncCall {
			break
		}
		if v.status.funcCallIdents == nil {
			v.status.funcCallIdents = make([]string, 0)
		}
		v.status.funcCallIdents = append(v.status.funcCallIdents, t.Name)
	case *ast.CallExpr:
		//fmt.Printf("*** call expr %#v\n", t)
		v.startFuncCall()
	case *ast.FuncDecl:
		//fmt.Printf("*** func decl %s\n", t.Name)
		v.status.currentFunction = addFunction(v.status.funcs, v.status.currentPackage, t.Name.Name)
		//v.Visit(t.Body)
	case *ast.FuncType:
		//fmt.Printf("*** func call %#v\n", t)
	case *ast.BlockStmt:
		//fmt.Printf("*** block\n")
		//for _, stmt := range t.List {
		//	v.Visit(stmt)
		//}
	}

	return v
}

func (v *visitor) startFuncCall() {
	v.status.isFuncCall = true
	v.status.funcCallIdents = make([]string, 0)
}

func (v *visitor) endFuncCall() {
	if v.status.currentFunction == nil || !v.status.isFuncCall || len(v.status.funcCallIdents) < 1 {
		return
	}
	var pkg, name string
	if len(v.status.funcCallIdents) == 1 {
		pkg = v.status.currentPackage
		name = v.status.funcCallIdents[0]
	} else {
		pkg = v.status.funcCallIdents[len(v.status.funcCallIdents)-2]
		name = v.status.funcCallIdents[len(v.status.funcCallIdents)-1]
	}
	function := addFunction(v.status.funcs, pkg, name)
	v.status.currentFunction.calls = append(v.status.currentFunction.calls, function)

	v.status.isFuncCall = false
}

func addFunction(funcs functions, pkg, name string) *function {
	if funcs[pkg] == nil {
		funcs[pkg] = make(map[string]*function, 0)
	}
	if funcs[pkg][name] == nil {
		funcs[pkg][name] = &function{
			pkg:   pkg,
			name:  name,
			calls: make([]*function, 0),
		}
	}
	return funcs[pkg][name]
}

type functions map[string]map[string]*function

type function struct {
	pkg   string
	name  string
	calls []*function
}
