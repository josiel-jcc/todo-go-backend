# Visão de produto — Módulo de finanças

**Última atualização:** 2026-05-29

---

## Reposicionamento

O Todo evolui de um **app de tarefas** para um **hub de casa e vida organizada**: tarefas e finanças no mesmo lugar, com a mesma identidade visual e os mesmos grupos familiares.

| Antes | Depois |
|-------|--------|
| Foco em listas e lembretes | **Tarefas + finanças** como pilares do dia a dia |
| Grupos só para colaborar em tarefas | Grupos como **unidade da casa** (tarefas e finanças compartilhadas) |
| Finanças inexistentes ou externas | Controle manual simples, sem depender de banco ou Open Finance no MVP |

**Proposta de valor:** “Organize o que precisa ser feito e quanto entra/sai — sozinho ou com quem mora com você.”

---

## Família via grupos existentes

Não criamos um conceito novo de “família”. Reutilizamos **`groups`** e **`group_members`** já usados em tarefas:

- Cada grupo representa um **lar ou núcleo** (ex.: “Os Silva”, “Apartamento 302”).
- Membros convidados pelo fluxo atual de convites veem finanças **da casa** conforme papel e visibilidade.
- Dados financeiros ficam vinculados a `group_id`; transações podem ser **privadas** (só o autor) ou **da casa** (visíveis a quem tem permissão no grupo).

Papéis financeiros por grupo (MVP): **admin**, **editor**, **viewer** — ver [FINANCE_MVP_SPEC.md](./FINANCE_MVP_SPEC.md).

---

## Escopo por versão

### MVP — controle manual essencial

- Contas (corrente, poupança, carteira, etc.)
- Categorias de receita e despesa
- Lançamentos: **receita**, **despesa** e **transferência** entre contas
- Visibilidade **privada** vs **casa** por transação
- Dashboard mensal: saldo por conta, totais por categoria, resumo receitas/despesas
- API REST sob `/api/v1/finance/groups/{groupId}/…`
- Papéis admin / editor / viewer no grupo

**Fora do MVP:** Open Finance, cartões, parcelamentos, orçamentos, metas, relatórios avançados, exportação contábil.

### V2 — crédito e automação leve

- Cartões de crédito e **parcelamentos** (fatura, parcela N de M)
- Conciliação manual aprimorada (match sugerido)
- Orçamento mensal por categoria
- Notificações (“fatura vence em 3 dias”)
- **Ainda sem** Open Finance obrigatório

### V3 — integração e inteligência

- **Open Finance** (importação de extratos / saldos com consentimento)
- Regras automáticas (categorizar por descrição, recorrências financeiras)
- Metas de economia e projeção de fluxo de caixa
- Relatórios exportáveis (CSV/PDF) e visão anual comparativa

---

## Princípios de produto

1. **Simples antes de completo** — MVP deve ser útil com lançamento manual em menos de um minuto.
2. **Privacidade explícita** — gasto pessoal pode ficar privado; gasto da casa é opt-in por lançamento.
3. **Mesma casa, mesma equipe** — quem já está no grupo de tarefas não precisa de outro cadastro.
4. **Backend primeiro** — especificação e API estáveis antes de UI rica no frontend.

---

## Documentos relacionados

| Documento | Conteúdo |
|-----------|----------|
| [FINANCE_MVP_SPEC.md](./FINANCE_MVP_SPEC.md) | Entidades, API, banco, critérios de aceite |
| [../PRODUCT_ROADMAP.md](../PRODUCT_ROADMAP.md) | Roadmap geral do produto |
| [../compliance/](../compliance/) | LGPD e retenção de dados |
