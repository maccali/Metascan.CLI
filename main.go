package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"metascan/pkg"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dirPath := flag.String("dir", ".", "Caminho para o diretório a ser processado")
	outputName := flag.String("output", "file_metadata_report", "Nome base do arquivo de saída (sem extensão)")
	recursive := flag.Bool("r", false, "Processar subdiretórios recursivamente")
	extFilter := flag.String("ext", "", "Filtrar apenas arquivos com esta extensão (ex: .jpg)")
	format := flag.String("format", "csv", "Formato de saída: csv ou json")

	flag.Parse()

	if *dirPath == "" {
		log.Println("Erro: O caminho do diretório é obrigatório.")
		flag.Usage()
		os.Exit(1)
	}

	dirInfo, err := os.Stat(*dirPath)
	if os.IsNotExist(err) {
		log.Fatalf("Erro: Diretório não encontrado em '%s'", *dirPath)
	}
	if err != nil {
		log.Fatalf("Erro ao obter informações do diretório '%s': %v", *dirPath, err)
	}
	if !dirInfo.IsDir() {
		log.Fatalf("Erro: O caminho '%s' não é um diretório.", *dirPath)
	}

	var results []*pkg.FileInfoData

	log.Printf("Processando diretório: %s (Recursivo: %t)", *dirPath, *recursive)
	var filesProcessed, filesAttempted, filesWithErrors int

	walkFunc := func(path string, d fs.DirEntry, errWalk error) error {
		if errWalk != nil {
			log.Printf("Erro ao acessar '%s': %v (pulando)", path, errWalk)
			return nil
		}

		if d.IsDir() {
			if path == *dirPath {
				return nil
			}
			if !*recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Filtra extensão, se especificado
		if *extFilter != "" && !strings.HasSuffix(strings.ToLower(path), strings.ToLower(*extFilter)) {
			return nil
		}

		filesAttempted++

		data, errProc := pkg.ProcessFile(path)
		if errProc != nil {
			log.Printf("Erro ao processar arquivo '%s': %v", path, errProc)
			filesWithErrors++
			return nil
		}

		if data != nil {
			results = append(results, data)
			filesProcessed++
		}

		return nil
	}

	if err := filepath.WalkDir(*dirPath, walkFunc); err != nil {
		log.Printf("Erro final ao percorrer o diretório '%s': %v", *dirPath, err)
	}

	// --- Output ---
	switch strings.ToLower(*format) {
	case "json":
		jsonFilePath := *outputName + ".json"
		jsonFile, err := os.Create(jsonFilePath)
		if err != nil {
			log.Fatalf("Erro ao criar arquivo JSON '%s': %v", jsonFilePath, err)
		}
		defer jsonFile.Close()

		encoder := json.NewEncoder(jsonFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			log.Fatalf("Erro ao escrever JSON: %v", err)
		}
		log.Printf("Relatório JSON gerado em: %s", jsonFilePath)

	default: // CSV como padrão
		csvFilePath := *outputName + ".csv"
		csvFile, err := os.Create(csvFilePath)
		if err != nil {
			log.Fatalf("Erro ao criar arquivo CSV '%s': %v", csvFilePath, err)
		}
		defer csvFile.Close()

		csvWriter := csv.NewWriter(csvFile)
		defer csvWriter.Flush()

		if err := pkg.WriteCSVHeader(csvWriter); err != nil {
			log.Fatalf("Erro ao escrever cabeçalho no CSV: %v", err)
		}

		for _, data := range results {
			if err := pkg.WriteCSVRecord(csvWriter, data); err != nil {
				log.Printf("Erro ao escrever registro no CSV: %v", err)
			}
		}

		log.Printf("Relatório CSV gerado em: %s", csvFilePath)
	}

	log.Printf("Processamento concluído. Arquivos tentados: %d. Arquivos incluídos: %d. Erros: %d.", filesAttempted, filesProcessed, filesWithErrors)
}
