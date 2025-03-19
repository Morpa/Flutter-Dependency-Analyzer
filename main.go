package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Estruturas para análise do pubspec.yaml
type PubspecYaml struct {
	Dependencies    map[string]interface{} `yaml:"dependencies"`
	DevDependencies map[string]interface{} `yaml:"dev_dependencies"`
}

// Estruturas para análise do pubspec.lock
type PubspecLock struct {
	Packages map[string]Package `yaml:"packages"`
	Sdks     map[string]string  `yaml:"sdks"`
}

type Package struct {
	Dependency string `yaml:"dependency"`
	Source     string `yaml:"source"`
	Version    string `yaml:"version"`
}

func main() {
	fmt.Println("\033[34m🔍 Verificando dependências...\033[0m")

	// Ler o pubspec.yaml
	pubspecContent, err := os.ReadFile("pubspec.yaml")
	if err != nil {
		fmt.Println("\033[31m❌ Erro: pubspec.yaml não encontrado!\033[0m")
		os.Exit(1)
	}

	// Ler o pubspec.lock
	lockContent, err := os.ReadFile("pubspec.lock")
	if err != nil {
		fmt.Println("\033[31m❌ Erro: pubspec.lock não encontrado!\033[0m")
		fmt.Println("\033[33m⚠️  Execute 'flutter pub get' ou 'dart pub get' primeiro.\033[0m")
		os.Exit(1)
	}

	// Parse do YAML do pubspec.yaml
	var pubspec PubspecYaml
	err = yaml.Unmarshal(pubspecContent, &pubspec)
	if err != nil {
		fmt.Println("\033[31m❌ Erro ao processar pubspec.yaml.\033[0m")
		fmt.Printf("\033[31mDetalhes do erro: %s\033[0m\n", err)
		os.Exit(1)
	}

	// Parse do YAML do pubspec.lock
	var lock PubspecLock
	err = yaml.Unmarshal(lockContent, &lock)
	if err != nil {
		fmt.Println("\033[31m❌ Erro ao processar pubspec.lock.\033[0m")
		fmt.Printf("\033[31mDetalhes do erro: %s\033[0m\n", err)
		os.Exit(1)
	}

	// Identificar pacotes que usam `git`
	gitPackageRegex1 := regexp.MustCompile(`(?m)^\s{2,}(\w+):\n\s{4}git:`)
	gitPackageRegex2 := regexp.MustCompile(`(?m)^\s{2,}(\w+):\s*git:`)

	matches1 := gitPackageRegex1.FindAllStringSubmatch(string(pubspecContent), -1)
	matches2 := gitPackageRegex2.FindAllStringSubmatch(string(pubspecContent), -1)

	gitPackages := make(map[string]bool)
	
	// Adicionar pacotes do primeiro padrão
	for _, match := range matches1 {
		gitPackages[match[1]] = true
	}
	
	// Adicionar pacotes do segundo padrão
	for _, match := range matches2 {
		gitPackages[match[1]] = true
	}

	// Identificar todos os pacotes declarados
	directDependencies := make(map[string]bool)
	for pkg := range pubspec.Dependencies {
		directDependencies[pkg] = true
	}

	// Verificar quais dependências são realmente usadas
	usedPackages := findUsedPackages("lib")
	unusedPackages := make([]string, 0)

	// Verificar cada dependência se está sendo usada
	for pkg := range directDependencies {
		// Ignoramos pacotes especiais que podem não ter imports diretos
		if pkg == "flutter" || pkg == "flutter_test" || pkg == "flutter_localizations" || pkg == "cupertino_icons" {
			continue
		}
		
		if !usedPackages[pkg] {
			unusedPackages = append(unusedPackages, pkg)
		}
	}

	// Exibir informações sobre pacotes
	fmt.Println("\033[34m🔍 Analisando dependências...\033[0m")
	
	// Contar pacotes para feedback
	total := len(directDependencies)
	gitCount := 0
	hostingCount := 0
	
	// Verificar pacotes em pubspec.lock
	for pkgName, pkgInfo := range lock.Packages {
		// Ignorar pacotes que não são dependências diretas
		if !directDependencies[pkgName] {
			continue
		}
		
		// Contar tipos de pacotes
		if gitPackages[pkgName] {
			gitCount++
			fmt.Printf("\033[90m📌 %s: %s (Git)\033[0m\n", pkgName, pkgInfo.Version)
			continue
		}
		
		// Verificar se é uma dependência não utilizada
		isUnused := false
		for _, unused := range unusedPackages {
			if unused == pkgName {
				isUnused = true
				break
			}
		}
		
		if isUnused {
			fmt.Printf("\033[31m❌ %s: versão %s (não utilizada)\033[0m\n", pkgName, pkgInfo.Version)
		} else if pkgInfo.Source == "hosted" {
			hostingCount++
			fmt.Printf("\033[36m📦 %s: versão %s\033[0m\n", pkgName, pkgInfo.Version)
		}
	}
	
	// Exibir resumo
	fmt.Println("\033[34m📊 Resumo:\033[0m")
	fmt.Printf("\033[34m   Total de dependências: %d\033[0m\n", total)
	fmt.Printf("\033[34m   Dependências Git: %d\033[0m\n", gitCount)
	fmt.Printf("\033[34m   Dependências hosted: %d\033[0m\n", hostingCount)
	
	// Mostrar pacotes não utilizados
	if len(unusedPackages) > 0 {
		fmt.Printf("\033[31m   Pacotes possivelmente não utilizados: %d\033[0m\n", len(unusedPackages))
		fmt.Println("\033[33m⚠️  Os seguintes pacotes parecem não ser utilizados em importações diretas:\033[0m")
		for _, pkg := range unusedPackages {
			fmt.Printf("\033[33m   - %s\033[0m\n", pkg)
		}
		fmt.Println("\033[33m   Você pode considerá-los para remoção, mas verifique se não são utilizados indiretamente.\033[0m")
	} else {
		fmt.Println("\033[32m✅ Todas as dependências parecem estar sendo utilizadas!\033[0m")
	}
}

// Função para encontrar pacotes utilizados no código
func findUsedPackages(rootDir string) map[string]bool {
	usedPackages := make(map[string]bool)
	
	// Regex para encontrar importações
	importRegex := regexp.MustCompile(`import\s+['"]package:([^\/]+)`)
	
	// Percorrer diretórios de código recursivamente
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		// Só analisar arquivos .dart
		if !info.IsDir() && strings.HasSuffix(path, ".dart") {
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()
			
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				
				// Verificar se a linha contém uma importação
				matches := importRegex.FindStringSubmatch(line)
				if len(matches) > 1 {
					packageName := matches[1]
					usedPackages[packageName] = true
				}
			}
		}
		
		return nil
	})
	
	return usedPackages
}