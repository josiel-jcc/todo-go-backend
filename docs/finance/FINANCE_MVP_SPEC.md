# Especificação MVP — Módulo de finanças

**Versão:** 1.0 (MVP)  
**Última atualização:** 2026-05-29  
**Visão de produto:** [PRODUCT_VISION.md](./PRODUCT_VISION.md)

---

## 1. Objetivo

Entregar controle financeiro manual integrado aos **grupos** existentes: contas, categorias, lançamentos (receita, despesa, transferência), visibilidade privada ou da casa, e um resumo mensal. Sem Open Finance, cartões ou parcelamentos nesta fase.

---

## 2. Vínculo com grupos (`group_id`)

Todo dado financeiro pertence a um **grupo** (`groups.id`), não diretamente ao usuário isolado.

| Regra | Descrição |
|-------|-----------|
| Escopo | Contas, categorias e transações têm `group_id` NOT NULL |
| Acesso | Usuário precisa ser membro de `group_members` para qualquer operação no grupo |
| Grupo padrão | Usuário pode usar seu `is_default` group para finanças pessoais “solo” |
| Casa compartilhada | Grupo com vários membros = finanças da casa conforme papel e visibilidade |

A API sempre recebe `groupId` na URL; o backend valida membership antes de ler ou escrever.

---

## 3. Entidades

### 3.1 Account (conta)

Representa onde o dinheiro “está” (corrente, poupança, carteira, etc.).

| Campo | Tipo | Notas |
|-------|------|-------|
| `id` | uint | PK |
| `group_id` | uint | FK → `groups.id` |
| `name` | string(100) | Ex.: “Nubank”, “Carteira” |
| `type` | enum | `checking`, `savings`, `cash`, `other` |
| `currency` | string(3) | Default `BRL` |
| `initial_balance` | decimal(15,2) | Saldo na abertura da conta |
| `is_active` | bool | Default true |
| `created_by` | uint | FK → `users.id` |
| `created_at`, `updated_at` | timestamp | |
| `deleted_at` | soft delete | |

**Saldo atual (calculado):** `initial_balance` + soma de transações que afetam a conta (receitas/despesas na conta; transferências como origem/destino).

### 3.2 Category (categoria)

Classifica receitas e despesas. Transferências **não** usam categoria (ou categoria nula).

| Campo | Tipo | Notas |
|-------|------|-------|
| `id` | uint | PK |
| `group_id` | uint | FK → `groups.id` |
| `name` | string(80) | |
| `kind` | enum | `income`, `expense` |
| `icon` | string(40) | Opcional, slug ou emoji |
| `color` | string(7) | Opcional, hex |
| `sort_order` | int | Ordenação na UI |
| `is_system` | bool | Categorias seed (Alimentação, Moradia…) não editáveis no MVP |
| `created_at`, `updated_at` | timestamp | |

### 3.3 Transaction (lançamento)

| Campo | Tipo | Notas |
|-------|------|-------|
| `id` | uint | PK |
| `group_id` | uint | FK → `groups.id` |
| `type` | enum | `income`, `expense`, `transfer` |
| `account_id` | uint | Conta principal (receita/despesa) ou origem (transferência) |
| `transfer_account_id` | uint | Nullable; destino em `transfer` |
| `category_id` | uint | Nullable; obrigatório para income/expense |
| `amount` | decimal(15,2) | Sempre positivo; sinal inferido pelo `type` |
| `description` | string(255) | |
| `occurred_on` | date | Data do fato (não confundir com `created_at`) |
| `visibility` | enum | `private`, `household` — ver §4 |
| `created_by` | uint | FK → `users.id` |
| `created_at`, `updated_at` | timestamp | |
| `deleted_at` | soft delete | |

**Regras:**

- `income`: credita `account_id`; exige `category_id` com `kind=income`.
- `expense`: debita `account_id`; exige `category_id` com `kind=expense`.
- `transfer`: debita `account_id`, credita `transfer_account_id`; `category_id` nulo; ambas contas no mesmo `group_id`.

### 3.4 FinanceMemberRole (papel financeiro no grupo)

Estende membership de tarefas com permissão financeira. Uma linha por `(group_id, user_id)`.

| Campo | Tipo | Notas |
|-------|------|-------|
| `group_id` | uint | PK composta |
| `user_id` | uint | PK composta |
| `role` | enum | `admin`, `editor`, `viewer` |
| `granted_by` | uint | Quem atribuiu o papel |
| `updated_at` | timestamp | |

