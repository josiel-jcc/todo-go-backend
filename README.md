# Todo Go Backend API

API RESTful para gerenciamento de tarefas desenvolvida em Go, com autenticação JWT e sistema de atribuição de tarefas entre usuários.

## Conformidade (LGPD e segurança)

**Fonte canônica:** [`todo-frontend/docs/compliance/`](../todo-frontend/docs/compliance/) (exibida no app em `/privacidade` e `/termos`). Cópia espelhada em [`docs/compliance/`](docs/compliance/).

- [Política de Privacidade](docs/compliance/PRIVACY.md)
- [Termos de Uso](docs/compliance/TERMS.md)
- [Postura de Segurança](docs/compliance/SECURITY.md)
- [ROPA](docs/compliance/LGPD-ROPA.md)
- [Retenção de dados](docs/compliance/DATA-RETENTION.md)
- [Resposta a incidentes](docs/compliance/INCIDENT-RESPONSE.md)

**Produção:** defina `APP_ENV=production`, `JWT_SECRET` (≥ 32 caracteres) e `CORS_ALLOWED_ORIGINS` (sem `*`). Veja `env.example` e [SECURITY.md](docs/compliance/SECURITY.md).

## Funcionalidades

- ✅ Autenticação JWT (registro e login)
- ✅ CRUD completo de tarefas
- ✅ Tarefas por usuário (cada usuário vê apenas suas tarefas)
- ✅ Tipos de tarefas: casa, trabalho, lazer, saúde
- ✅ Tempo de realização (due_date) para cada tarefa
- ✅ Usuários podem criar tarefas para outros usuários
- ✅ Filtros por tipo e status de conclusão
- ✅ Sistema de Tags para categorizar tarefas
- ✅ Comentários em tarefas
- ✅ Notificações por Email, Telegram e Web Push (PWA)
- ✅ Grupos, convites e notificações in-app
- ✅ Exportação e exclusão de conta (LGPD)
- ✅ Health check endpoint
- ✅ CORS configurável
- ✅ Suporte a MySQL e SQLite
- ✅ Cloudflare Tunnel integrado (exposição segura da API)

## Tecnologias

- **Go 1.21+**
- **Gin** - Framework web
- **GORM** - ORM para banco de dados
- **SQLite/MySQL** - Banco de dados (suporte a ambos)
- **JWT-Go** - Autenticação JWT
- **Bcrypt** - Hash de senhas
- **Swagger/OpenAPI** - Documentação da API
- **Docker & Docker Compose** - Containerização
- **GitHub Actions** - CI/CD Pipeline
- **Cloudflare Tunnel** - Exposição segura da API

## Estrutura do Projeto

O projeto segue as melhores práticas de arquitetura em Go, utilizando separação de responsabilidades:

```
todo-go-backend/
├── cmd/
│   └── api/
│       └── main.go              # Ponto de entrada da aplicação
├── internal/
│   ├── config/                  # Configurações da aplicação
│   ├── database/                # Conexão e setup do banco de dados
│   ├── errors/                  # Erros customizados da aplicação
│   ├── handlers/                # Handlers HTTP (camada de apresentação)
│   ├── middleware/              # Middlewares (autenticação, CORS)
│   ├── models/                  # Modelos de dados (entidades)
│   ├── notifications/           # Sistema de notificações (Email, Telegram)
│   ├── repositories/            # Camada de acesso a dados (Repository Pattern)
│   └── services/                # Camada de lógica de negócio (Service Layer)
├── pkg/
│   └── utils/                   # Utilitários (JWT, password hashing)
├── docs/                        # Documentação Swagger/OpenAPI
├── .github/
│   └── workflows/               # Pipelines CI/CD (GitHub Actions)
├── docker-compose.yml           # Configuração Docker Compose
├── Dockerfile                    # Imagem Docker da aplicação
├── env.example                  # Exemplo de variáveis de ambiente
└── go.mod                       # Dependências do projeto
```

### Arquitetura

O projeto utiliza uma arquitetura em camadas:

1. **Handlers**: Recebem requisições HTTP, validam entrada e chamam services
2. **Services**: Contêm a lógica de negócio da aplicação
3. **Repositories**: Abstraem o acesso aos dados, permitindo fácil troca de banco de dados
4. **Models**: Definem as entidades do domínio
5. **Errors**: Erros customizados com códigos HTTP apropriados

## Instalação

1. Clone o repositório:
```bash
git clone <repository-url>
cd todo-go-backend
```

2. Instale as dependências:
```bash
go mod download
```

