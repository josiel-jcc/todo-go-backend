# Política de Retenção de Dados

**Controlador:** [RAZAO_SOCIAL]  
**Referência:** LGPD Art. 15; alinhado à [Política de Privacidade](./PRIVACY.md).

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
| Perfil (username, e-mail) | Vigência da conta | Anonimização ou exclusão |
| Senha (hash) | Vigência da conta | Exclusão |
| Tarefas (dono) | Vigência da conta | Exclusão |
| Tarefas (criadas para terceiros, `assigned_by`) | Vigência da conta do criador | `assigned_by` anulado; tarefa permanece com o destinatário |
| Comentários | Vigência da conta | Exclusão dos comentários do usuário |
| Tags | Vigência da conta | Exclusão |
| Notificações (registro) | 12 meses ou vigência da conta | Exclusão |
| `terms_accepted_at` | Vigência da conta | Exclusão |
| Logs de aplicação | 90 dias | Sem vínculo identificável quando possível |

---

## 3. Cascata de exclusão de conta (implementação)

Ordem executada em transação:

1. `notifications` onde `user_id = X`
2. `comments` onde `user_id = X`
3. `tags` onde `user_id = X`
4. `task_shared_with` onde `user_id = X`
5. `tasks` onde `user_id = X` (dono)
6. `tasks` onde `assigned_by = X` → `assigned_by = NULL`
7. `token_denylist` / sessões do usuário
8. `users` — soft delete ou anonimização

---

## 4. Backup

Backups de banco de dados podem reter dados por até **30 dias** após exclusão lógica, sendo sobrescritos no ciclo normal de backup.

---

## 5. Revisão

Esta política é revisada anualmente ou quando houver mudança relevante no tratamento.
