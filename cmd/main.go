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

	"golang.org/x/net/proxy"
)

// Configurações do cliente HTTP para Tor
type TorConfig struct {
	proxyAddress string
	timeout      time.Duration
	maxRetries   int
	forceTor     bool // Nova opção para forçar uso do Tor
}

func main() {
	// Configurações de entrada via CLI
	startEdition := flag.Int("start", 0, "Número inicial da edição (obrigatório)")
	endEdition := flag.Int("end", 0, "Número final da edição (opcional, padrão: igual ao inicial)")
	csvPath := flag.String("csv", "csvdje2024-2025.csv", "Caminho para o arquivo CSV")
	logPath := flag.String("log", "download_log.txt", "Caminho para o arquivo de log")
	randomMin := flag.Int("min", 3, "Tempo mínimo de pausa entre downloads (em segundos)")
	randomMax := flag.Int("max", 7, "Tempo máximo de pausa entre downloads (em segundos)")

	// Novas flags para configuração do Tor
	proxyAddr := flag.String("proxy", "127.0.0.1:9050", "Endereço do proxy SOCKS5 do Tor")
	timeout := flag.Int("timeout", 60, "Timeout em segundos para requisições HTTP")
	maxRetries := flag.Int("retries", 3, "Número máximo de tentativas de download")
	forceTor := flag.Bool("tor", false, "Forçar uso do Tor (padrão: falso, tenta conexão direta primeiro)")

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

	// Configurando cliente Tor
	torConfig := TorConfig{
		proxyAddress: *proxyAddr,
		timeout:      time.Duration(*timeout) * time.Second,
		maxRetries:   *maxRetries,
		forceTor:     *forceTor,
	}

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
		if err := downloadFile(url, filePath, torConfig); err != nil {
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

// Função para criar um cliente HTTP configurado para usar o Tor
func createTorClient(config TorConfig) (*http.Client, error) {
	// Criar dialer SOCKS5
	dialer, err := proxy.SOCKS5("tcp", config.proxyAddress, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar dialer SOCKS5: %v", err)
	}

	// Criar transport personalizado com configurações otimizadas
	transport := &http.Transport{
		Dial:               dialer.Dial,
		DisableKeepAlives:  true,
		DisableCompression: true,
		// Aumentar timeouts do transport
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}

	// Criar cliente HTTP com timeout
	client := &http.Client{
		Transport: transport,
		Timeout:   config.timeout,
		// Não seguir redirecionamentos automaticamente
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return client, nil
}

// Função para criar um cliente HTTP direto (sem Tor)
func createDirectClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		DisableKeepAlives:     true,
		DisableCompression:    true,
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// Função para tentar download com um cliente específico
func tryDownload(client *http.Client, url, filePath string, attempt int) error {
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

// Função para baixar um arquivo com configuração do Tor
func downloadFile(url, filePath string, config TorConfig) error {
	var client *http.Client
	var err error

	// Se não forçar Tor, tenta primeiro conexão direta
	if !config.forceTor {
		fmt.Println("Tentando conexão direta...")
		client = createDirectClient(config.timeout)
		// Tenta download direto
		if err := tryDownload(client, url, filePath, 1); err == nil {
			return nil // Download direto bem sucedido
		} else {
			fmt.Printf("Conexão direta falhou: %v\nTentando via Tor...\n", err)
		}
	}

	// Se conexão direta falhou ou forceTor está ativo, usa Tor
	client, err = createTorClient(config)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente Tor: %v", err)
	}

	var lastErr error
	for attempt := 0; attempt < config.maxRetries; attempt++ {
		if attempt > 0 {
			// Backoff exponencial entre tentativas
			backoff := time.Duration(attempt*attempt) * time.Second
			fmt.Printf("Tentativa %d de %d, aguardando %v...\n", attempt+1, config.maxRetries, backoff)
			time.Sleep(backoff)
		}

		if err := tryDownload(client, url, filePath, attempt+1); err != nil {
			lastErr = err
			fmt.Printf("Erro na tentativa %d: %v\n", attempt+1, err)
			continue
		}

		return nil // Download bem sucedido
	}

	return fmt.Errorf("todas as tentativas falharam. Último erro: %v", lastErr)
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
	fmt.Println("  ./downloaddje -start=<edição inicial> -end=<edição final> [opções]")
	fmt.Println("\nParâmetros obrigatórios:")
	fmt.Println("  -start=N                 Número da edição inicial")
	fmt.Println("  -end=N                  Número da edição final (opcional, padrão: igual ao inicial)")
	fmt.Println("\nParâmetros opcionais:")
	fmt.Println("  -csv=arquivo.csv        Arquivo CSV com mapeamento (padrão: csvdje2024-2025.csv)")
	fmt.Println("  -log=arquivo.txt        Arquivo de log (padrão: download_log.txt)")
	fmt.Println("  -min=N                  Pausa mínima entre downloads em segundos (padrão: 3)")
	fmt.Println("  -max=N                  Pausa máxima entre downloads em segundos (padrão: 7)")
	fmt.Println("\nOpções do Tor:")
	fmt.Println("  -tor                    Forçar uso do Tor (padrão: falso)")
	fmt.Println("  -proxy=host:porta       Endereço do proxy SOCKS5 (padrão: 127.0.0.1:9050)")
	fmt.Println("  -timeout=N              Timeout das requisições em segundos (padrão: 60)")
	fmt.Println("  -retries=N              Número máximo de tentativas (padrão: 3)")
	fmt.Println("\nExemplos:")
	fmt.Println("1. Download direto (tenta conexão normal primeiro):")
	fmt.Println("  ./downloaddje -start=1 -end=10")
	fmt.Println("\n2. Download forçando uso do Tor (útil no Whonix):")
	fmt.Println("  ./downloaddje -start=1 -end=10 -tor -timeout=180 -retries=5")
	fmt.Println("\n3. Download com configurações personalizadas:")
	fmt.Println("  ./downloaddje -start=1 -end=10 -csv=csvdje2024-2025.csv -min=10 -max=20 \\")
	fmt.Println("              -tor -timeout=300 -retries=7 -proxy=127.0.0.1:9050")
	fmt.Println("===========================================")
}