3. Configure as variáveis de ambiente (opcional):
```bash
# Crie um arquivo .env ou exporte as variáveis
export PORT=8080
export JWT_SECRET=your-secret-key-change-in-production
export DATABASE_PATH=todo.db
```

4. Execute a aplicação:
```bash
go run cmd/api/main.go
```

A API estará disponível em `http://localhost:8080`

## Endpoints

> **Referência atualizada:** use o Swagger/OpenAPI (`/swagger/index.html` ou `/openapi.json`) como fonte da verdade. A API inclui paginação, busca, filtros avançados, grupos, compartilhamento, lembretes com horário, Web Push (`/notifications/push/*`), exportação/exclusão de conta (`/users/me/export`, `DELETE /users/me`) e rate limit em `/auth`.

### Autenticação

#### Registrar usuário
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "usuario",
  "email": "usuario@example.com",
  "password": "senha123"
}
```

**Resposta:**
```json
{
  "message": "User created successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "usuario",
    "email": "usuario@example.com"
  }
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "usuario",
  "password": "senha123"
}
```

**Resposta:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "usuario",
    "email": "usuario@example.com"
  }
}
```

### Tarefas (Requer autenticação)

Todas as rotas de tarefas requerem o header:
```
Authorization: Bearer <token>
```

#### Criar tarefa
```http
POST /api/v1/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Limpar a casa",
  "description": "Limpar todos os cômodos",
  "type": "casa",
  "due_date": "2024-12-31T23:59:59Z",
  "user_id": 2  // Opcional: criar tarefa para outro usuário
}
```

**Tipos válidos:** `casa`, `trabalho`, `lazer`, `saude`

#### Listar tarefas
```http
GET /api/v1/tasks?type=casa&completed=false&page=1&limit=10&search=texto&priority=alta
Authorization: Bearer <token>
```

**Query parameters opcionais (principais):**
- `type`, `completed`, `search`, `priority`, `tag_ids`
- `page`, `limit`, `sort_by`, `order`
- Filtros de data: `due_date_from`, `due_date_to`, `overdue`, `today`, `this_week`, `this_month`
- `GET /api/v1/tasks/assigned` — tarefas que você delegou a outros

#### Obter tarefa específica
```http
GET /api/v1/tasks/:id
Authorization: Bearer <token>
```

#### Atualizar tarefa
```http
PUT /api/v1/tasks/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Título atualizado",
  "completed": true,
  "due_date": "2024-12-31T23:59:59Z"
}
```

#### Deletar tarefa
```http
DELETE /api/v1/tasks/:id
Authorization: Bearer <token>
```

### Tags (Requer autenticação)

#### Criar tag
```http
POST /api/v1/tags
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "urgente",
  "color": "#FF0000"
}
```

#### Listar tags
```http
GET /api/v1/tags
Authorization: Bearer <token>
```

#### Obter tag específica
```http
GET /api/v1/tags/:id
Authorization: Bearer <token>
```

#### Atualizar tag
```http
PUT /api/v1/tags/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "importante",
  "color": "#00FF00"
}
```

#### Deletar tag
```http
DELETE /api/v1/tags/:id
Authorization: Bearer <token>
```

### Comentários (Requer autenticação)

#### Criar comentário
```http
POST /api/v1/comments
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Este é um comentário na tarefa",
  "task_id": 1
}
```

#### Listar comentários de uma tarefa
```http
GET /api/v1/tasks/:id/comments
Authorization: Bearer <token>
```

#### Obter comentário específico
```http
GET /api/v1/comments/:id
Authorization: Bearer <token>
```

#### Atualizar comentário
```http
PUT /api/v1/comments/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "content": "Comentário atualizado"
}
```

#### Deletar comentário
```http
DELETE /api/v1/comments/:id
Authorization: Bearer <token>
```

### Notificações (Requer autenticação)

#### Configurar Telegram Chat ID
```http
PUT /api/v1/users/telegram-chat-id
Authorization: Bearer <token>
Content-Type: application/json

{
  "telegram_chat_id": "123456789"
}
```

#### Habilitar/Desabilitar notificações
```http
PUT /api/v1/users/notifications-enabled
Authorization: Bearer <token>
Content-Type: application/json

{
  "notifications_enabled": true
}
```

#### Testar notificações
```http
POST /api/v1/notifications/test
Authorization: Bearer <token>
```

### Health Check

#### Verificar saúde da API
```http
GET /health
```

**Resposta:**
```json
{
  "status": "ok"
}
```

### Documentação (Swagger/OpenAPI)

