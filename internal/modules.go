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
func (ms Modules) GenerateDOT(filename string) {
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
