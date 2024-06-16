package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Imports []string

type Module struct {
	Name string
	Imports
}

func NewModule(name string) Module {
	imps := make(Imports, 0)
	return Module{name, imps}
}
func (m Module) String() string {
	res := fmt.Sprintf("%s:\n", m.Name)
	for _, imp := range m.Imports {
		res += fmt.Sprintf("\t%s\n", imp)
	}
	res += "===============================================\n"
	return res
}

func (m Modules) calculateRank(mod Module, visited map[string]bool) int {
	if visited[name] {
		return 0
	}
	visited[name] = true

	moduleImports, exists := modules[name]
	if !exists || len(moduleImports) == 0 {
		return 0
	}

	maxRank := 0
	for _, imp := range moduleImports {
		rank := calculateRank(imp, modules, visited)
		if rank > maxRank {
			maxRank = rank
		}
	}
	return maxRank + 1
}

type Modules []Module

func NewModules() Modules {
	return make(Modules, 0)
}

func (m Modules) String() string {
	res := ""
	for _, mod := range m {
		res += mod.String()
	}
	return res
}

func (m Modules) Sort() {
	sort.Slice(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})
}

func (m Modules) Add(mod Module) {
	m = append(m, mod)
}

func (m Modules) Contains(name string) bool {
	for _, mod := range m {
		if mod.Name == name {
			return true
		}
	}
	return false
}

func main() {
	dir := "/home/andrejjj/Projects/Aos/A2-oberon/source"
	// dir := "/home/andrejjj/Projects/Aos/xlam/11"
	modules := parseModules(dir)
	modules.Sort()
	fmt.Println(modules)
	generateDOT(modules, "modules_graph.dot")
}

func parseModules(dir string) Modules {
	modules := NewModules()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".Mod") || strings.HasSuffix(info.Name(), ".mod")) {
			mod, err := parseModuleFile(path)
			if err != nil || mod.Name == "" {
				return fmt.Errorf("error parsing module file %q: module name not found", path)
			}
			modules.Add(mod)
			for _, imp := range mod.Imports {
				if !modules.Contains(imp) {
					modules.Add(NewModule(imp))
				}
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", dir, err)
		return nil
	}
	return modules
}

func parseModuleFile(path string) (Module, error) {
	file, err := os.Open(path)
	if err != nil {
		return NewModule(""), err
	}
	defer file.Close()

	mod := NewModule("")

	// Читаем весь файл
	fileInfo, err := file.Stat()
	if err != nil {
		return "", nil, err
	}
	fileSize := fileInfo.Size()
	fileContent := make([]byte, fileSize)
	_, err = file.Read(fileContent)
	if err != nil {
		return "", nil, err
	}

	text := string(fileContent)

	// Используем флаг (?s) чтобы . соответствовал любому символу, включая новую строку
	reModule := regexp.MustCompile(`(?is)MODULE\s+([\/\.\s\w\(\*\)]+);`)
	// reModule := regexp.MustCompile(`(?is)MODULE\s+([^;]+?);`)
	reImport := regexp.MustCompile(`(?is)IMPORT\s+([^;]+?);`)
	reComment := regexp.MustCompile(`(?s)\(\*\*?.*?\*\)`)

	moduleMatches := reModule.FindStringSubmatch(text)
	if moduleMatches != nil {
		moduleName = reComment.ReplaceAllString(moduleMatches[1], "")
		// fmt.Println("Module name: ", moduleName)
	}

	importMatches := reImport.FindAllStringSubmatch(text, -1)
	for _, match := range importMatches {
		imports := match[1]
		// imports = strings.ReplaceAll(imports, "\n", "")
		imports = string(reComment.ReplaceAll([]byte(imports), []byte("")))
		// fmt.Println("Import matches: ", imports)
		importList := regexp.MustCompile(`[,]+`).Split(imports, -1)
		for _, imp := range importList {
			imp = strings.TrimSpace(imp)
			impRight := strings.Split(imp, ":=")
			if len(impRight) > 1 {
				imp = strings.TrimSpace(impRight[1])
			} else {
				imp = strings.TrimSpace(impRight[0])
			}
			if imp != "" {
				// fmt.Println(imp)
				moduleImports = append(moduleImports, reComment.ReplaceAllString(imp, ""))
			}
		}
	}
	return mod, nil
}

func generateDOT(modules Modules, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating DOT file:", err)
		return
	}
	defer f.Close()

	f.WriteString("digraph G {\n")
	f.WriteString("  node [shape=box, style=filled, color=lightblue];\n")

	ranks := make(map[int][]string)
	visited := make(map[string]bool)
	var maxRank int

	for _, mod := range modules {
		if !visited[mod.Name] {
			rank := calculateRank(mod, modules, visited)
			ranks[rank] = append(ranks[rank], mod.Name)
			if rank > maxRank {
				maxRank = rank
			}
		}
	}

	for rank := 0; rank <= maxRank; rank++ {
		if moduleNames, exists := ranks[rank]; exists {
			f.WriteString(fmt.Sprintf("  { rank=same; %s }\n", strings.Join(moduleNames, " ")))
		}
	}

	for _, mod := range modules {
		for _, imp := range mod.Imports {
			if modules.Contains(imp) {
				f.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", mod.Name, imp))
			}
		}
	}

	f.WriteString("}\n")
}
