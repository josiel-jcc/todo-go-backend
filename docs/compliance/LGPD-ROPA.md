# Registro de Operações de Tratamento (ROPA)

Template conforme Art. 37 da LGPD. Preencher e manter atualizado pela organização controladora.

**Controlador:** [RAZAO_SOCIAL] — CNPJ [CNPJ]  
**Encarregado:** [EMAIL_PRIVACIDADE]  
**Última revisão:** 2026-05-27

> **Fonte canônica:** `todo-frontend/docs/compliance/LGPD-ROPA.md` (sincronizar cópias no monorepo).

---

## Operação 1 — Cadastro e autenticação

| Campo | Valor |
|-------|-------|
| Finalidade | Criar conta, autenticar usuário |
| Categorias de titulares | Usuários do app |
| Dados pessoais | Username, e-mail, senha (hash), data de aceite dos termos |
| Categorias de destinatários | Hospedagem (BD) |
| Transferência internacional | [SIM/NAO — detalhar] |
| Base legal | Execução de contrato (Art. 7, V) |
| Medidas de segurança | Bcrypt, JWT HttpOnly, HTTPS |
| Prazo de retenção | Até exclusão da conta + ver DATA-RETENTION.md |

---

## Operação 2 — Gestão de tarefas

| Campo | Valor |
|-------|-------|
| Finalidade | CRUD de tarefas, tags, comentários, compartilhamento |
| Dados pessoais | Conteúdo das tarefas, metadados, IDs de usuários relacionados, prioridade, lembrete por tarefa (`reminder_minutes_before` opcional) |
| Base legal | Execução de contrato |
| Compartilhamento | Outros usuários do mesmo grupo (mediante ação do titular) |
| Retenção | Enquanto conta ativa; exclusão em cascata ao apagar conta |

---

## Operação 3 — Lembretes (e-mail, Telegram, Web Push)

| Campo | Valor |
|-------|-------|
| Finalidade | Lembretes com horário (`due_date` − offset configurável) |
| Dados pessoais | E-mail, Telegram Chat ID, título/descrição da tarefa; para push: endpoint, `p256dh`, `auth`, `user_agent` opcional |
| Base legal | Consentimento (Art. 7, I) — `notifications_enabled`, Telegram, permissão de push no navegador |
| Operadores | SMTP, Telegram Bot API, serviço de push do navegador (FCM/Mozilla, etc.) |
| Transferência internacional | Possível (Telegram, push, nuvem) |
| Retenção | Registros em `notifications`: 12 meses ou até exclusão da conta; subscriptions push: até revogação ou exclusão da conta |

---

## Operação 4 — Listagem de usuários (delegação e convites)

| Campo | Valor |
|-------|-------|
| Finalidade | Atribuir/compartilhar tarefas e convidar para grupos |
| Dados expostos | ID e username (e-mail **não** exposto na listagem padrão) |
| Base legal | Execução de contrato |
| Minimização | Listagem filtrada por colegas de grupo; `GET /users?scope=invite&group_id=` para convites |

---

## Operação 5 — Grupos, convites e notificações in-app

| Campo | Valor |
|-------|-------|
| Finalidade | Organizar colaboração; convites com aceite; alertas no sino (convite de grupo, lembrete `task_reminder`) |
| Dados pessoais | Nome do grupo, membership, status de convite, payload de notificação in-app |
| Base legal | Execução de contrato |
| Retenção | Enquanto conta ativa; removido na exclusão da conta |
| Nota | Grupo padrão pode ser usado na migração de usuários existentes |

---

## Operação 6 — Web Push (subscriptions)

| Campo | Valor |
|-------|-------|
| Finalidade | Entregar lembretes nativos no PWA |
| Dados pessoais | Endpoint, chaves da subscription, user-agent opcional, vínculo com `user_id` |
| Base legal | Consentimento (Art. 7, I) |
| Retenção | Até desativação pelo usuário, invalidação do endpoint ou exclusão da conta |
| Medidas | Remoção no servidor ao desinscrever; limpeza na exclusão de conta |

---

## Operação 7 — Logs e segurança

| Campo | Valor |
|-------|-------|
| Finalidade | Diagnóstico, segurança, auditoria |
| Dados | IP, user_id, timestamps, erros (sem PII desnecessária em produção) |
| Base legal | Legítimo interesse |
| Retenção | Até 90 dias |

---

## Histórico de alterações

| Data | Alteração | Responsável |
|------|-----------|-------------|
| [DATA] | Criação do ROPA | [NOME] |
| 2026-05-27 | Grupos, Web Push, lembretes por minuto, sincronização com implementação | [NOME] |
