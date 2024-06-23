package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	config "github.com/andrqxa/A2-modules-graph/configs"
	"github.com/spf13/viper"
)

type Imports []string

type Module struct {
	Name string
	Imports
}

// Constructor for Module
func NewModule(name string) Module {
	imps := make(Imports, 0)
	return Module{name, imps}
}

// String representation of Module
func (m Module) String() string {
	res := fmt.Sprintf("%s:\n", m.Name)
	for _, imp := range m.Imports {
		res += fmt.Sprintf("\t%s\n", imp)
	}
	res += "===============================================\n"
	return res
}

// Check if module has no imports
func (m Module) IsBareModule() bool {
	return len(m.Imports) == 0
}

// Add an import to the module
func (m *Module) AddImport(imp string) {
	(*m).Imports = append((*m).Imports, imp)
}

type Modules []Module

// Constructor for Modules
func NewModules() Modules {
	return make(Modules, 0)
}

// String representation of Modules
func (m Modules) String() string {
	res := ""
	for _, mod := range m {
		res += mod.String()
	}
	return res
}

// Sort modules by name
func (m Modules) Sort() {
	sort.Slice(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})
}

// Add a module to the collection
func (m *Modules) Add(mod Module) {
	*m = append(*m, mod)
}

// Check if a module is contained in the collection
func (m Modules) Contains(name string) bool {
	for _, mod := range m {
		if mod.Name == name {
			return true
		}
	}
	return false
}

// Calculate the rank of a module
func (m Modules) calculateRank(mod Module, visited map[string]bool) int {
	if visited[mod.Name] {
		return 0
	}
	visited[mod.Name] = true

	if !m.Contains(mod.Name) || mod.IsBareModule() {
		return 0
	}

	maxRank := 0
	for _, mdl := range m {
		rank := m.calculateRank(mdl, visited)
		if rank > maxRank {
			maxRank = rank
		}
	}
	return maxRank + 1
}

// Generate a DOT file for visualization
func (ms Modules) generateDOT(filename string) {
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

	// Calculate ranks for modules
	for _, mod := range ms {
		if !visited[mod.Name] {
			rank := ms.calculateRank(mod, visited)
			ranks[rank] = append(ranks[rank], mod.Name)
			if rank > maxRank {
				maxRank = rank
			}
		}
	}

	// Write nodes with the same rank
	for rank := 0; rank <= maxRank; rank++ {
		if moduleNames, exists := ranks[rank]; exists {
			f.WriteString("  { rank=same; ")
			for _, name := range moduleNames {
				f.WriteString(fmt.Sprintf("\"%s\" ", name))
			}
			f.WriteString("}\n")
		}
	}

	// Write edges
	for _, mod := range ms {
		for _, imp := range mod.Imports {
			if ms.Contains(imp) {
				f.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", mod.Name, imp))
			}
		}
	}

	f.WriteString("}\n")
}

func main() {

	err := config.ConfigViper()
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	dir := viper.GetString("app.test-dir")
	txt := viper.GetString("app.output-txt")

	modules := parseModules(dir)
	modules.Sort()

	outputFile, err := os.Create(txt)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	_, err = outputFile.WriteString(modules.String())
	if err != nil {
		fmt.Printf("Error writing to output file: %v\n", err)
		os.Exit(1)
	}

	modules.generateDOT("modules_graph.dot")
}

// Parse modules from the given directory
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
			// for _, imp := range mod.Imports {
			// 	if !modules.Contains(imp) {
			// 		modules.Add(NewModule(imp))
			// 	}
			// }
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", dir, err)
		return nil
	}
	return modules
}

// Parse a module file to extract module name and imports
func parseModuleFile(path string) (Module, error) {
	file, err := os.Open(path)
	if err != nil {
		return NewModule(""), err
	}
	defer file.Close()

	mod := NewModule("")

	// Read the entire file
	fileInfo, err := file.Stat()
	if err != nil {
		return mod, err
	}
	fileSize := fileInfo.Size()
	fileContent := make([]byte, fileSize)
	_, err = file.Read(fileContent)
	if err != nil {
		return mod, err
	}

	text := string(fileContent)

	// Use regex to find module name and imports
	reModule := regexp.MustCompile(`(?is)MODULE\s+([\/\.\s\w\(\*\)]+);`)
	reImport := regexp.MustCompile(`(?is)IMPORT\s+([^;]+?);`)
	reComment := regexp.MustCompile(`(?is)\(\*\*?.*?\*\)`)

	text = reComment.ReplaceAllString(text, "")

	moduleMatches := reModule.FindStringSubmatch(text)
	if moduleMatches != nil {
		mod.Name = moduleMatches[1]
	}

	importMatches := reImport.FindAllStringSubmatch(text, -1)
	var imports string
	var importList []string
	switch {
	case len(importMatches) == 0:
		imports = ""
	case len(importMatches[0]) > 0:
		imports = importMatches[0][1]
	default:
		imports = importMatches[0][0]
	}
	importList = regexp.MustCompile(`[,]+`).Split(imports, -1)
	for _, imp := range importList {
		imp = strings.TrimSpace(imp)
		impRight := strings.Split(imp, ":=")
		if len(impRight) > 1 {
			imp = strings.TrimSpace(impRight[1])
		} else {
			imp = strings.TrimSpace(impRight[0])
		}
		if imp != "" {
			mod.AddImport(imp)
		}
	}
	return mod, nil
}
