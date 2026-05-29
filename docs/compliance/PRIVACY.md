# Política de Privacidade

**Última atualização:** 2026-05-29

Esta Política de Privacidade descreve como **[RAZAO_SOCIAL]** (CNPJ **[CNPJ]**, endereço: **[ENDERECO]**) — o **Controlador** — trata dados pessoais no aplicativo de tarefas (Todo App), em conformidade com a Lei Geral de Proteção de Dados (LGPD — Lei nº 13.709/2018).

**Contato do encarregado / privacidade:** [EMAIL_PRIVACIDADE]

> **Fonte canônica:** este arquivo é exibido no app (`/privacidade`). Cópias em outros diretórios do monorepo devem ser mantidas em sincronia com `todo-frontend/docs/compliance/`.

---

## 1. Dados pessoais coletados

| Categoria | Exemplos | Origem |
|-----------|----------|--------|
| Identificação | nome de usuário (username), e-mail | Cadastro |
| Autenticação | senha (armazenada apenas em hash bcrypt) | Cadastro |
| Preferências | notificações habilitadas, lembrete padrão (minutos), ID de chat Telegram | Configurações |
| Web Push | endpoint da subscription, chaves de criptografia (`p256dh`, `auth`), user-agent opcional | Opt-in ao ativar push no navegador |
| Grupos e colaboração | nome do grupo, participação, convites pendentes/aceitos, notificações in-app (ex.: convite de grupo, lembrete de tarefa) | Uso do app |
| Conteúdo | títulos, descrições, comentários, tags, datas e prioridades de tarefas | Uso do app |
| Financeiros (dados sensíveis) | contas (`financial_accounts`), categorias (`finance_categories`), lançamentos (`finance_transactions`: valor, descrição, data, visibilidade) | Módulo finanças do lar — **planejado / MVP** |
| Técnicos | endereço IP, logs de acesso (via infraestrutura) | Uso automático |

Não coletamos intencionalmente dados de crianças. O serviço é destinado a maiores de 18 anos.

---

## 2. Finalidades e bases legais (Art. 7, LGPD)

| Finalidade | Base legal |
|------------|------------|
| Criar e autenticar sua conta | Execução de contrato (Art. 7, V) |
| Gerenciar tarefas, tags e comentários | Execução de contrato |
| Atribuir e compartilhar tarefas com outros usuários do mesmo grupo | Execução de contrato |
| Organizar grupos, enviar convites e exibir notificações in-app relacionadas | Execução de contrato |
| Enviar lembretes por e-mail ou Telegram | Consentimento (Art. 7, I) — opt-in nas configurações |
| Enviar lembretes por Web Push no navegador/PWA | Consentimento (Art. 7, I) — permissão do navegador + ativação nas configurações |
| Exibir lembretes no sino in-app do aplicativo | Execução de contrato (entrega do serviço de lembretes solicitado) |
| Segurança, prevenção a fraudes e logs | Legítimo interesse (Art. 7, IX), com balanceamento |
| Cumprir obrigações legais | Obrigação legal (Art. 7, II) |
| Organizar orçamento e finanças do lar familiar (contas, categorias, lançamentos) | Execução de contrato (Art. 7, V) — módulo **planejado / MVP** |

---

## 3. Compartilhamento com operadores (terceiros)

| Operador | Finalidade | Dados compartilhados | Localização |
|----------|------------|----------------------|-------------|
| Provedor de hospedagem / banco de dados | Armazenamento | Todos os dados da conta e conteúdo | [PAIS_HOSPEDAGEM] |
| Provedor SMTP | E-mails de notificação | E-mail, conteúdo da tarefa | Conforme contrato do provedor |
| Telegram (API Bot) | Lembretes por mensagem | Chat ID, título/descrição de tarefas | Possível transferência internacional |
| Navegador / serviço de push (FCM, Mozilla, etc.) | Notificações Web Push no PWA | Endpoint e metadados da subscription; conteúdo do lembrete no payload | Conforme política do provedor de push |
| Cloudflare | Proxy, túnel, proteção | Tráfego, IP, metadados de requisição | Conforme política Cloudflare |

Contratos ou cláusulas contratuais padrão são utilizados quando aplicável (Art. 33–36, LGPD).

---

## 4. Grupos e colaboração

