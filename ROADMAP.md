# Roadmap - Todo Go Backend API

Este documento apresenta o plano de melhorias futuras para o projeto, organizado por categorias e prioridades.

> **Planejamento de produto:** ideias priorizadas por impacto, ciclos sugeridos e escolha por público-alvo — ver [docs/PRODUCT_ROADMAP.md](./docs/PRODUCT_ROADMAP.md).

## 📋 Índice

1. [Funcionalidades](#funcionalidades)
2. [Performance e Escalabilidade](#performance-e-escalabilidade)
3. [Segurança](#segurança)
4. [Observabilidade e Monitoramento](#observabilidade-e-monitoramento)
5. [DevOps e CI/CD](#devops-e-cicd)
6. [Qualidade de Código e Testes](#qualidade-de-código-e-testes)
7. [UX/API](#uxapi)
8. [Arquitetura](#arquitetura)

---

## 🚀 Funcionalidades

### Alta Prioridade

#### 1. Paginação nas Listagens ✅
- **Descrição**: Implementar paginação para listagem de tarefas
- **Benefícios**: Melhor performance e UX em listas grandes
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Adicionar parâmetros `page` e `limit` nos endpoints de listagem
  - ✅ Retornar metadata: `total`, `page`, `limit`, `total_pages`
  - ✅ Padrão: 10 itens por página, máximo 100

#### 2. Busca por Texto nas Tarefas ✅
- **Descrição**: Permitir busca por título e descrição
- **Benefícios**: Facilita encontrar tarefas específicas
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Parâmetro `search` no GET `/tasks`
  - ✅ Busca case-insensitive em `title` e `description`
  - ✅ Suporte a busca parcial (LIKE)

#### 3. Filtros Avançados ✅
- **Descrição**: Expandir filtros de busca
- **Benefícios**: Maior flexibilidade na consulta
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Filtro por data: `due_date_from`, `due_date_to`
  - ✅ Filtro por período: `overdue`, `today`, `this_week`, `this_month`
  - ✅ Filtro por usuário que atribuiu: `assigned_by`
  - ✅ Filtro por prioridade: `priority` (baixa, media, alta, urgente)
  - ✅ Filtro por tags: `tag_ids` (IDs separados por vírgula, ex.: `1,2,3`)
  - ✅ Ordenação: `sort_by` (created_at, due_date, title, **priority**) e `order` (asc, desc)
  - ✅ Listagem de tarefas que o usuário delegou: `GET /tasks/assigned` (mesmos filtros de paginação)

#### 4. Notificações de Tarefas ✅
- **Descrição**: Sistema de notificações para tarefas próximas do vencimento
- **Benefícios**: Ajuda usuários a não esquecerem tarefas importantes
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Job scheduler (cron) para verificar tarefas próximas
  - ✅ Notificações via email (SMTP)
  - ✅ Notificações via Telegram
  - ✅ Configuração de lembretes (1 dia antes, hoje, atrasadas)
  - ✅ Histórico de notificações enviadas
  - ✅ Configuração por usuário (ativar/desativar, Telegram Chat ID)

### Média Prioridade

#### 5. Prioridades de Tarefas ✅
- **Descrição**: Adicionar campo de prioridade (baixa, média, alta, urgente)
- **Benefícios**: Organização melhor das tarefas
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Enum `Priority` no modelo Task
  - ✅ Filtro por prioridade
  - ✅ Ordenação por prioridade

#### 6. Tags/Categorias Customizadas ✅
- **Descrição**: Permitir que usuários criem tags personalizadas além dos tipos fixos
- **Benefícios**: Maior flexibilidade na organização
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Modelo `Tag` com relação many-to-many com `Task`
  - ✅ CRUD de tags por usuário
  - ✅ Filtro por tags

#### 7. Subtarefas
- **Descrição**: Permitir criar subtarefas dentro de uma tarefa principal
- **Benefícios**: Organização de tarefas complexas
- **Implementação**:
  - Campo `parent_task_id` no modelo Task
  - Validação de recursão
  - Endpoint para listar subtarefas

#### 8. Comentários em Tarefas ✅
- **Descrição**: Sistema de comentários para comunicação entre usuários
- **Benefícios**: Colaboração melhorada
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Modelo `Comment` relacionado a `Task`
  - ✅ CRUD de comentários

#### 9. Grupos com convites e colaboração restrita ✅
- **Descrição**: Grupos de usuários; compartilhar/atribuir tarefas apenas entre membros do mesmo grupo; convites com aceite e notificações in-app
- **Benefícios**: Controle de quem colabora; migração segura com grupo padrão "Os de casa"
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Modelos `Group`, `GroupMember`, `GroupInvitation`, `UserNotification`
  - ✅ APIs `/groups`, `/group-invitations`, `/notifications/in-app`
  - ✅ Migração idempotente no startup; novos usuários no grupo padrão
  - ✅ `GET /users` filtrado; `?scope=invite` para convites
  - ⏳ Notificações quando alguém comenta (deixado para implementação futura)

#### 10. Anexos/Arquivos
- **Descrição**: Permitir anexar arquivos às tarefas
- **Benefícios**: Contexto adicional para tarefas
- **Implementação**:
  - Modelo `Attachment`
  - Upload de arquivos (S3 ou storage local)
  - Limite de tamanho e tipos permitidos

#### 11. Histórico de Alterações
- **Descrição**: Log de todas as alterações em tarefas
- **Benefícios**: Auditoria e rastreabilidade
- **Implementação**:
  - Modelo `TaskHistory` ou usar soft deletes do GORM
  - Registrar: quem, quando, o que mudou
  - Endpoint para visualizar histórico

### Baixa Prioridade

#### 12. Templates de Tarefas
- **Descrição**: Criar templates reutilizáveis de tarefas
- **Benefícios**: Agilidade na criação de tarefas recorrentes

#### 13. Recorrência de Tarefas
- **Descrição**: Tarefas que se repetem automaticamente
- **Benefícios**: Para tarefas periódicas (diárias, semanais, mensais)

#### 14. Compartilhamento de Tarefas ✅
- **Descrição**: Compartilhar tarefas com outros usuários para visualização e edição colaborativa (dono da tarefa = `user_id` da tarefa; apenas o dono adiciona/remove compartilhados)
- **Benefícios**: Colaboração sem precisar reatribuir a tarefa
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Relação many-to-many `task_shared_with` / `SharedWithUsers` no modelo
  - ✅ `POST /api/v1/tasks/{id}/share` com corpo `{ "user_ids": [2,3] }`
  - ✅ `DELETE /api/v1/tasks/{id}/share/{user_id}` para remover um usuário da lista
  - ✅ Documentado no OpenAPI (Swagger)

#### 15. Dashboard/Estatísticas ✅
- **Descrição**: Endpoint com estatísticas do usuário
- **Benefícios**: Visão geral do progresso
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ `GET /api/v1/stats` — resumo, hoje, por tipo, por prioridade, em progresso
  - ✅ Frontend consome via `useTaskStats` na página inicial

---

## ⚡ Performance e Escalabilidade

### Alta Prioridade

#### 1. Cache de Dados
- **Descrição**: Implementar cache para consultas frequentes
- **Benefícios**: Redução de carga no banco de dados
- **Implementação**:
  - Redis para cache
  - Cache de listas de tarefas (TTL: 5 minutos)
  - Cache de informações de usuário
  - Invalidação inteligente

#### 2. Índices de Banco de Dados
- **Descrição**: Otimizar índices para queries frequentes
- **Benefícios**: Queries mais rápidas
- **Implementação**:
  - Índice composto em `(user_id, completed, due_date)`
  - Índice em `assigned_by`
  - Índice full-text para busca (se suportado)

#### 3. Connection Pooling
- **Descrição**: Configurar pool de conexões do banco
- **Benefícios**: Melhor gerenciamento de recursos
- **Implementação**:
  - Configurar `SetMaxOpenConns`, `SetMaxIdleConns` no GORM
  - Monitorar uso de conexões

### Média Prioridade

#### 4. Lazy Loading de Relacionamentos
- **Descrição**: Carregar relacionamentos apenas quando necessário
- **Benefícios**: Redução de queries desnecessárias
- **Implementação**:
  - Usar `Select()` do GORM
  - Parâmetro `include` para escolher relacionamentos

#### 5. Compressão de Respostas
- **Descrição**: Comprimir respostas HTTP (gzip)
- **Benefícios**: Menor uso de banda
- **Implementação**:
  - Middleware de compressão no Gin

#### 6. Rate Limiting por Usuário
- **Descrição**: Limitar requisições por usuário
- **Benefícios**: Prevenção de abuso e DDoS
- **Implementação**:
  - Middleware de rate limiting
  - Diferentes limites por tipo de endpoint

---

## 🔒 Segurança

### Alta Prioridade

#### 1. Rate Limiting Global
- **Descrição**: Limitar requisições por IP em rotas públicas e protegidas
- **Benefícios**: Proteção contra DDoS e brute force
- **Status**: ⏳ **Parcial** — `AuthRateLimitMiddleware` em `/auth` (5 req/min por IP)
- **Implementação**:
  - ✅ Rate limit em `/api/v1/auth/*`
  - ⏳ Middleware global (ex.: 100 req/min por IP) e limites por usuário autenticado

#### 2. Validação de Entrada Robusta
- **Descrição**: Melhorar validações de input
- **Benefícios**: Prevenção de injection e dados inválidos
- **Implementação**:
  - Validação de tamanho máximo de strings
  - Sanitização de inputs
  - Validação de formato de email mais rigorosa

#### 3. Refresh Tokens
- **Descrição**: Implementar sistema de refresh tokens
- **Benefícios**: Segurança melhorada e melhor UX
- **Implementação**:
  - Access token (15 min) + Refresh token (7 dias)
  - Endpoint `/auth/refresh`
  - Rotação de refresh tokens

#### 4. CORS Configurável ✅
- **Descrição**: Configurar CORS adequadamente
- **Benefícios**: Segurança em aplicações web
- **Status**: ✅ **Implementado**
- **Implementação**:
  - ✅ Middleware CORS no Gin
  - ✅ Configuração via variáveis de ambiente
  - ✅ Suporte a origens, métodos, headers, credentials e max-age configuráveis

### Média Prioridade

#### 5. Logs de Auditoria
- **Descrição**: Registrar ações importantes
- **Benefícios**: Rastreabilidade e segurança
- **Implementação**:
  - Log de login/logout
  - Log de criação/edição/exclusão de tarefas
  - Log de tentativas de acesso não autorizado

#### 6. Hashing de Senhas Mais Forte
- **Descrição**: Considerar Argon2 além de bcrypt
- **Benefícios**: Segurança adicional
- **Implementação**:
  - Avaliar migração para Argon2
  - Manter compatibilidade com bcrypt

#### 7. 2FA (Two-Factor Authentication)
- **Descrição**: Autenticação de dois fatores
- **Benefícios**: Segurança adicional para contas
- **Implementação**:
  - TOTP (Google Authenticator, Authy)
  - Backup codes

#### 8. Política de Senhas
- **Descrição**: Forçar senhas mais seguras
- **Benefícios**: Redução de contas comprometidas
- **Implementação**:
  - Mínimo 8 caracteres
  - Requer maiúsculas, minúsculas, números
  - Validação no registro

---

## 📊 Observabilidade e Monitoramento

### Alta Prioridade

#### 1. Logs Estruturados
- **Descrição**: Implementar logging estruturado
- **Benefícios**: Melhor análise e debugging
- **Implementação**:
  - Usar `zerolog` ou `zap`
  - Formato JSON para produção
  - Níveis: DEBUG, INFO, WARN, ERROR
  - Contexto: request_id, user_id, etc.

#### 2. Health Check Melhorado
- **Descrição**: Health check que verifica dependências
- **Benefícios**: Monitoramento adequado
- **Implementação**:
  - Verificar conexão com banco
  - Status: `healthy`, `degraded`, `unhealthy`
  - Endpoint `/health/ready` e `/health/live`

#### 3. Métricas de Aplicação
- **Descrição**: Expor métricas Prometheus
- **Benefícios**: Monitoramento de performance
- **Implementação**:
  - Endpoint `/metrics`
  - Métricas: request duration, error rate, active users
  - Integração com Prometheus

### Média Prioridade

#### 4. Distributed Tracing
- **Descrição**: Rastreamento de requisições
- **Benefícios**: Debugging em sistemas distribuídos
- **Implementação**:
  - OpenTelemetry
  - Trace IDs em logs

#### 5. Alertas
- **Descrição**: Sistema de alertas para problemas
- **Benefícios**: Resposta rápida a incidentes
- **Implementação**:
  - Alertas para: alta taxa de erro, latência alta, banco offline
  - Integração com PagerDuty, Slack, etc.

#### 6. APM (Application Performance Monitoring)
- **Descrição**: Monitoramento detalhado de performance
- **Benefícios**: Identificação de gargalos
- **Implementação**:
  - Integração com New Relic, Datadog, ou similar
  - Profiling de queries lentas

---

## 🚢 DevOps e CI/CD

### Alta Prioridade

#### 1. CI/CD Pipeline ✅
- **Descrição**: Pipeline automatizado
- **Benefícios**: Deploy confiável e rápido
- **Status**: ✅ **Implementado (base)** — GitHub Actions com testes e deploy (ex.: runner self-hosted / Raspberry Pi conforme `deploy.yml`)
- **Implementação**:
  - ✅ GitHub Actions (`.github/workflows/ci.yml`, `deploy.yml`)
  - ✅ Testes no CI (MySQL)
  - ⏳ Security scan dedicado, staging separado e gates de aprovação: evolução futura

#### 2. Testes Automatizados
- **Descrição**: Suite completa de testes
- **Benefícios**: Confiança nas mudanças
- **Implementação**:
  - Testes unitários (cobertura > 80%)
  - Testes de integração
  - Testes E2E
  - Executar no CI

#### 3. Docker Multi-stage Otimizado
- **Descrição**: Otimizar Dockerfile
- **Benefícios**: Imagens menores e builds mais rápidos
- **Implementação**:
  - Usar distroless ou scratch
  - Cache de layers
  - .dockerignore otimizado

### Média Prioridade

#### 4. Kubernetes Deployment
- **Descrição**: Deploy em Kubernetes
- **Benefícios**: Escalabilidade e alta disponibilidade
- **Implementação**:
  - Helm charts
  - ConfigMaps e Secrets
  - HPA (Horizontal Pod Autoscaler)

#### 5. Secrets Management
- **Descrição**: Gerenciamento seguro de secrets
- **Benefícios**: Segurança
- **Implementação**:
  - HashiCorp Vault ou AWS Secrets Manager
  - Não commitar secrets

#### 6. Blue-Green Deployment
- **Descrição**: Deploy sem downtime
- **Benefícios**: Zero downtime em deploys
- **Implementação**:
  - Estratégia de deploy blue-green
  - Health checks antes de trocar tráfego

---

## 🧪 Qualidade de Código e Testes

### Alta Prioridade

#### 1. Aumentar Cobertura de Testes
- **Descrição**: Atingir > 80% de cobertura
- **Benefícios**: Menos bugs em produção
- **Implementação**:
  - Testes para todos os services
  - Testes para repositories
  - Testes para handlers
  - Mocks adequados

#### 2. Linters e Formatters
- **Descrição**: Padronização de código
- **Benefícios**: Código consistente
- **Implementação**:
  - `golangci-lint`
  - `gofmt` / `goimports`
  - Pre-commit hooks

#### 3. Documentação de Código
- **Descrição**: Melhorar documentação
- **Benefícios**: Manutenibilidade
- **Implementação**:
  - Comentários em todas as funções públicas
  - Exemplos de uso
  - README atualizado

### Média Prioridade

#### 4. Testes de Carga
- **Descrição**: Testes de performance
- **Benefícios**: Identificar gargalos
- **Implementação**:
  - k6 ou Apache Bench
  - Cenários: alta concorrência, muitos dados

#### 5. Análise Estática de Código
- **Descrição**: Detectar problemas no código
- **Benefícios**: Qualidade e segurança
- **Implementação**:
  - SonarQube ou similar
  - Integração no CI

---

## 🎨 UX/API

### Alta Prioridade

#### 1. Versionamento de API ✅ (parcial)
- **Descrição**: Suporte a múltiplas versões
- **Benefícios**: Evolução sem quebrar clientes
- **Status**: ✅ **Prefixo `/api/v1`** em uso em todas as rotas da API atual
- **Implementação**:
  - ✅ Base de rotas `/api/v1`
  - ⏳ `/api/v2`, deprecation headers e política de convivência de versões: futuro

#### 2. Respostas Consistentes
- **Descrição**: Padronizar formato de respostas
- **Benefícios**: Melhor experiência do desenvolvedor
- **Implementação**:
  - Wrapper para todas as respostas
  - Metadata consistente

#### 3. Validação de Erros Melhorada
- **Descrição**: Mensagens de erro mais claras
- **Benefícios**: Debugging mais fácil
- **Implementação**:
  - Códigos de erro específicos
  - Mensagens descritivas
  - Links para documentação

### Média Prioridade

#### 4. Webhooks
- **Descrição**: Notificações via webhooks
- **Benefícios**: Integração com outros sistemas
- **Implementação**:
  - CRUD de webhooks
  - Assinatura de eventos
  - Retry logic

#### 5. GraphQL API
- **Descrição**: Endpoint GraphQL opcional
- **Benefícios**: Flexibilidade para clientes
- **Implementação**:
  - GraphQL sobre REST
  - Schema bem definido

#### 6. Batch Operations
- **Descrição**: Operações em lote
- **Benefícios**: Eficiência
- **Implementação**:
  - POST `/tasks/batch` para criar múltiplas
  - PUT `/tasks/batch` para atualizar múltiplas

---

## 🏗️ Arquitetura

### Média Prioridade

#### 1. Event-Driven Architecture
- **Descrição**: Eventos para ações importantes
- **Benefícios**: Desacoplamento e escalabilidade
- **Implementação**:
  - Message broker (RabbitMQ, Kafka)
  - Eventos: task.created, task.completed, etc.

#### 2. Repository Pattern Melhorado
- **Descrição**: Abstrações mais robustas
- **Benefícios**: Testabilidade e flexibilidade
- **Implementação**:
  - Interfaces mais completas
  - Unit of Work pattern

#### 3. Dependency Injection
- **Descrição**: DI container
- **Benefícios**: Melhor testabilidade
- **Implementação**:
  - Wire ou Fx
  - Injeção automática de dependências

#### 4. Clean Architecture
- **Descrição**: Reorganizar em camadas mais claras
- **Benefícios**: Manutenibilidade
- **Implementação**:
  - Domain, Use Cases, Infrastructure
  - Dependências unidirecionais

---

## 📅 Priorização Sugerida

### Fase 1 (1-2 meses)
1. ✅ Paginação nas listagens
2. ✅ Busca por texto
3. ✅ Filtros avançados (incl. prioridade, tags, `GET /tasks/assigned`)
4. Logs estruturados
5. Rate limiting global (parcial: `/auth` já limitado)
6. Health check melhorado
7. ✅ CI/CD básico (GitHub Actions)

### Fase 2 (2-4 meses)
1. Cache (Redis)
2. ✅ Notificações de tarefas
3. Métricas Prometheus
4. Refresh tokens
5. Testes automatizados completos (aumentar cobertura)

### Fase 3 (4-6 meses)
1. ✅ Prioridades de tarefas
2. ✅ Tags customizadas
3. ✅ Compartilhamento de tarefas
4. Subtarefas
5. ✅ Comentários
6. Kubernetes deployment
7. Webhooks

---

## 📝 Notas

- Este roadmap é um documento vivo e deve ser atualizado conforme o projeto evolui
- Prioridades podem mudar baseado em feedback de usuários
- Algumas melhorias podem ser implementadas em paralelo
- Sempre considerar trade-offs entre complexidade e benefício

---

## ✅ Itens Concluídos

- ✅ **Paginação nas Listagens** (Dezembro 2025) - Implementado com suporte a `page` e `limit`, retornando metadata completa
- ✅ **Busca por Texto nas Tarefas** (Dezembro 2025) - Busca case-insensitive em título e descrição
- ✅ **Filtros Avançados** (Dezembro 2025 / Maio 2026) - Filtros por data, período, `assigned_by`, **prioridade** (`priority`), **tags** (`tag_ids`), ordenação incluindo `priority`, e listagem **`GET /tasks/assigned`**
- ✅ **Prioridades de Tarefas** (Dezembro 2025) - Campo de prioridade (baixa, média, alta, urgente) com filtro e ordenação
- ✅ **Tags/Categorias Customizadas** (Dezembro 2025) - Sistema completo de tags com CRUD e filtros, relação many-to-many com tarefas
- ✅ **Comentários em Tarefas** (Dezembro 2025) - Sistema completo de comentários com CRUD, controle de acesso baseado em propriedade/atribuição da tarefa
- ✅ **CORS Configurável** (Dezembro 2025) - Middleware CORS configurável via variáveis de ambiente com suporte a origens, métodos, headers, credentials e max-age
- ✅ **Notificações de Tarefas** (Dezembro 2024) - Sistema completo de notificações com email e Telegram, scheduler cron, histórico e configuração por usuário
- ✅ **Compartilhamento de Tarefas** (Maio 2026) - `POST/DELETE` em `/tasks/{id}/share`, many-to-many com usuários, controle pelo dono da tarefa
- ✅ **OpenAPI / listagens** (Maio 2026) - Parâmetros `priority`, `tag_ids` e `sort_by` com `priority` documentados em `GET /tasks` e `GET /tasks/assigned`
- ✅ **Grupos e convites** (Maio 2026) - Colaboração restrita por grupo, notificações in-app
- ✅ **Lembretes com horário e Web Push** (Maio 2026) - `task_reminder`, VAPID, subscriptions no backend
- ✅ **LGPD — exportação/exclusão** (Maio 2026) - Export JSON e exclusão em cascata incluindo push e grupos; docs de compliance sincronizadas
- ✅ **Dashboard / Estatísticas** (Maio 2026) - `GET /api/v1/stats` com agregados por resumo, hoje, tipo e prioridade

---

**Última atualização**: Maio 2026

