# Roblox Username → UserID

Converte usernames do Roblox em userIDs usando a **API oficial** da Roblox (`users.roblox.com`). Sem login, sem API key, sem browser — instantâneo.

## Download

➡️ **[Baixar username-to-userid.exe](../../releases/latest)** (Windows, standalone — não precisa instalar nada)

## Como usar

Abra o `.exe` e escolha uma das formas:

1. **Colar lista**: rode o programa, cole os usernames (um por linha) e dê Enter numa linha vazia.
2. **Arquivo**: `username-to-userid.exe < lista.txt`
3. **Argumentos**: `username-to-userid.exe AceFaaam DashBl0xRBX`

### Saída

```
[1] userid -> AceFaaam = 10468392042
[2] userid -> DashBl0xRBX = 10352384284
[3] userid -> Dreamydin0Pixel = 10351452559
```

### Comandos no menu

- `all` — copia **todos** os resultados pro clipboard
- `1`, `2`, `3`... — copia o userID daquela linha
- `q` — sair

## Build a partir do código

Requer [Go](https://go.dev/dl/) 1.21+:

```
go build -o username-to-userid.exe .
```
