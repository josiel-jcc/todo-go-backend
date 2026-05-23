# 📧 Configuração de Notificações

Este guia explica como configurar lembretes de tarefas por email, Telegram e Web Push (PWA).

## 📋 Pré-requisitos

1. **Email SMTP**: Conta de email com acesso SMTP (Gmail, Outlook, etc.)
2. **Telegram Bot** (opcional): Bot do Telegram para notificações
3. **Web Push VAPID** (opcional): Par de chaves VAPID para notificações no navegador

---

## 📧 Configuração de Email (SMTP)

### Gmail

1. Ative a verificação em duas etapas na sua conta Google
2. Gere uma "Senha de app":
   - Acesse: https://myaccount.google.com/apppasswords
   - Selecione "App" e "Outro (nome personalizado)"
   - Digite "Todo API" e clique em "Gerar"
   - Copie a senha gerada (16 caracteres)

3. Configure no `.env`:
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=seu-email@gmail.com
SMTP_PASSWORD=senha-de-app-gerada
SMTP_FROM=noreply@todoapp.com
```

### Outlook/Hotmail

```env
SMTP_HOST=smtp-mail.outlook.com
SMTP_PORT=587
SMTP_USER=seu-email@outlook.com
SMTP_PASSWORD=sua-senha
SMTP_FROM=seu-email@outlook.com
```

### Outros provedores

Consulte a documentação do seu provedor de email para as configurações SMTP.

---

## 🤖 Configuração do Telegram Bot

### Passo 1: Criar o Bot

1. Abra o Telegram e procure por **@BotFather**
2. Envie o comando `/newbot`
3. Escolha um nome para o bot (ex: "Todo Notifications Bot")
4. Escolha um username (deve terminar em `bot`, ex: `todo_notifications_bot`)
5. **Copie o token** fornecido pelo BotFather (algo como: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

### Passo 2: Configurar no `.env`

```env
TELEGRAM_BOT_TOKEN=seu-token-aqui
```

### Passo 3: Obter o Chat ID do Usuário

**Opção A: Via Bot (Recomendado)**

1. Envie uma mensagem para o seu bot no Telegram
2. Acesse: `https://api.telegram.org/bot<SEU_TOKEN>/getUpdates`
3. Procure por `"chat":{"id":123456789}` no JSON retornado
4. O número `123456789` é o seu Chat ID

**Opção B: Via @userinfobot**

1. Procure por **@userinfobot** no Telegram
2. Inicie uma conversa
3. O bot retornará seu Chat ID

### Passo 4: Configurar Chat ID no Sistema

Use o endpoint da API para configurar:

```bash
PUT /api/v1/users/telegram-chat-id
Authorization: Bearer <seu-token-jwt>

{
  "telegram_chat_id": "123456789"
}
```

---

## 🔔 Web Push (VAPID)

### Gerar chaves VAPID

```bash
npx web-push generate-vapid-keys
```

Copie `publicKey` e `privateKey` para o `.env`:

```env
VAPID_PUBLIC_KEY=sua-chave-publica
VAPID_PRIVATE_KEY=sua-chave-privada
VAPID_SUBJECT=mailto:admin@example.com
```

- `VAPID_SUBJECT` deve ser um `mailto:` ou URL HTTPS de contato do aplicativo.
- Em produção, a API deve ser servida via **HTTPS** (requisito do Web Push).

### Fluxo de inscrição (subscribe)

1. O frontend obtém a chave pública:
   ```bash
   GET /api/v1/notifications/push/vapid-public-key
   Authorization: Bearer <token>
   ```
   Resposta: `{ "public_key": "..." }`

2. O navegador solicita permissão e cria a subscription via `pushManager.subscribe()` usando a chave VAPID.

3. O frontend envia a subscription ao backend:
   ```bash
   POST /api/v1/notifications/push/subscribe
   Authorization: Bearer <token>
   Content-Type: application/json

   {
     "endpoint": "https://fcm.googleapis.com/fcm/send/...",
     "keys": {
       "p256dh": "...",
       "auth": "..."
     },
     "user_agent": "Mozilla/5.0 ..." 
   }
   ```

4. Para remover (logout ou desativar push no dispositivo):
   ```bash
   DELETE /api/v1/notifications/push/subscribe
   Authorization: Bearer <token>

   {
     "endpoint": "https://fcm.googleapis.com/fcm/send/..."
   }
   ```

Cada dispositivo/navegador gera um `endpoint` distinto; o backend armazena por usuário.

---

## ⚙️ Configuração Geral

### Variáveis de Ambiente

Adicione ao seu `.env`:

```env
# Ativar/desativar notificações
NOTIFICATIONS_ENABLED=true

# Intervalo de verificação (formato cron)
# Produção: a cada minuto
NOTIFICATION_CHECK_INTERVAL=* * * * *

# Executar o scheduler neste processo (default: true)
# Em múltiplas réplicas da API, defina false nas réplicas extras
RUN_SCHEDULER=true

# Email SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=seu-email@gmail.com
SMTP_PASSWORD=sua-senha-app
SMTP_FROM=noreply@todoapp.com

# Telegram Bot
TELEGRAM_BOT_TOKEN=seu-token-do-botfather

# Web Push VAPID
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=
VAPID_SUBJECT=mailto:admin@example.com
```

---

## 🔔 Tipos de Notificações

O sistema envia **lembretes com horário** (`task_reminder`):

- Um lembrete por tarefa e canal quando `reminder_at = due_date - offset` cai na janela do scheduler (último minuto).
- `offset` efetivo = `task.reminder_minutes_before` (se definido) ou `user.reminder_minutes_before` (padrão: **10** minutos).
- Valores permitidos: **5, 10, 15, 30, 60** minutos antes do vencimento.
- Canais: email, Telegram (se configurado), Web Push (se inscrito), e sino in-app (`UserNotification` tipo `task_reminder`).

Os tipos legados `due_soon`, `due_today` e `overdue` **não são mais gerados**.

### Escala (≤ 5 000 tarefas ativas)

O scheduler **não varre a tabela inteira**. A cada tick, uma consulta indexada retorna apenas tarefas com:

- `completed = false`
- `due_date` não nulo
- `users.notifications_enabled = true`
- `due_date` na janela `[now + 4 min, now + 61 min]` (cobre offsets 5–60 min)

O envio em si filtra `reminder_at` na janela `[now - 1 min, now)`. Projetado para até **5 000** tarefas ativas com `due_date`; fora dessa faixa horária, a carga por tick permanece baixa.

---

## 👤 Configuração por Usuário

### Ativar/Desativar Notificações

```bash
PUT /api/v1/users/notifications-enabled
Authorization: Bearer <token>

{
  "notifications_enabled": true
}
```

### Lembrete padrão (minutos antes do vencimento)

```bash
PUT /api/v1/users/reminder-settings
Authorization: Bearer <token>

{
  "reminder_minutes_before": 10
}
```

Valores: `5`, `10`, `15`, `30` ou `60`.

### Configurar Telegram Chat ID

```bash
PUT /api/v1/users/telegram-chat-id
Authorization: Bearer <token>

{
  "telegram_chat_id": "123456789"
}
```

Para remover o Telegram:
```json
{
  "telegram_chat_id": null
}
```

### Override por tarefa

Ao criar ou editar uma tarefa, envie `reminder_minutes_before` (opcional). Se omitido, usa o padrão do usuário; `null` na edição remove o override.

---

## 🧪 Testando

### Teste Manual

1. Configure `NOTIFICATION_CHECK_INTERVAL=* * * * *` (a cada minuto)
2. Defina `reminder_minutes_before` (ex.: 5) e crie uma tarefa com `due_date` ≈ agora + 5 minutos
3. Aguarde o próximo tick do scheduler
4. Verifique email, Telegram, push e o sino in-app

### Verificar Logs

O scheduler registra no log:
```
Running notification check...
Notification check completed
```

---

## ❓ Troubleshooting

### Email não está sendo enviado

- Verifique as credenciais SMTP
- Para Gmail, use "Senha de app" (não a senha normal)
- Verifique se o firewall não está bloqueando a porta SMTP

### Telegram não está funcionando

- Verifique se o token do bot está correto
- Verifique se o Chat ID está correto
- Envie uma mensagem para o bot antes de configurar o Chat ID
- Verifique os logs do servidor para erros

### Web Push não funciona

- Confirme `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY` e `VAPID_SUBJECT`
- API e PWA devem estar em HTTPS em produção
- O usuário deve ter chamado `POST /notifications/push/subscribe` após conceder permissão
- Subscriptions expiradas (410) são removidas automaticamente pelo servidor

### Notificações não estão sendo enviadas

- Verifique se `NOTIFICATIONS_ENABLED=true`
- Verifique se `RUN_SCHEDULER=true` neste processo (em deploy com várias réplicas, apenas uma deve rodar o scheduler)
- Verifique se o usuário tem `notifications_enabled=true`
- Verifique se a tarefa tem `due_date` configurado e `completed=false`
- Confirme que `reminder_at` ainda está na janela (lembretes atrasados > 1 minuto não são reenviados)

---

## 📝 Notas

- Deduplicação por `(task_id, channel, due_date)` — alterar `due_date` permite novo lembrete
- Tarefas completadas não entram na consulta de candidatos
- O scheduler roda em background e não bloqueia a API
- Histórico de notificações é salvo no banco de dados (`notifications` e `user_notifications`)
