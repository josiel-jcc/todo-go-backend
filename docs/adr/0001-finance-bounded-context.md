# ADR 0001: Finanças como bounded context no hub familiar

**Status:** Proposto  
**Data:** 2026-05-29  
**Decisores:** Josiel / equipe  
**Afeta:** `todo-go-backend`, `todo-frontend`

## Contexto

O produto evolui de um app de tarefas colaborativo para um **hub pessoal/familiar** (Life Hub): tarefas, grupos, notificações e, em seguida, finanças no mesmo ecossistema de conta e lar.

Finanças têm regras de domínio distintas das tarefas (saldos, categorias, visibilidade por membro, papéis específicos, exportação de dados sensíveis). **Não** devem ser modeladas como tipo de tarefa nem compartilhar tabelas/rotas de `tasks`.

O conceito de **grupo** já existente (`groups`, convites, “Os de casa”) representa o **lar familiar (household)**. O módulo de finanças reutiliza esse agregado em vez de inventar um segundo “grupo familiar”.

## Decisão

1. **Bounded context**
   - Finanças é um módulo separado no backend e no frontend, com contratos e persistência próprios.
   - Tarefas permanecem em `/api/v1/tasks/*` (e domínios correlatos); finanças em **`/api/v1/finance/*`**.

2. **Household = `group_id`**
   - Todo contexto financeiro compartilhado do lar ancora em `group_id` (mesmo ID do grupo de colaboração).
   - Usuário pode pertencer a vários grupos; o contexto financeiro ativo é o grupo selecionado na sessão/UI.

3. **Papéis financeiros** (por membro, por grupo)
   - `admin` — configuração do lar, categorias, convites financeiros, exclusão de dados do household.
   - `editor` — criar/editar lançamentos visíveis ao household.
   - `viewer` — somente leitura dos lançamentos do household.

   Papéis de finanças são **independentes** de papéis genéricos de grupo/tarefa onde fizer sentido; a API valida permissão no escopo financeiro.

4. **Visibilidade dos lançamentos**
   - **Private** — visível apenas ao `user_id` dono; não entra nos totais compartilhados do household.
   - **Household** — visível aos membros do grupo conforme papel financeiro (`viewer`+).

5. **Frontend**
   - Código do domínio em **`modules/finance`** (rotas, hooks, telas), integrado ao shell do hub sem acoplar a `modules/tasks`.

## Consequências

### Positivas

- Domínio financeiro evolui sem regressão em tarefas.
- Um único lar familiar (grupo) para tarefas e finanças — alinhado ao plano de **hub único**.
- API e schema explícitos para LGPD (escopo de exportação por módulo).

### Negativas / custo

- **Tabelas separadas** (`finance_*` ou equivalente), migrations e repositórios adicionais.
- Duplicação leve de conceitos (membros de grupo) com camada de autorização financeira própria.
- **Export LGPD** do usuário deve ser **estendida** em fase posterior para incluir lançamentos e metadados financeiros (fora do escopo imediato desta ADR).

### Operacional

- Documentar em `docs/compliance/` quando o módulo entrar em produção.
- OpenAPI/Swagger: tag `finance` distinta de `tasks`.

## Referências

- Plano de produto: **hub único** — [docs/PRODUCT_ROADMAP.md](../PRODUCT_ROADMAP.md) (evolução Life Hub / lar familiar).
- Grupos e convites (household): espelho no frontend em `todo-frontend/docs/adr/0002-groups-and-invitations.md`.
- ADR espelhada no frontend: `todo-frontend/docs/adr/0004-finance-bounded-context.md` (API canônica neste repositório).