**Bootstrap:** criador do grupo = `admin`. Demais membros entram como `editor` ao aceitar convite (configurável no MVP como default `editor`).

| Papel | Contas / categorias | Transações | Ver privadas alheias | Gerenciar papéis |
|-------|---------------------|------------|----------------------|------------------|
| **admin** | CRUD | CRUD todas visíveis + criar | Não (exceto as próprias) | Sim |
| **editor** | CRUD | CRUD próprias + household | Não | Não |
| **viewer** | Leitura | Leitura household + próprias | Não | Não |

---

## 4. Visibilidade: privado vs casa

| Valor | Quem vê | Uso |
|-------|---------|-----|
| `private` | Apenas `created_by` | Gasto pessoal dentro do grupo da casa |
| `household` | Membros com papel no grupo (respeitando viewer/editor) | Despesa/receita compartilhada |

**Listagens e dashboard:**

- Agregações mensais (`/dashboard`, `/summary`) consideram só transações `household` **ou** `private` do usuário autenticado.
- Admin **não** vê transações `private` de outros membros (privacidade por design no MVP).

---

## 5. Esboço de tabelas (MySQL)

```sql
-- Contas
CREATE TABLE finance_accounts (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  group_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(100) NOT NULL,
  type ENUM('checking','savings','cash','other') NOT NULL DEFAULT 'checking',
  currency CHAR(3) NOT NULL DEFAULT 'BRL',
  initial_balance DECIMAL(15,2) NOT NULL DEFAULT 0,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  created_by BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  deleted_at DATETIME(3) NULL,
  INDEX idx_finance_accounts_group (group_id),
  CONSTRAINT fk_finance_accounts_group FOREIGN KEY (group_id) REFERENCES `groups`(id)
);

-- Categorias
CREATE TABLE finance_categories (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  group_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(80) NOT NULL,
  kind ENUM('income','expense') NOT NULL,
  icon VARCHAR(40) NULL,
  color VARCHAR(7) NULL,
  sort_order INT NOT NULL DEFAULT 0,
  is_system TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  UNIQUE KEY uk_finance_categories_group_name_kind (group_id, name, kind),
  INDEX idx_finance_categories_group (group_id)
);

-- Lançamentos
CREATE TABLE finance_transactions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  group_id BIGINT UNSIGNED NOT NULL,
  type ENUM('income','expense','transfer') NOT NULL,
  account_id BIGINT UNSIGNED NOT NULL,
  transfer_account_id BIGINT UNSIGNED NULL,
  category_id BIGINT UNSIGNED NULL,
  amount DECIMAL(15,2) NOT NULL,
  description VARCHAR(255) NOT NULL DEFAULT '',
  occurred_on DATE NOT NULL,
  visibility ENUM('private','household') NOT NULL DEFAULT 'household',
  created_by BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  deleted_at DATETIME(3) NULL,
  INDEX idx_finance_tx_group_date (group_id, occurred_on),
  INDEX idx_finance_tx_account (account_id),
  INDEX idx_finance_tx_created_by (created_by)
);

-- Papéis financeiros (1:1 com membership ativo)
CREATE TABLE finance_member_roles (
  group_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  role ENUM('admin','editor','viewer') NOT NULL DEFAULT 'editor',
  granted_by BIGINT UNSIGNED NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  PRIMARY KEY (group_id, user_id)
);
```

**Índices adicionais recomendados:** `(group_id, visibility, occurred_on)` para dashboard; FKs formais para `account_id`, `category_id` após estabilizar migrations.

---

## 6. API (outline)

Base: `/api/v1/finance/groups/{groupId}`  
Autenticação: JWT (mesmo middleware das tarefas).  
Respostas: JSON; erros 401/403/404/422 padronizados.

### 6.1 Contas

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/accounts` | Lista contas ativas do grupo (+ saldo calculado) |
| POST | `/accounts` | Cria conta (editor+) |
| GET | `/accounts/{accountId}` | Detalhe + saldo |
| PATCH | `/accounts/{accountId}` | Atualiza nome/tipo/active (editor+) |
| DELETE | `/accounts/{accountId}` | Soft delete se sem transações ou bloqueio se houver (admin) |

### 6.2 Categorias

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/categories` | Lista por `kind` (query `?kind=expense`) |
| POST | `/categories` | Cria categoria custom (editor+) |
| PATCH | `/categories/{categoryId}` | Edita (não `is_system`) |
| DELETE | `/categories/{categoryId}` | Remove se não referenciada |

