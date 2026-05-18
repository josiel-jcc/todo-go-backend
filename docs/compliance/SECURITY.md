# Postura de Segurança

Documento técnico de segurança do Todo App (frontend + backend). Complementa a [Política de Privacidade](./PRIVACY.md).

---

## Auditoria — achados e status

| ID | Achado | Severidade | Status |
|----|--------|------------|--------|
| S1 | JWT em localStorage (XSS) | Alta | Remediado — cookie HttpOnly |
| S2 | JWT_SECRET padrão fraco | Crítica | Remediado — validação em produção |
| S3 | CORS `*` + credenciais | Alta | Remediado — origens explícitas em produção |
| S4 | Sem rate limit em `/auth` | Média | Remediado |
| S5 | Swagger/debug públicos | Média | Remediado — desabilitado em produção |
| S6 | Enumeração de usuários no registro | Baixa | Remediado — mensagem genérica |
| S7 | `GET /users` expõe e-mails | Média | Remediado — só id + username |
| S8 | Logs com e-mail/Telegram | Média | Remediado |
| S9 | Erros 500 com `err.Error()` | Média | Remediado em produção |
| S10 | Senha mínima 6 caracteres | Média | Remediado — mínimo 8 + complexidade |

---

## Configuração de produção

### Variáveis obrigatórias (backend)

```env
APP_ENV=production
JWT_SECRET=<mínimo 32 caracteres aleatórios>
CORS_ALLOWED_ORIGINS=https://seu-frontend.exemplo.com
```

**Nunca** use `JWT_SECRET=your-secret-key-change-in-production` ou `CORS_ALLOWED_ORIGINS=*` em produção.

### Checklist de deploy

- [ ] `APP_ENV=production`
- [ ] `JWT_SECRET` com ≥ 32 caracteres (gerado com `openssl rand -base64 32`)
- [ ] `CORS_ALLOWED_ORIGINS` lista apenas o domínio do frontend
- [ ] Swagger (`/swagger/*`, `/openapi.json`) retorna 404
- [ ] `/notifications/debug` e `/notifications/test` indisponíveis
- [ ] MySQL sem porta publicada no host
- [ ] Secrets via variáveis de ambiente ou vault (não no repositório)
- [ ] HTTPS terminado (Cloudflare ou reverse proxy)

### Checklist Cloudflare

- [ ] TLS Full (strict) quando possível
- [ ] WAF / regras básicas habilitadas
- [ ] Token do Tunnel rotacionado periodicamente
- [ ] HSTS habilitado no painel Cloudflare
- [ ] Rate limiting na borda (opcional, complementar ao app)

---

## Autenticação

- Senhas: bcrypt (cost padrão).
- Sessão: JWT HS256 em cookie `auth_token` (`HttpOnly`, `Secure` em HTTPS, `SameSite=Lax`).
- Logout: limpa cookie e registra `jti` na denylist até expiração do token.
- Middleware aceita cookie ou header `Authorization: Bearer` (transição).

---

## Headers de segurança (API)

Aplicados via middleware Gin:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), microphone=(), camera=()`
- `Strict-Transport-Security` quando `APP_ENV=production` e request HTTPS

---

## Verificação manual pós-deploy

1. `document.cookie` no browser **não** contém `auth_token`.
2. Requisição de origem não listada no CORS é bloqueada.
3. `GET /openapi.json` retorna 404 em produção.
4. Exportação e exclusão de conta funcionam em Configurações.
5. `go test ./...` e testes do frontend passam.

---

## Reporte de vulnerabilidades

Envie detalhes para [EMAIL_PRIVACIDADE] com assunto "Segurança — Todo App". Não divulgue publicamente antes de correção.