#### Interface interativa
```http
GET /swagger/index.html
```

Acesse a documentação interativa da API no navegador.

#### Especificação JSON
```http
GET /swagger/swagger.json
```

Retorna a especificação OpenAPI em formato JSON. Útil para ferramentas como `openapi-typescript`.

**Resposta:** Especificação OpenAPI completa em JSON

#### Especificação YAML
```http
GET /swagger/swagger.yaml
```

Retorna a especificação OpenAPI em formato YAML.

**Resposta:** Especificação OpenAPI completa em YAML

#### Especificação JSON (fallback)
```http
GET /swagger/doc.json
```

Endpoint alternativo para a especificação JSON.

## Exemplos de Uso

### Criar tarefa para outro usuário

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer <seu-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Revisar código",
    "description": "Revisar PR #123",
    "type": "trabalho",
    "due_date": "2024-12-25T18:00:00Z",
    "user_id": 2
  }'
```

### Listar tarefas pendentes de casa

```bash
curl -X GET "http://localhost:8080/api/v1/tasks?type=casa&completed=false" \
  -H "Authorization: Bearer <seu-token>"
```

### Grupos e convites

Usuários só podem **compartilhar** e **atribuir** tarefas a quem compartilha pelo menos um grupo com eles. Entrada em grupos é por **convite aceito** (exceto o criador do grupo).

No primeiro deploy, todos os usuários existentes são migrados automaticamente para o grupo padrão **"Os de casa"** (idempotente, sem perda de tarefas). Novos registros também entram nesse grupo.

```bash
# Listar meus grupos
curl -X GET http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer <seu-token>"

# Criar grupo
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer <seu-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Equipe"}'

# Convidar usuário (envia notificação in-app)
curl -X POST http://localhost:8080/api/v1/groups/1/invitations \
  -H "Authorization: Bearer <seu-token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 2}'

# Aceitar convite
curl -X POST http://localhost:8080/api/v1/group-invitations/1/accept \
  -H "Authorization: Bearer <seu-token>"

# Notificações in-app
curl -X GET http://localhost:8080/api/v1/notifications/in-app/unread-count \
  -H "Authorization: Bearer <seu-token>"
```

**Listagem de usuários:**

- `GET /users` — colegas de grupo (share/atribuir)
- `GET /users?scope=invite&group_id=1` — candidatos a convite

### Compartilhar tarefa com outros usuários

O dono da tarefa pode compartilhar apenas com usuários do **mesmo grupo**. Quem recebe o compartilhamento pode ver e editar a tarefa.

```bash
curl -X POST http://localhost:8080/api/v1/tasks/1/share \
  -H "Authorization: Bearer <seu-token>" \
  -H "Content-Type: application/json" \
  -d '{"user_ids": [2, 3, 4]}'
```

### Remover compartilhamento de uma tarefa

Apenas o dono da tarefa pode remover um usuário da lista de compartilhamento:

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1/share/2 \
  -H "Authorization: Bearer <seu-token>"
```

**Nota:** Quando um usuário cria uma tarefa para outro (`user_id` no body do POST), a tarefa já fica compartilhada entre os dois automaticamente.

## Testes

Execute os testes com:

```bash
go test ./...
```

Para ver cobertura:

```bash
go test -cover ./...
```

### Configuração de Testes

Os testes suportam tanto MySQL quanto SQLite:

- **MySQL (CI)**: Configure as variáveis de ambiente `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`
- **SQLite (Local)**: Requer CGO habilitado (`CGO_ENABLED=1`)

**Nota**: Os testes que utilizam SQLite requerem CGO habilitado. Se você encontrar erros relacionados ao CGO, certifique-se de que o CGO está habilitado no seu ambiente Go ou configure MySQL.

Para executar testes específicos:

```bash
go test ./internal/services/... -v
go test ./internal/handlers/... -v
```

### Testes na CI

A pipeline de CI usa MySQL automaticamente. Os testes são executados automaticamente em cada push e pull request.

## Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `PORT` | Porta do servidor | `8080` |
| `JWT_SECRET` | Chave secreta para JWT | `your-secret-key-change-in-production` |
| `DATABASE_PATH` | Caminho do arquivo SQLite | `todo.db` |
| `DATABASE_HOST` | Host do MySQL (se usando MySQL) | - |
| `DATABASE_PORT` | Porta do MySQL | `3306` |
| `DATABASE_USER` | Usuário do MySQL | - |
| `DATABASE_PASSWORD` | Senha do MySQL | - |
| `DATABASE_NAME` | Nome do banco de dados MySQL | - |
| `CORS_ALLOWED_ORIGINS` | Origens permitidas CORS (separadas por vírgula) | `*` |
| `CORS_ALLOWED_METHODS` | Métodos HTTP permitidos | `GET,POST,PUT,DELETE,OPTIONS,PATCH` |
| `CORS_ALLOWED_HEADERS` | Headers permitidos | `Content-Type,Authorization,Accept,Origin` |
| `CORS_ALLOW_CREDENTIALS` | Permitir credenciais | `true` |
| `CORS_MAX_AGE` | Max age para preflight requests (segundos) | `3600` |
| `NOTIFICATIONS_ENABLED` | Habilitar notificações | `true` |
| `NOTIFICATION_CHECK_INTERVAL` | Intervalo de verificação (cron) | `* * * * *` |
| `RUN_SCHEDULER` | Executar scheduler neste processo (`false` em réplicas extras) | `true` |
| `VAPID_PUBLIC_KEY` | Chave pública VAPID para Web Push | - |
| `VAPID_PRIVATE_KEY` | Chave privada VAPID para Web Push | - |
| `VAPID_SUBJECT` | Contato VAPID (`mailto:` ou URL HTTPS) | - |
| `SMTP_HOST` | Host SMTP para email | - |
| `SMTP_PORT` | Porta SMTP | `587` |
| `SMTP_USER` | Usuário SMTP | - |
| `SMTP_PASSWORD` | Senha SMTP | - |
| `SMTP_FROM` | Email remetente | - |
| `TELEGRAM_BOT_TOKEN` | Token do bot Telegram | - |
| `CLOUDFLARE_TUNNEL_TOKEN` | Token do Cloudflare Tunnel | - |

Veja o arquivo `env.example` para um exemplo completo de configuração.

## Swagger Documentation

The API is fully documented with Swagger/OpenAPI. After starting the server, you can access the interactive documentation at:

```
http://localhost:8080/swagger/index.html
```

### Swagger Endpoints

The following endpoints are available for accessing the Swagger/OpenAPI specification:

- **Interactive UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI 3.0 JSON**: `http://localhost:8080/openapi.json` - OpenAPI 3.0 specification (recommended)
- **OpenAPI 3.0 YAML**: `http://localhost:8080/swagger/swagger.yaml` - Alternative format for the OpenAPI spec
- **Legacy JSON**: `http://localhost:8080/swagger.json` - Redirects to openapi.json (backward compatibility)
- **Default JSON**: `http://localhost:8080/swagger/doc.json` - Fallback endpoint

### Regenerating Documentation

To regenerate the OpenAPI 3.0 documentation after making changes:

```bash
swag-openapi3 init -g cmd/api/main.go -o ./docs --requiredByDefault
```

Or if you have it installed globally:

```bash
go install github.com/dimasdanz/swag-openapi3@latest
swag-openapi3 init -g cmd/api/main.go -o ./docs --requiredByDefault
```

Isso atualiza `docs/openapi.json` a partir dos comentários `// @Param` nos handlers (por exemplo, `priority` e `tag_ids` em `GET /tasks` e `GET /tasks/assigned`). Rode de novo sempre que alterar a documentação Swagger nos handlers.

## Docker

### Using Docker Compose

The easiest way to run the application is using Docker Compose, which will start the API, MySQL database, and Cloudflare Tunnel:

```bash
# Build and start containers
docker compose up -d

# View logs
docker compose logs -f api

# View Cloudflare Tunnel logs
docker compose logs -f cloudflare

# Stop containers
docker compose down

# Stop and remove volumes (clean database)
docker compose down -v
```

The API will be available at `http://localhost:8080` and MySQL at `localhost:3306`.

### Cloudflare Tunnel

The docker-compose configuration includes a Cloudflare Tunnel service that automatically exposes your API through Cloudflare's network. This allows you to:

- Access your API from anywhere without exposing ports directly
- Benefit from Cloudflare's DDoS protection and CDN
- Use Cloudflare's authentication and access policies

To use the Cloudflare Tunnel:

