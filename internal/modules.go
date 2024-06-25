package internal

import (
	"fmt"
	"os"
	"sort"
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
	res := fmt.Sprintf("%s:\\n", m.Name)
	for _, imp := range m.Imports {
		res += fmt.Sprintf("\\t%s\\n", imp)
	}
	res += "===============================================\\n"
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
func (ms Modules) GenerateDOT(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating DOT file:", err)
		return
	}
	defer f.Close()

	f.WriteString("digraph G {\n")
	f.WriteString("\trankdir=LR;\n")
	f.WriteString("\tsize=\"10,10\";\n")
	f.WriteString("\tnode [shape=box, style=filled, color=lightblue];\n") // Добавлен стиль для узлов

	// Создание кластеров
	clusters := make(map[string][]Module)
	for _, mod := range ms {
		clusterName := getClusterName(mod.Name) // Получаем имя кластера на основе имени модуля
		clusters[clusterName] = append(clusters[clusterName], mod)
	}

	for clusterName, mods := range clusters {
		f.WriteString(fmt.Sprintf("\tsubgraph cluster_%s {\n", clusterName))
		f.WriteString(fmt.Sprintf("\t\tlabel = \"%s\";\n", clusterName))
		f.WriteString("\t\tstyle=filled;\n")
		f.WriteString("\t\tcolor=lightgrey;\n")
		for _, mod := range mods {
			f.WriteString(fmt.Sprintf("\t\t\"%s\";\n", mod.Name))
		}
		f.WriteString("\t}\n")
	}

	// Создание рёбер
	for _, mod := range ms {
		for _, imp := range mod.Imports {
			f.WriteString(fmt.Sprintf("\t\"%s\" -> \"%s\";\n", mod.Name, imp))
		}
	}

	f.WriteString("}\n")
}

// Вспомогательная функция для получения имени кластера
func getClusterName(moduleName string) string {
	// Пример: возвращаем первые буквы имени модуля в качестве имени кластера
	if len(moduleName) > 0 {
		return string(moduleName[0])
	}
	return "default"
}

// Вспомогательная функция для объединения строк
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, str := range strs {
		if i > 0 {
			result += sep
		}
		result += str
	}
	return result
}
