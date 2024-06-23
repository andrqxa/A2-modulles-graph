package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	config "github.com/andrqxa/A2-modules-graph/configs"
	"github.com/andrqxa/A2-modules-graph/internal"
	"github.com/spf13/viper"
)

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

	modules.GenerateDOT("modules_graph.dot")
}

// Parse modules from the given directory
func parseModules(dir string) internal.Modules {
	modules := internal.NewModules()

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
func parseModuleFile(path string) (internal.Module, error) {
	file, err := os.Open(path)
	if err != nil {
		return internal.NewModule(""), err
	}
	defer file.Close()

	mod := internal.NewModule("")

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