1. Get your tunnel token from the [Cloudflare Zero Trust dashboard](https://one.dash.cloudflare.com/)
2. Add it to your `.env` file:
   ```env
   CLOUDFLARE_TUNNEL_TOKEN=your-tunnel-token-here
   ```
3. The tunnel will start automatically when you run `docker compose up -d`

**Note**: The Cloudflare Tunnel container will only start if `CLOUDFLARE_TUNNEL_TOKEN` is set. If you don't need the tunnel, you can comment out the `cloudflare` service in `docker-compose.yml` or simply not set the token.

### Environment Variables for Docker

You can create a `.env` file in the root directory. See `env.example` for all available variables:

```env
PORT=8080
JWT_SECRET=your-secret-key-change-in-production
DATABASE_HOST=mysql
DATABASE_PORT=3306
DATABASE_USER=todo_user
DATABASE_PASSWORD=todo_password
DATABASE_NAME=todo_db
MYSQL_ROOT_PASSWORD=root_password
MYSQL_DATABASE=todo_db
MYSQL_USER=todo_user
MYSQL_PASSWORD=todo_password
MYSQL_PORT=3306

# Cloudflare Tunnel (optional)
CLOUDFLARE_TUNNEL_TOKEN=your-cloudflare-tunnel-token
```

### Building Docker Image Manually

```bash
# Build the image
docker build -t todo-api .

# Run the container (requires MySQL to be running)
docker run -p 8080:8080 \
  -e DATABASE_HOST=mysql \
  -e DATABASE_USER=todo_user \
  -e DATABASE_PASSWORD=todo_password \
  -e DATABASE_NAME=todo_db \
  -e JWT_SECRET=your-secret-key \
  todo-api
```

## CI/CD

O projeto inclui pipelines de CI/CD usando GitHub Actions:

### Pipeline de Testes (CI)

- Executa automaticamente em push e pull requests
- Roda todos os testes usando MySQL
- Verifica se o código compila
- Localização: `.github/workflows/ci.yml`

### Pipeline de Deploy (CD)

- Deploy automático para Raspberry Pi usando runner self-hosted
- Build e deploy com Docker Compose
- Health check automático após deploy
- Localização: `.github/workflows/deploy.yml`

### Configuração do Runner Self-Hosted

Para configurar o runner na Raspberry Pi, veja a documentação em `RASPBERRY_PI_SETUP.md` (se disponível) ou siga a [documentação oficial do GitHub Actions](https://docs.github.com/en/actions/hosting-your-own-runners).

## Notificações

A API suporta notificações por Email e Telegram. Veja os arquivos de documentação:

- `NOTIFICATIONS_SETUP.md` - Configuração de notificações
- `TELEGRAM_CHAT_ID_GUIDE.md` - Como obter o Chat ID do Telegram
- `TEST_NOTIFICATIONS.md` - Como testar notificações
- `TROUBLESHOOTING_NOTIFICATIONS.md` - Solução de problemas

## Roadmap

Para ver o plano completo de melhorias futuras, consulte o arquivo [ROADMAP.md](./ROADMAP.md).

### Próximas Melhorias Prioritárias

- [x] Paginação nas listagens
- [x] Busca por texto nas tarefas
- [x] Filtros avançados (datas, período, `assigned_by`, prioridade, tags, ordenação)
- [x] Listagem de tarefas delegadas (`GET /tasks/assigned`)
- [x] Compartilhamento de tarefas (`POST/DELETE .../share`)
- [x] Notificações de tarefas próximas do vencimento
- [x] Suporte a múltiplos bancos de dados (SQLite, MySQL)
- [x] Documentação Swagger/OpenAPI
- [x] Docker e Docker Compose
- [x] CI/CD Pipeline
- [x] Sistema de Tags
- [x] Comentários em tarefas
- [x] Health check endpoint
- [x] CORS configurável
- [~] Rate limiting (parcial: `/auth` — 5 req/min por IP)
- [ ] Logs estruturados
- [ ] Cache com Redis
- [ ] Refresh tokens

## Documentação Adicional

- [ROADMAP.md](./ROADMAP.md) - Plano de melhorias futuras (técnico)
- [docs/PRODUCT_ROADMAP.md](./docs/PRODUCT_ROADMAP.md) - Ideias de produto e priorização por ciclo
- [NOTIFICATIONS_SETUP.md](./NOTIFICATIONS_SETUP.md) - Guia de configuração de notificações
- [TELEGRAM_CHAT_ID_GUIDE.md](./TELEGRAM_CHAT_ID_GUIDE.md) - Como obter o Chat ID do Telegram
- [TEST_NOTIFICATIONS.md](./TEST_NOTIFICATIONS.md) - Como testar notificações
- [TROUBLESHOOTING_NOTIFICATIONS.md](./TROUBLESHOOTING_NOTIFICATIONS.md) - Solução de problemas com notificações

## Contribuindo

Contribuições são bem-vindas! Por favor:

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

Certifique-se de que os testes passem antes de submeter o PR.

## Licença

MIT

