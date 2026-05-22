# WhatsApp Bot Pro - Instagram Downloader

Este projeto é um bot de WhatsApp profissional desenvolvido em Go, focado em detectar links do Instagram e realizar o download automático de vídeos para envio direto no chat.

## 🚀 Funcionalidades

- **Download Automático**: Detecta links de Reels e Posts do Instagram e envia o vídeo diretamente.
- **Feedback em Tempo Real**: Notifica o usuário quando o download começa e exibe mensagens de erro claras.
- **Estrutura Modular**: Organizado seguindo padrões profissionais de Go (`cmd/` e `internal/`), facilitando a manutenção e expansão.
- **Persistência de Sessão**: Utiliza SQLite para manter a conexão ativa sem precisar escanear o QR Code a cada reinício.
- **Assíncrono**: Processamento de mensagens em goroutines para alta performance.

## 📁 Estrutura do Projeto

A arquitetura do projeto foi desenhada para ser escalável:

```text
├── cmd/
│   └── bot/             # Ponto de entrada (main.go)
├── internal/
│   ├── whatsapp/        # Biblioteca interna para gestão do cliente WhatsApp
│   ├── handlers/        # Processamento de lógica de mensagens e eventos
│   └── instagram/       # Lógica de negócio específica para download do Instagram
├── go.mod               # Dependências do projeto
└── .gitignore           # Proteção contra arquivos binários e bancos de dados
```

## 🛠️ Como Executar

### Pré-requisitos
- [Go](https://golang.org/dl/) (v1.25+)
- GCC (necessário para o SQLite)

### Instalação e Execução
1. Clone o repositório.
2. Instale as dependências:
   ```bash
   go mod tidy
   ```
3. Execute o bot:
   ```bash
   go run cmd/bot/main.go
   ```
4. Escaneie o QR Code exibido no terminal com o seu WhatsApp.

## 📦 Tecnologias Utilizadas

- [Whatsmeow](https://go.mau.fi/whatsmeow): Biblioteca de WhatsApp para Go.
- [SQLite](https://modernc.org/sqlite): Banco de dados leve para sessões.
- [QRTerminal](https://github.com/mdp/qrterminal): Exibição de QR Code no console.

## 📝 Licença

Este projeto está sob a licença MIT.
