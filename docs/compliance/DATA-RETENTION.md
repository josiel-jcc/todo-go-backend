# Política de Retenção de Dados

**Controlador:** [RAZAO_SOCIAL]  
**Referência:** LGPD Art. 15; alinhado à [Política de Privacidade](./PRIVACY.md).

> **Fonte canônica:** `todo-frontend/docs/compliance/DATA-RETENTION.md`

---

## 1. Conta de usuário

| Situação | Retenção | Ação |
|----------|----------|------|
| Conta ativa | Indefinida enquanto o usuário utilizar o serviço | Dados mantidos no banco |
| Conta excluída pelo titular | Eliminação em até **30 dias** | Cascata conforme seção 3 |
| Conta inativa (sem login por **24 meses**) | Notificação por e-mail; após 30 dias sem resposta, exclusão | A definir operacionalmente |

---

## 2. Dados por categoria

| Categoria | Retenção | Após exclusão da conta |
|-----------|----------|------------------------|
| Perfil (username, e-mail, preferências de lembrete) | Vigência da conta | Exclusão |
| Senha (hash) | Vigência da conta | Exclusão |
| Tarefas (dono) | Vigência da conta | Exclusão |
| Tarefas (criadas para terceiros, `assigned_by`) | Vigência da conta do criador | `assigned_by` anulado; tarefa permanece com o destinatário |
| Comentários | Vigência da conta | Exclusão dos comentários do usuário |
| Tags | Vigência da conta | Exclusão |
| Grupos (membership) | Vigência da conta | Remoção de `group_members` do usuário |
| Convites de grupo (enviados/recebidos) | Vigência da conta | Exclusão de registros vinculados ao usuário |
| Notificações in-app (`user_notifications`) | Vigência da conta | Exclusão |
| Registros de envio (`notifications`) | 12 meses ou vigência da conta | Exclusão |
| Web Push (`push_subscriptions`) | Até revogação ou vigência da conta | Exclusão |
| `terms_accepted_at` | Vigência da conta | Exclusão |
| Logs de aplicação | 90 dias | Sem vínculo identificável quando possível |
| `financial_accounts` | Vigência da conta e do grupo (household) — **planejado / MVP** | Exclusão dos registros do titular; contas do household conforme política do grupo |
| `finance_categories` | Vigência do grupo (household) — **planejado / MVP** | Exclusão com o household ou quando não houver mais categorias em uso |
| `finance_transactions` | Vigência da conta — **planejado / MVP** | Exclusão em cascata (ver seção 3.1) |

---

## 3. Cascata de exclusão de conta (implementação)

Ordem executada em **transação** (`UserService.DeleteAccount`):

1. `user_notifications` onde `user_id = X`
2. `group_invitations` onde `invited_user_id = X` ou `invited_by_user_id = X`
3. `group_members` onde `user_id = X`
4. `push_subscriptions` onde `user_id = X`
5. `notifications` onde `user_id = X`
6. `comments` onde `user_id = X`
7. `tags` onde `user_id = X`
8. `task_shared_with` onde `user_id = X`
9. `tasks` onde `user_id = X` (dono)
10. `tasks` onde `assigned_by = X` → `assigned_by = NULL`
11. `users` — exclusão definitiva (`Unscoped`)

Sessões JWT revogadas no logout usam `token_denylist` conforme [SECURITY.md](./SECURITY.md); entradas expiram com o TTL do token.

### 3.1 Finanças — cascata na exclusão de conta (**planejado / MVP**)

Quando o módulo financeiro estiver implementado, a exclusão de conta do titular incluirá, em transação:

1. `finance_transactions` onde `user_id = X` (lançamentos privados e do titular)
2. `financial_accounts` pessoais onde `user_id = X` e não compartilhadas com o household
3. Remoção de papéis financeiros do titular no grupo (`finance_group_members` ou equivalente)
4. Lançamentos **household** criados pelo titular: exclusão ou anonimização do `user_id` conforme regra de negócio (a definir no MVP)

**Exclusão de conta financeira (`financial_accounts`):** ao excluir uma conta no app, todos os `finance_transactions` vinculados a essa conta serão removidos em **cascata** (FK `ON DELETE CASCADE` ou serviço equivalente). Categorias (`finance_categories`) permanecem no nível do grupo até remoção explícita pelo `admin` financeiro.

---

## 4. Exportação (portabilidade)

O endpoint `GET /users/me/export` inclui, entre outros: perfil (incl. preferências de notificação e lembrete), tarefas, tarefas compartilhadas, tags, comentários, grupos, convites de grupo e subscriptions Web Push ativas. **Finanças (planejado / MVP):** extensão futura para contas, categorias e lançamentos do titular (incl. visibilidade `private` e `household` acessíveis ao usuário).

---

## 5. Backup

Backups de banco de dados podem reter dados por até **30 dias** após exclusão lógica, sendo sobrescritos no ciclo normal de backup.

---

## 6. Revisão

Esta política é revisada anualmente ou quando houver mudança relevante no tratamento.
