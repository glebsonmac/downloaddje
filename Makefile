# Variáveis
BINARY_NAME=downloaddje
BUILD_DIR=bin
SOURCE_DIR=cmd

# Limpa e cria diretório de build
.PHONY: prepare
prepare:
	@echo "Preparando diretório de build..."
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

# Build para todas as plataformas
.PHONY: all
all: prepare windows-386 windows-amd64 linux-386 linux-amd64

# Windows 32-bit
.PHONY: windows-386
windows-386:
	@echo "Building para Windows 32-bit..."
	@GOOS=windows GOARCH=386 go build -o $(BUILD_DIR)/$(BINARY_NAME)_windows_386.exe $(SOURCE_DIR)/main.go

# Windows 64-bit
.PHONY: windows-amd64
windows-amd64:
	@echo "Building para Windows 64-bit..."
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(SOURCE_DIR)/main.go

# Linux 32-bit
.PHONY: linux-386
linux-386:
	@echo "Building para Linux 32-bit..."
	@GOOS=linux GOARCH=386 go build -o $(BUILD_DIR)/$(BINARY_NAME)_linux_386 $(SOURCE_DIR)/main.go

# Linux 64-bit
.PHONY: linux-amd64
linux-amd64:
	@echo "Building para Linux 64-bit..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(SOURCE_DIR)/main.go