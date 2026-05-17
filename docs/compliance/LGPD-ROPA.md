# Registro de Operações de Tratamento (ROPA)

Template conforme Art. 37 da LGPD. Preencher e manter atualizado pela organização controladora.

**Controlador:** [RAZAO_SOCIAL] — CNPJ [CNPJ]  
**Encarregado:** [EMAIL_PRIVACIDADE]  
**Última revisão:** [DATA_ATUALIZACAO]

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
| Dados pessoais | Conteúdo das tarefas, metadados, IDs de usuários relacionados |
| Base legal | Execução de contrato |
| Compartilhamento | Outros usuários do app (mediante ação do titular) |
| Retenção | Enquanto conta ativa; exclusão em cascata ao apagar conta |

---

## Operação 3 — Notificações (e-mail e Telegram)

| Campo | Valor |
|-------|-------|
| Finalidade | Lembretes de vencimento de tarefas |
| Dados pessoais | E-mail, Telegram Chat ID, título/descrição da tarefa |
| Base legal | Consentimento (Art. 7, I) — opt-in |
| Operadores | SMTP, Telegram Bot API |
| Transferência internacional | Possível (Telegram) |
| Retenção | Registros de notificação: 12 meses ou até exclusão da conta |

---

## Operação 4 — Listagem de usuários (delegação)

| Campo | Valor |
|-------|-------|
| Finalidade | Permitir atribuir tarefas a outros usuários |
| Dados expostos | ID e username (e-mail **não** exposto na listagem) |
| Base legal | Execução de contrato |
| Minimização | Apenas dados necessários para seleção |

---

## Operação 5 — Logs e segurança

| Campo | Valor |
|-------|-------|
| Finalidade | Diagnóstico, segurança, auditoria |
| Dados | IP, user_id, timestamps, erros (sem PII em produção) |
| Base legal | Legítimo interesse |
| Retenção | Até 90 dias |

---

## Histórico de alterações

| Data | Alteração | Responsável |
|------|-----------|-------------|
| [DATA] | Criação do ROPA | [NOME] |
