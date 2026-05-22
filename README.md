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

### Opção 1: Nativo (Go)
1. Instale as dependências: `go mod tidy`
2. Execute: `go run cmd/bot/main.go`

### Opção 2: Docker Compose (Recomendado)
Esta é a forma mais fácil e organizada de rodar o bot 24/7.

1. Crie um arquivo `docker-compose.yml`:
```yaml
services:
  wa-bot:
    image: ghcr.io/robertocjunior/wa-bot-teste:latest
    container_name: wa-bot-inst
    restart: unless-stopped
    volumes:
      - ./data:/app/data
    environment:
      - LOG_LEVEL=info
      - TZ=America/Sao_Paulo
```

2. Suba o serviço:
```bash
docker compose up -d
```

## 🛡️ Resiliência (24/7)
- **Panic Recovery**: O bot recupera-se automaticamente de falhas inesperadas no processamento de mensagens.
- **Structured Logging**: Logs profissionais com `zerolog`.
- **Auto-Reconnect**: Gestão inteligente de conexão com o WhatsApp.
- **Timeouts**: Proteção contra APIs externas lentas.

## 📦 Tecnologias Utilizadas
...
- [Whatsmeow](https://go.mau.fi/whatsmeow): Biblioteca de WhatsApp para Go.
- [SQLite](https://modernc.org/sqlite): Banco de dados leve para sessões.
- [QRTerminal](https://github.com/mdp/qrterminal): Exibição de QR Code no console.

## 📝 Licença

Este projeto está sob a licença MIT.
