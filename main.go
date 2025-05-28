package main

import (
	"encoding/csv"
	"flag"
	"io/fs"
	"log"
	"metascan/pkg"
	"os"
	"path/filepath"
)

func main() {
	dirPath := flag.String("dir", ".", "Caminho para o diretório a ser processado")
	outputCSVName := flag.String("output", "file_metadata_report.csv", "Nome do arquivo CSV de saída")
	recursive := flag.Bool("r", false, "Processar subdiretórios recursivamente")
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

	csvFilePath := *outputCSVName
	csvFile, err := os.Create(csvFilePath)
	if err != nil {
		log.Fatalf("Erro ao criar arquivo CSV '%s': %v", csvFilePath, err)
	}
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Escreve cabeçalho usando a função separada
	if err := pkg.WriteCSVHeader(csvWriter); err != nil {
		log.Fatalf("Erro ao escrever cabeçalho no CSV: %v", err)
	}

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

		filesAttempted++

		data, errProc := pkg.ProcessFile(path)
		if errProc != nil {
			log.Printf("Erro ao processar arquivo '%s': %v", path, errProc)
			filesWithErrors++
			return nil
		}

		if data != nil {
			if err := pkg.WriteCSVRecord(csvWriter, data); err != nil {
				log.Printf("Erro ao escrever registro para '%s' no CSV: %v", path, err)
				filesWithErrors++
			} else {
				filesProcessed++
			}
		}

		return nil
	}

	if err := filepath.WalkDir(*dirPath, walkFunc); err != nil {
		log.Printf("Erro final ao percorrer o diretório '%s': %v", *dirPath, err)
	}

	log.Printf("Processamento concluído. Arquivos tentados: %d. Arquivos incluídos no CSV: %d. Erros críticos: %d.", filesAttempted, filesProcessed, filesWithErrors)
	log.Printf("Relatório CSV gerado em: %s", csvFilePath)
}
