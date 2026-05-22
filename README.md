# WhatsApp Bot em Go

Este é um bot simples de WhatsApp construído em Go usando a biblioteca `whatsmeow`. O bot exibe um QR code no terminal para autenticação e responde com "oi" a qualquer mensagem recebida.

## Funcionalidades

- Exibição de QR Code no terminal para login via WhatsApp Web.
- Persistência de sessão usando SQLite.
- Resposta automática inteligente:
    - Responde com `reels: <link>` ao detectar links de Instagram (Reels ou Posts), extraindo a URL inclusive de metadados de prévia.
    - Suporta variações como `reel/`, `reels/` e `p/`, de forma insensível a maiúsculas.
    - Responde com "oi" para outras mensagens de texto.
- Animação de "digitando" (`composing`) para uma interação mais humana.
- Processamento assíncrono de mensagens para maior agilidade.
- Registro de logs detalhados para depuração.

## Pré-requisitos

- [Go](https://golang.org/dl/) (versão 1.25 ou superior recomendada).
- GCC (necessário para o driver SQLite `go-sqlite3`).

## Como Executar

1. Clone o repositório (ou acesse o diretório do projeto).
2. Instale as dependências:
   ```bash
   go mod tidy
   ```
3. Execute a aplicação:
   ```bash
   go run main.go
   ```
4. Escaneie o QR code que aparecerá no terminal com o seu aplicativo do WhatsApp (Configurações > Aparelhos conectados > Conectar um aparelho).

## Estrutura do Projeto

- `main.go`: Contém a lógica principal da aplicação, incluindo conexão, tratamento de eventos e resposta automática.
- `go.mod` e `go.sum`: Gerenciamento de dependências.
- `examplestore.db`: Arquivo SQLite gerado automaticamente para armazenar os dados da sessão.

## Boas Práticas

Este projeto segue princípios de desenvolvimento ágil:
- **Simplicidade**: Foca na funcionalidade principal solicitada.
- **Documentação**: Código comentado e README claro.
- **Responsabilidade Única**: Funções modulares para tratamento de eventos.
- **Tratamento de Erros**: Verificação básica de erros em operações críticas.

## Licença

Este projeto está licenciado sob a licença MIT.
