# Documentação de conformidade (LGPD)

**Fonte canônica:** arquivos neste diretório (`todo-frontend/docs/compliance/`).

O frontend exibe a política e os termos em `/privacidade` e `/termos` importando estes markdowns.

## Sincronização no monorepo

Ao alterar qualquer documento aqui, copie o mesmo conteúdo para:

- `docs/compliance/` (raiz do repositório)
- `todo-go-backend/docs/compliance/`

Ou execute (na raiz do repo):

```bash
cp todo-frontend/docs/compliance/*.md docs/compliance/
cp todo-frontend/docs/compliance/*.md todo-go-backend/docs/compliance/
```

## Preenchimento para produção

Substitua os placeholders `[RAZAO_SOCIAL]`, `[CNPJ]`, `[ENDERECO]`, `[EMAIL_PRIVACIDADE]`, `[PAIS_HOSPEDAGEM]` e registre alterações no histórico do ROPA antes do go-live.
