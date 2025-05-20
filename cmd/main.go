package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Configurações de entrada via CLI
	startEdition := flag.Int("start", 0, "Número inicial da edição (obrigatório)")
	endEdition := flag.Int("end", 0, "Número final da edição (opcional, padrão: igual ao inicial)")
	csvPath := flag.String("csv", "csvdje2024-2025.csv", "Caminho para o arquivo CSV")
	logPath := flag.String("log", "download_log.txt", "Caminho para o arquivo de log")
	randomMin := flag.Int("min", 3, "Tempo mínimo de pausa entre downloads (em segundos)")
	randomMax := flag.Int("max", 7, "Tempo máximo de pausa entre downloads (em segundos)")
	flag.Parse()

	// Exibir cabeçalho
	printHeader()

	if *startEdition == 0 {
		fmt.Println("Erro: Você deve especificar pelo menos a edição inicial usando o parâmetro -start.")
		return
	}
	if *endEdition == 0 {
		*endEdition = *startEdition
	}
	if *randomMin > *randomMax {
		fmt.Println("Erro: O tempo mínimo não pode ser maior que o tempo máximo.")
		return
	}

	// Configurações iniciais
	baseURL := "https://www2.tjpe.jus.br/dje/DownloadServlet?dj=DJ%d_%d-ASSINADO.PDF&statusDoDiario=ASSINADO"
	year := 2024

	// Ler mapeamento do CSV
	dateMap, err := loadDateMapping(*csvPath)
	if err != nil {
		fmt.Printf("Erro ao carregar o CSV: %v\n", err)
		return
	}

	// Criar pasta de download
	downloadDir := "downloadDJE"
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		fmt.Printf("Erro ao criar pasta de download: %v\n", err)
		return
	}

	// Abrir arquivo de log
	logFile, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Erro ao abrir arquivo de log: %v\n", err)
		return
	}
	defer logFile.Close()

	// Simulação de download com pausas aleatórias
	for edition := *startEdition; edition <= *endEdition; edition++ {
		key := fmt.Sprintf("%d", edition)
		date, exists := dateMap[key]
		if !exists {
			fmt.Printf("Data não encontrada para a edição %d\n", edition)
			continue
		}

		url := fmt.Sprintf(baseURL, edition, year)

		// Substituir barras na data por hífens para evitar problemas no sistema de arquivos
		safeDate := strings.ReplaceAll(date, "/", "-")
		fileName := fmt.Sprintf("DJE_%d_%s.pdf", edition, safeDate)
		filePath := filepath.Join(downloadDir, fileName)

		fmt.Printf("Baixando: %s\n", url)
		if err := downloadFile(url, filePath); err != nil {
			fmt.Printf("Erro ao baixar arquivo: %v\n", err)
			continue
		}

		// Registrar no log
		logEntry := fmt.Sprintf("Arquivo baixado: %s - %s\n", fileName, time.Now().Format("02/01/2006 15:04:05"))
		if _, err := logFile.WriteString(logEntry); err != nil {
			fmt.Printf("Erro ao escrever no log: %v\n", err)
		}

		// Pausa aleatória para simular acesso humano
		pause := rand.Intn(*randomMax-*randomMin+1) + *randomMin
		fmt.Printf("Pausando por %d segundos...\n", pause)
		time.Sleep(time.Duration(pause) * time.Second)
	}

	fmt.Println("Download concluído.")
}

// Função para baixar um arquivo
func downloadFile(url, filePath string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Simular cabeçalho de navegador
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status inesperado: %s", resp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// Função para carregar o mapeamento do CSV
func loadDateMapping(csvPath string) (map[string]string, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	dateMap := make(map[string]string)

	// Ignorar cabeçalho
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Erro ao ler linha do CSV: %v\n", err)
			continue
		}

		// Verificar se a linha tem pelo menos 2 colunas
		if len(record) < 2 {
			fmt.Println("Linha inválida no CSV, ignorando:", record)
			continue
		}

		// Remover espaços extras das colunas
		edition := strings.TrimSpace(record[0])
		date := strings.TrimSpace(record[1])

		// Adicionar ao mapa
		dateMap[edition] = date
	}

	return dateMap, nil
}

// Função para exibir o cabeçalho do CLI
func printHeader() {
	fmt.Println("===========================================")
	fmt.Println("      TJPE Downloader - Diário Oficial     ")
	fmt.Println("===========================================")
	fmt.Println("Desenvolvido para baixar edições do TJPE.")
	fmt.Println("===========================================")
	fmt.Println("Como usar:")
	fmt.Println("  ./goprogram or goprogram.exe -start=<edição inicial> -end=<edição final> -csv=<caminho do CSV> -log=<arquivo de log> -min=<pausa mínima> -max=<pausa máxima>")
	fmt.Println("Exemplo:")
	fmt.Println("  ./goprogram or goprogram.exe -start=1 -end=10 -csv=csvdje2024-2025.csv -log=download_log.txt -min=5 -max=10")
	fmt.Println("===========================================")
}
