# scripts/

## create-github-org.mjs

Cria uma **GitHub Organization** (plano Free) automatizando a web — o GitHub não
permite criar org via API. A sessão fica salva num perfil persistente
(`.gh-profile/`, gitignored), então você loga **uma vez** e nas próximas execuções
ele cria a org sem pedir login.

### Uso

```
node create-github-org.mjs <org-name> [billing-email]
```

Exemplo (o que foi usado pra criar `roblox-tools-hub`):

```
node create-github-org.mjs roblox-tools-hub raphaelsantos595@gmail.com
```

- **1ª vez**: abre o navegador na tela de login → você loga (com 2FA se tiver) →
  o script segue sozinho. Se aparecer CAPTCHA, resolva na janela.
- **Próximas**: reusa a sessão de `.gh-profile/`.

### Dependência

Reaproveita o Playwright já instalado no BloxStock. Se estiver noutra máquina/projeto,
aponte:

```
PLAYWRIGHT_DIR="C:/caminho/pro/projeto-com-playwright/" node create-github-org.mjs <org>
```

### Depois de criar a org

Transferir um repo pra ela (sem expor seu usuário pessoal no link):

```
gh api -X POST repos/<voce>/<repo>/transfer -f new_owner=<org>
git remote set-url origin https://github.com/<org>/<repo>.git
```
