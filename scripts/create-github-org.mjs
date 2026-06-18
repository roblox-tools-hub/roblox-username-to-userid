// create-github-org.mjs
// Cria uma GitHub Organization (plano Free) via navegador automatizado.
//
// Por que browser e nao API: o GitHub NAO permite criar org via REST/GraphQL —
// so pela web logada. Este script automatiza esse fluxo e GUARDA a sessao num
// perfil persistente, entao voce loga so UMA vez; nas proximas execucoes ele
// cria a org sem pedir login de novo.
//
// Uso:
//   node create-github-org.mjs <org-name> [billing-email]
//
// Exemplo:
//   node create-github-org.mjs roblox-tools-hub raphaelsantos595@gmail.com
//
// 1a execucao: abre o navegador na tela de login do GitHub -> voce loga (e 2FA
//              se tiver) -> o script detecta o login e segue sozinho.
// Proximas:    reusa a sessao salva em ./.gh-profile e vai direto.
//
// Requer Playwright instalado (este script reaproveita o do BloxStock):
//   node create-github-org.mjs ...   (rodar a partir desta pasta)

import { fileURLToPath } from "url";
import { dirname, join } from "path";
import { createRequire } from "module";

// Resolve o pacote 'playwright' a partir de uma instalacao existente, mesmo que
// este script viva fora de um node_modules. Aponte PLAYWRIGHT_DIR pra qualquer
// projeto que tenha 'playwright' instalado (default: o BloxStock).
const playwrightBase =
  process.env.PLAYWRIGHT_DIR || "C:/ClaudeAIProjects/BloxStock/";
const require = createRequire(playwrightBase.endsWith("/") ? playwrightBase : playwrightBase + "/");
const { chromium } = require("playwright");

const __dirname = dirname(fileURLToPath(import.meta.url));
const PROFILE_DIR = join(__dirname, ".gh-profile"); // sessao persistente (gitignored)

const orgName = process.argv[2];
const billingEmail = process.argv[3] || "";

if (!orgName) {
  console.error("Uso: node create-github-org.mjs <org-name> [billing-email]");
  process.exit(1);
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function ensureLoggedIn(page) {
  await page.goto("https://github.com/", { waitUntil: "domcontentloaded" });
  // Se aparecer botao "Sign in", precisa logar.
  const signedOut = await page.locator('a[href="/login"]').first().isVisible().catch(() => false);
  if (!signedOut) {
    console.log("✅ Sessao do GitHub ja ativa (perfil reutilizado).");
    return;
  }
  console.log("\n🔑 Faca login no GitHub na janela que abriu (usuario + senha + 2FA se houver).");
  console.log("   O script continua sozinho assim que detectar o login...\n");
  await page.goto("https://github.com/login", { waitUntil: "domcontentloaded" });
  // Espera ate o usuario logar (avatar aparece) — ate 5 minutos.
  await page.waitForSelector('summary[aria-label*="View profile"], img.avatar-user', {
    timeout: 5 * 60 * 1000,
  });
  console.log("✅ Login detectado. Sessao salva pra proximas vezes.");
}

async function createOrg(page) {
  console.log(`\n🏗️  Criando a organization "${orgName}" (plano Free)...`);
  await page.goto("https://github.com/account/organizations/new?plan=free", {
    waitUntil: "domcontentloaded",
  });

  // Nome da org. GitHub ja teve varios names de campo; tentamos os comuns.
  const nameInput = page
    .locator('input[name="organization[profile_name]"], input[name="organization[login]"], #organization_profile_name, #organization_login')
    .first();
  await nameInput.waitFor({ timeout: 20000 });
  await nameInput.fill(orgName);

  if (billingEmail) {
    const emailInput = page
      .locator('input[name="organization[billing_email]"], #organization_billing_email')
      .first();
    if (await emailInput.isVisible().catch(() => false)) {
      await emailInput.fill(billingEmail);
    }
  }

  // Tipo de conta: "My personal account" / free — geralmente ja default.
  const personal = page.locator('input[value="personal"], label:has-text("My personal account")').first();
  if (await personal.isVisible().catch(() => false)) {
    await personal.click().catch(() => {});
  }

  await sleep(500);
  // Botao de submit ("Next" ou "Create organization").
  const submit = page
    .locator('button:has-text("Next"), button:has-text("Create organization"), input[type="submit"]')
    .first();
  await submit.click();

  await sleep(3000);
  const url = page.url();
  console.log("➡️  URL apos submit:", url);
  console.log("\n⚠️  Se o GitHub mostrar passos extras (convidar membros, etc.),");
  console.log("    pode clicar em 'Skip'/'Complete setup' — a org JA foi criada.");
  console.log(`\n✅ Org provavelmente criada: https://github.com/${orgName}`);
}

(async () => {
  const ctx = await chromium.launchPersistentContext(PROFILE_DIR, {
    headless: false,
    viewport: { width: 1280, height: 900 },
  });
  const page = ctx.pages()[0] || (await ctx.newPage());
  try {
    await ensureLoggedIn(page);
    await createOrg(page);
    console.log("\n🎉 Pronto. Conferir em: https://github.com/" + orgName);
    console.log("   (Deixe a janela aberta alguns segundos pra confirmar visualmente.)");
    await sleep(8000);
  } catch (e) {
    console.error("❌ Erro:", e.message);
    console.error("   A janela ficou aberta pra voce concluir manualmente se preciso.");
    await sleep(60000);
  } finally {
    await ctx.close();
  }
})();