### 6.3 Transações

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/transactions` | Filtros: `from`, `to`, `account_id`, `category_id`, `type`, `visibility` |
| POST | `/transactions` | Cria receita/despesa/transferência |
| GET | `/transactions/{transactionId}` | Detalhe |
| PATCH | `/transactions/{transactionId}` | Edita (autor ou admin para household) |
| DELETE | `/transactions/{transactionId}` | Soft delete (mesmas regras de edição) |

### 6.4 Dashboard / resumo mensal

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/dashboard` | Query: `month=YYYY-MM` (default mês atual) |
| GET | `/summary/monthly` | Alias ou sub-recurso; totais receita/despesa/saldo líquido |

**Payload exemplo — `GET .../dashboard?month=2026-05`:**

```json
{
  "month": "2026-05",
  "currency": "BRL",
  "totals": {
    "income": 8500.00,
    "expense": 6230.50,
    "net": 2269.50
  },
  "by_category": [
    { "category_id": 3, "name": "Alimentação", "kind": "expense", "total": 1200.00 }
  ],
  "accounts": [
    { "account_id": 1, "name": "Nubank", "balance": 3450.00 }
  ]
}
```

Respeita visibilidade: usuário só agrega o que pode ver (§4).

### 6.5 Papéis (MVP mínimo)

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/members/roles` | Lista papéis do grupo |
| PATCH | `/members/{userId}/role` | Admin altera papel |

---

## 7. Critérios de aceite (MVP)

### Funcional

- [ ] Membro do grupo consegue criar conta, categoria e lançamento de receita, despesa e transferência entre duas contas do mesmo grupo.
- [ ] Transferência atualiza saldo lógico de origem e destino de forma consistente.
- [ ] Lançamento `private` aparece apenas para o autor; `household` aparece para editores/viewers/admins do grupo.
- [ ] Viewer não consegue POST/PATCH/DELETE em contas, categorias ou transações household de outros.
- [ ] Admin consegue alterar papel de membro; editor não consegue.
- [ ] Dashboard mensal retorna totais e breakdown por categoria coerentes com transações visíveis ao usuário.
- [ ] Usuário não membro do grupo recebe 403 em qualquer rota financeira do grupo.

### Técnico

- [ ] Migrations criam as quatro tabelas sem quebrar `groups` / `group_members` existentes.
- [ ] Handlers registrados sob `/api/v1/finance/...` com testes de integração cobrindo visibilidade e papéis.
- [ ] OpenAPI (`docs/openapi.json`) documenta endpoints MVP.
- [ ] Soft delete em contas/transações onde especificado; categorias system não deletáveis.

### Produto / UX (backend pronto para frontend)

- [ ] Seed de categorias default por grupo na primeira operação financeira do grupo (ou migration de seed).
- [ ] Mensagens de erro claras: categoria errada para tipo, contas de grupos diferentes em transferência, saldo insuficiente **não** bloqueia MVP (sem validação de saldo negativo).

---

## 8. Fora de escopo (MVP)

| Item | Versão alvo |
|------|-------------|
| Open Finance / importação bancária | V3 |
| Cartão de crédito e fatura | V2 |
| Parcelamentos (N de M) | V2 |
| Orçamento por categoria | V2 |
| Anexos (comprovante PDF/foto) | V2+ |
| Multi-moeda com câmbio | V3 |
| Recorrência de lançamentos | V2 |
| Exportação CSV/OFX | V3 |
| Admin ver transações private de terceiros | Não previsto |

---

## 9. Ordem de implementação sugerida

1. Migrations + models (`FinanceAccount`, `FinanceCategory`, `FinanceTransaction`, `FinanceMemberRole`)
2. Middleware: `RequireGroupMember` + `RequireFinanceRole(minRole)`
3. CRUD contas e categorias + seed de categorias
4. CRUD transações com regras de tipo e visibilidade
5. Serviço de saldo + dashboard mensal
6. Testes + OpenAPI
7. Frontend (fora deste repo): telas de lançamento e resumo

---

## 10. Histórico

| Data | Alteração |
|------|-----------|
| 2026-05-29 | Criação da especificação MVP |
