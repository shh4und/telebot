# Telegram Bot (Go)

Bot multifuncional para Telegram desenvolvido em Go, integrando Inteligência Artificial (Ollama), leitor de feeds RSS de notícias, monitor de cotações financeiras, gerador de stickers personalizados e publicação automática de respostas formatadas no Telegraph.

---

## Funcionalidades

### Inteligência Artificial & Telegraph
- **Integração com Ollama:** Consultas a modelos de IA locais ou remotos (via rede ou Tailscale).
- **Seleção de Modelos (`/modelo`):** Troca dinâmica do modelo ativo de IA por chat/usuário.
- **Renderização no Telegraph:** Respostas longas ou ricas em Markdown são convertidas em páginas do [Telegraph](https://telegra.ph) com suporte a Instant View nativo.

### Notícias & Feeds RSS (`/noticias`)
- Agregação em tempo real de notícias categorizadas:
  - 🇧🇷 **Brasil:** CNN Brasil, G1
  - 🌍 **Mundo:** BBC News Brasil
  - 📈 **Economia:** InfoMoney, G1 Economia
  - 💻 **Tecnologia:** Canaltech
- Navegação interativa com botões inline e paginação.

### Cotações Financeiras & Criptomoedas (`/moedas`)
- Cotações em tempo real de moedas fiduciárias (USD, EUR, GBP, JPY, CAD, etc.).
- Cotações de criptomoedas (BTC, ETH, SOL, etc.).
- Exibição de variação percentual diária, máxima, mínima e data/hora da atualização.

### Stickers & Mídia
- **`/fig`:** Converte imagens enviadas ou respondidas para o formato WebP (512x512).
- **`/addfig`:** Adiciona a imagem convertida diretamente a um pacote de stickers.
- **`/gif` & `/addgif`:** Processamento e conversão de animações/GIFs para stickers.

---

## Comandos Disponíveis

| Comando | Descrição |
| :--- | :--- |
| `/ping` | Verifica se o bot está online |
| `/ajuda` | Exibe informações de ajuda e comandos |
| `/pergunta <texto>` | Faz uma pergunta à IA (Ollama) |
| `/modelo` | Abre menu interativo para escolher o modelo de IA |
| `/noticias` | Exibe as principais notícias e categorias |
| `/moedas` | Exibe cotação de moedas e criptomoedas |
| `/fig` | Converte imagem respondida em sticker |
| `/addfig` | Adiciona figurinha ao pacote |
| `/gif` | Converte GIF respondido em sticker |
| `/addgif` | Adiciona GIF ao pacote |

---

## Estrutura do Projeto

```text
telegram-bot/
├── .github/
│   └── workflows/
│       ├── ci.yml              # Validação de build CGO e testes em PRs
│       └── deploy.yml          # Build e deploy contínuo na VM via SSH/SCP
├── cmd/
│   └── bot/
│       └── main.go             # Ponto de entrada (Loop de Long Polling e Dispatcher)
├── internal/
│   ├── ai/                     # Cliente Ollama, schemas e listagem de modelos
│   ├── bot/                    # Handlers, comandos, dispatcher e sessões
│   ├── config/                 # Carregamento de variáveis de ambiente
│   ├── news/                   # Parser RSS, agregador de notícias e categorias
│   ├── quotes/                 # Integração com APIs financeiras e cotações
│   ├── stickers/               # Processamento de imagem WebP e API de stickers
│   └── telegraph/              # Parser Markdown -> HTML -> DOM do Telegraph
├── deploy/
│   └── telebot.service         # Unit systemd para execução na VM
├── .env.example                # Template de variáveis de ambiente
├── .gitignore
├── go.mod
└── go.sum
```

---

## Configuração do Ambiente

### 1. Pré-requisitos
- **Go 1.23+**
- **Dependências CGO** (necessárias para processamento WebP com `chai2010/webp`):
  ```bash
  # Debian / Ubuntu
  sudo apt-get update && sudo apt-get install -y build-essential libwebp-dev
  
  # macOS (Homebrew)
  brew install webp
  ```
- **Ollama** em execução local ou acessível via rede.

### 2. Variáveis de Ambiente
Copie o arquivo `.env.example` para `.env`:

```bash
cp .env.example .env
```

Defina os valores das variáveis:

```env
# Token do bot obtido via @BotFather
BOT_TK=123456789:ABCdefGHIjklMNOpqrsTUVwxyz

# ID de usuário do Telegram (para restrições ou administração)
USER_ID=123456789

# Endpoint da API do Ollama
API_HOST=http://localhost:11434
```

---

## Desenvolvimento Local

Para desenvolvimento com recarregamento automático (live reload):

1. Instale o [Air](https://github.com/air-verse/air):
   ```bash
   go install github.com/air-verse/air@latest
   ```

2. Execute o projeto:
   ```bash
   air
   ```

> **Nota:** Para evitar conflito de polling (`Error 409: Conflict`) com a instância de produção na VM, crie um bot de desenvolvimento no @BotFather (ex: `@meubot_dev_bot`) e use o token correspondente no `.env` local.

---

## Deploy em Produção (GCP VM + systemd + CI/CD)

O repositório possui pipeline automatizado via GitHub Actions configurado em `.github/workflows/deploy.yml`:

1. **Build:** O binário Linux `amd64` é compilado no ambiente do GitHub Actions com `CGO_ENABLED=1`.
2. **Transferência (SCP):** O executável é transferido de forma segura via SSH para a VM.
3. **Substituição Atômica & Restart:** O binário é movido para `/opt/telebot/telebot` e o serviço `telebot.service` é reiniciado pelo `systemd`.

### Secrets do GitHub Actions (Environment: `telebot`):
- `VM_HOST`: Endereço IP externo da VM.
- `VM_USER`: Usuário SSH na VM (ex: `dnxx`).
- `SSH_PRIVATE_KEY`: Chave privada SSH autorizada.

---

## Testes

Para rodar os testes unitários do projeto:

```bash
go test -v ./...
```

---

## Licença

Distribuído sob a licença MIT. Consulte `LICENSE` para mais detalhes.
