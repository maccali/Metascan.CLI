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
	"strconv"
	"strings"
	"time"
)

type Manifest struct {
	OutputFile       string         `json:"output_file"`
	OutputFormat     string         `json:"output_format"`
	TotalAttempted   int            `json:"total_attempted"`
	TotalProcessed   int            `json:"total_processed"`
	TotalWithErrors  int            `json:"total_with_errors"`
	OutputFileHashes pkg.FileHashes `json:"output_file_hashes"`
	GeneratedAt      string         `json:"generated_at"`
}

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

	outputFilePath := *outputName + "." + strings.ToLower(*format)

	switch strings.ToLower(*format) {
	case "json":
		jsonFile, err := os.Create(outputFilePath)
		if err != nil {
			log.Fatalf("Erro ao criar arquivo JSON '%s': %v", outputFilePath, err)
		}
		defer jsonFile.Close()

		encoder := json.NewEncoder(jsonFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(results); err != nil {
			log.Fatalf("Erro ao escrever JSON: %v", err)
		}
		log.Printf("Relatório JSON gerado em: %s", outputFilePath)

	default: // CSV como padrão
		csvFile, err := os.Create(outputFilePath)
		if err != nil {
			log.Fatalf("Erro ao criar arquivo CSV '%s': %v", outputFilePath, err)
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
		log.Printf("Relatório CSV gerado em: %s", outputFilePath)
	}

	// --- Gerar manifesto ---
	manifest := Manifest{
		OutputFile:      outputFilePath,
		OutputFormat:    *format,
		TotalAttempted:  filesAttempted,
		TotalProcessed:  filesProcessed,
		TotalWithErrors: filesWithErrors,
		GeneratedAt:     time.Now().Format(time.RFC3339),
	}

	hashes, err := pkg.CalcFileHashes(outputFilePath)
	if err != nil {
		log.Printf("Aviso: não foi possível calcular hashes do arquivo de saída: %v", err)
	} else {
		manifest.OutputFileHashes = hashes
	}

	manifestFilePath := *outputName + "-manifest." + strings.ToLower(*format)
	manifestFile, err := os.Create(manifestFilePath)
	if err != nil {
		log.Fatalf("Erro ao criar manifesto: %v", err)
	}
	defer manifestFile.Close()

	switch strings.ToLower(*format) {
	case "json":
		encoder := json.NewEncoder(manifestFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(manifest); err != nil {
			log.Fatalf("Erro ao escrever manifesto JSON: %v", err)
		}
	case "csv":
		manifestWriter := csv.NewWriter(manifestFile)
		defer manifestWriter.Flush()

		manifestWriter.Write([]string{
			"OutputFile", "OutputFormat", "TotalAttempted", "TotalProcessed", "TotalWithErrors",
			"MD5", "SHA1", "SHA256", "GeneratedAt",
		})

		manifestWriter.Write([]string{
			manifest.OutputFile,
			manifest.OutputFormat,
			strconv.Itoa(manifest.TotalAttempted),
			strconv.Itoa(manifest.TotalProcessed),
			strconv.Itoa(manifest.TotalWithErrors),
			manifest.OutputFileHashes.MD5,
			manifest.OutputFileHashes.SHA1,
			manifest.OutputFileHashes.SHA256,
			manifest.GeneratedAt,
		})
	}

	log.Printf("Manifesto gerado em: %s", manifestFilePath)

	log.Printf("Processamento concluído. Arquivos tentados: %d. Arquivos incluídos: %d. Erros: %d.",
		filesAttempted, filesProcessed, filesWithErrors)
}
