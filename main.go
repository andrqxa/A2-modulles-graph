package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type imports []string
type Modules map[string]imports

func main() {
	// dir := "/home/andrejjj/Projects/Aos/A2-oberon/source"
	dir := "/home/andrejjj/Projects/Aos/xlam/11"
	modules := parseModules(dir)

	moduleNames := make([]string, 0, len(modules))
	for name := range modules {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)

	Print(moduleNames, modules)

	generateDOT(modules, "modules_graph.dot")
}

func Print(moduleNames []string, modules map[string]imports) {
	for _, name := range moduleNames {
		moduleImports := modules[name]
		fmt.Printf("%s: \n", name)
		for _, imp := range moduleImports {
			fmt.Printf("\t%s\n", imp)
		}
		fmt.Println("===============================================")
	}
}

func parseModules(dir string) Modules {
	modules := make(Modules)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".Mod") || strings.HasSuffix(info.Name(), ".mod")) {
			moduleName, moduleImports, err := parseModuleFile(path)
			if moduleName == "" {
				return fmt.Errorf("error parsing module file %q: module name not found", path)
			}

			if err == nil {
				modules[moduleName] = moduleImports
				for _, imp := range moduleImports {
					if _, ok := modules[imp]; !ok {
						modules[imp] = nil
					}
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

func parseModuleFile(path string) (string, imports, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	var moduleName string
	var moduleImports imports

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
	// reModule := regexp.MustCompile(`(?is)MODULE\s+([\/\.\s\w\(\*\)]+);`)
	reModule := regexp.MustCompile(`(?is)MODULE\s+([^;]+?);`)
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
	return moduleName, moduleImports, nil
}

func generateDOT(modules map[string]imports, filename string) {
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

	for name := range modules {
		if !visited[name] {
			rank := calculateRank(name, modules, visited)
			ranks[rank] = append(ranks[rank], name)
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

	for name, moduleImports := range modules {
		for _, imp := range moduleImports {
			if _, exists := modules[imp]; exists {
				f.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", name, imp))
			}
		}
	}

	f.WriteString("}\n")
}

func calculateRank(name string, modules map[string]imports, visited map[string]bool) int {
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
