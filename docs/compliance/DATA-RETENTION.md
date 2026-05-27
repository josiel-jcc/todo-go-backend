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

---

## 4. Exportação (portabilidade)

O endpoint `GET /users/me/export` inclui, entre outros: perfil (incl. preferências de notificação e lembrete), tarefas, tarefas compartilhadas, tags, comentários, grupos, convites de grupo e subscriptions Web Push ativas.

---

## 5. Backup

Backups de banco de dados podem reter dados por até **30 dias** após exclusão lógica, sendo sobrescritos no ciclo normal de backup.

---

## 6. Revisão

Esta política é revisada anualmente ou quando houver mudança relevante no tratamento.