Tarefas podem ser atribuídas ou compartilhadas apenas com usuários que compartilham um **grupo** com você. Para convidar alguém, usamos o ID e o username (e-mail não é exibido na listagem padrão de colegas). Convites exigem **aceite** do convidado; até lá, mantemos o status do convite e notificações in-app associadas. Novos usuários podem ser incluídos automaticamente em um grupo padrão configurado pelo controlador (ex.: uso familiar).

---

## 5. Dados financeiros do lar (planejado / MVP)

O módulo de finanças do hub familiar trata **dados pessoais sensíveis** (Art. 5, II, LGPD), pois podem revelar situação econômica e padrão de consumo do titular e do lar.

| Aspecto | Descrição |
|---------|-----------|
| **Finalidade** | Gestão de orçamento doméstico: cadastro de contas, categorias e lançamentos (receitas/despesas) vinculados ao **grupo familiar** (`group_id` / household). |
| **Dados tratados** | Nome e tipo de conta; categorias do household; valor, descrição, data e categoria de cada lançamento; visibilidade (`private` ou `household`); papéis financeiros por membro (`admin`, `editor`, `viewer`). |
| **Base legal** | Execução de contrato (Art. 7, V) — funcionalidade solicitada no âmbito do hub familiar. |
| **Compartilhamento entre titulares** | Lançamentos com visibilidade **household** são acessíveis aos membros do mesmo grupo, conforme o **papel financeiro** (leitura ou edição). Lançamentos **private** permanecem visíveis apenas ao titular que os criou e não entram nos totais compartilhados do lar. |
| **Status** | Tabelas e API em **planejamento (MVP)**; ver [ADR 0001](../adr/0001-finance-bounded-context.md). Política de retenção em [DATA-RETENTION.md](./DATA-RETENTION.md). |

A exportação LGPD (`GET /users/me/export`) será **estendida** em fase posterior para incluir metadados e lançamentos financeiros do titular, quando o módulo estiver em produção.

---

## 6. Retenção

Consulte [DATA-RETENTION.md](./DATA-RETENTION.md). Em resumo:

- Dados da conta: enquanto a conta estiver ativa.
- Após solicitação de exclusão: eliminação ou anonimização em até 30 dias, salvo obrigação legal de retenção.
- Logs técnicos: até 90 dias, salvo incidente de segurança.

---

## 7. Seus direitos (Art. 18, LGPD)

Você pode, mediante requisição a [EMAIL_PRIVACIDADE] ou pelo app:

- **Confirmar** existência de tratamento e **acessar** seus dados (exportação em Configurações).
- **Corrigir** dados incompletos ou desatualizados.
- **Anonimizar, bloquear ou eliminar** dados desnecessários (**excluir conta** em Configurações).
- **Portabilidade** dos dados a outro fornecedor (exportação JSON).
- **Revogar consentimento** (desativar notificações; remover Telegram Chat ID; desativar Web Push nas configurações ou no navegador).
- **Opor-se** a tratamentos baseados em legítimo interesse, quando aplicável.

Responderemos em prazo razoável, conforme a LGPD.

---

## 8. Web Push — armazenamento de endpoints

Quando você ativa notificações push no navegador, o aplicativo envia ao nosso backend o **endpoint** fornecido pelo serviço de push do navegador (por exemplo, FCM ou Mozilla), as chaves de criptografia da subscription (`p256dh` e `auth`) e, opcionalmente, o identificador do navegador (`user_agent`). Esses dados são vinculados à sua conta para entregar lembretes de tarefas apenas a dispositivos que você registrou. Você pode revogar o consentimento desativando push nas configurações do app (que remove a subscription no servidor) ou bloqueando notificações nas configurações do sistema operacional/navegador. Endpoints inválidos ou expirados são excluídos automaticamente.

---

## 9. Segurança

Medidas incluem: senhas com hash bcrypt, autenticação JWT em cookie HttpOnly, HTTPS, controle de acesso por usuário e por grupo nas tarefas, rate limiting em rotas de autenticação e configuração restritiva em produção. Detalhes em [SECURITY.md](./SECURITY.md).

---

## 10. Transferência internacional

O uso do Telegram, de serviços de push do navegador e de serviços em nuvem pode implicar transferência de dados para outros países. Garantimos mecanismos adequados conforme Arts. 33–36 da LGPD.

---

## 11. Alterações

Esta política pode ser atualizada. A data da última versão constará no topo. Alterações relevantes serão comunicadas no app ou por e-mail.

---

## 12. Autoridade nacional

Você pode apresentar reclamação à Autoridade Nacional de Proteção de Dados (ANPD): https://www.gov.br/anpd
