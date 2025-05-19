# TJPE Downloader - Diário Oficial

Este projeto é um utilitário CLI para baixar edições do Diário Oficial do Tribunal de Justiça de Pernambuco (TJPE) com base em um arquivo CSV que contém o mapeamento de edições e datas.

## Funcionalidades

- Baixa edições do Diário Oficial do TJPE.
- Simula acesso humano com pausas aleatórias configuráveis.
- Registra os downloads em um arquivo de log.
- Permite especificar o intervalo de edições a serem baixadas.

## Requisitos

- [Go](https://golang.org/) instalado na máquina.
- Arquivo CSV com o mapeamento de edições e datas (exemplo: `csvdje2024-2025.csv`).

## Como usar

### 1. Clonar o repositório

```bash
git clone <url-do-repositorio>
cd tjedownload
```

### 2. Estrutura do CSV

O arquivo CSV deve ter o seguinte formato:

```csv
Edição,Data
1,02/01/2024
2,03/01/2024
3,04/01/2024
...
```

### 3. Executar o programa

Compile e execute o programa com os parâmetros desejados:

```bash
go run cmd/main.go -start=<edição inicial> -end=<edição final> -csv=<caminho do CSV> -log=<arquivo de log> -min=<pausa mínima> -max=<pausa máxima>
```

#### Parâmetros

- `-start`: Número inicial da edição (obrigatório).
- `-end`: Número final da edição (opcional, padrão: igual ao inicial).
- `-csv`: Caminho para o arquivo CSV (opcional, padrão: `csvdje2024-2025.csv`).
- `-log`: Caminho para o arquivo de log (opcional, padrão: `download_log.txt`).
- `-min`: Tempo mínimo de pausa entre downloads (em segundos, padrão: 3).
- `-max`: Tempo máximo de pausa entre downloads (em segundos, padrão: 7).

#### Exemplo

```bash
go run cmd/main.go -start=1 -end=10 -csv=csvdje2024-2025.csv -log=download_log.txt -min=5 -max=10
```

### 4. Resultado

- Os arquivos baixados serão salvos na pasta `downloadDJE`.
- O log dos downloads será salvo no arquivo especificado (padrão: `download_log.txt`).

## Estrutura do Projeto

```
tjedownload/
├── cmd/
│   └── main.go        # Código principal do programa
├── csvdje2024-2025.csv # Arquivo CSV com o mapeamento de edições e datas
├── downloadDJE/       # Pasta onde os arquivos baixados serão salvos
└── README.md          # Documentação do projeto
```

## Observações

- Certifique-se de que o arquivo CSV está no formato correto e no local especificado.
- Ajuste os tempos de pausa (`-min` e `-max`) para evitar bloqueios de IP.

## Licença

Este projeto é de uso interno e não possui uma licença pública.