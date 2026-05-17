# Procedimento de Resposta a Incidentes de Segurança

**Referência:** LGPD Art. 48  
**Responsável:** [NOME_DPO_OU_EQUIPE] — [EMAIL_PRIVACIDADE]

---

## 1. Definição

**Incidente de segurança:** evento que comprometa ou possa comprometer a confidencialidade, integridade ou disponibilidade de dados pessoais (ex.: vazamento de banco, JWT comprometido, acesso não autorizado).

---

## 2. Fases

### 2.1 Detecção e triagem (0–4 h)

- Registrar data/hora, quem detectou, sistemas afetados.
- Classificar severidade: **Baixa** / **Média** / **Alta** / **Crítica**.
- Acionar equipe técnica e encarregado de dados.

### 2.2 Contenção (4–24 h)

- Revogar credenciais comprometidas (`JWT_SECRET`, tokens Cloudflare, SMTP, Telegram).
- Bloquear IPs ou rotas afetadas.
- Desabilitar funcionalidades se necessário.
- Preservar logs para análise forense.

### 2.3 Erradicação e recuperação

- Corrigir vulnerabilidade.
- Restaurar serviços a partir de backup íntegro se aplicável.
- Validar com checklist em [SECURITY.md](./SECURITY.md).

### 2.4 Comunicação (LGPD Art. 48)

| Prazo | Ação |
|-------|------|
| Até **72 horas úteis** (se risco ou dano relevante) | Comunicar à **ANPD** |
| Em tempo adequado | Comunicar **titulares afetados** com linguagem clara |
| Contínuo | Documentar lições aprendidas |

**Conteúdo mínimo da comunicação:** natureza do incidente, dados afetados, medidas tomadas, contato do encarregado, recomendações aos titulares.

### 2.5 Pós-incidente

- Atualizar ROPA e avaliação de risco.
- Registrar em planilha interna de incidentes.
- Revisar políticas e controles.

---

## 3. Contatos úteis

| Entidade | Contato |
|----------|---------|
| ANPD | https://www.gov.br/anpd |
| Encarregado interno | [EMAIL_PRIVACIDADE] |
| Hospedagem / Cloudflare | Conforme contrato |

---

## 4. Registro de incidentes (template)

| Campo | Valor |
|-------|-------|
| ID do incidente | INC-YYYY-NNN |
| Data detecção | |
| Data contenção | |
| Dados pessoais afetados | |
| Titulares estimados | |
| Comunicação ANPD (S/N) | |
| Comunicação titulares (S/N) | |
| Causa raiz | |
| Ações corretivas | |
