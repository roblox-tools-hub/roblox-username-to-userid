# Roblox Username → UserID

Convert Roblox usernames into userIDs using the **official Roblox API** (`users.roblox.com`). No login, no API key, no browser — instant.

## Download

➡️ **[Download username-to-userid.exe](../../releases/latest)** (Windows, standalone — nothing to install)

## How to use

Open the `.exe` and pick one of these:

1. **Paste a list**: run the program, paste usernames (one per line), then press Enter on an empty line.
2. **File**: `username-to-userid.exe < list.txt`
3. **Arguments**: `username-to-userid.exe AceFaaam DashBl0xRBX`

### Output

```
[1] userid -> AceFaaam = 10468392042
[2] userid -> DashBl0xRBX = 10352384284
[3] userid -> Dreamydin0Pixel = 10351452559
```

All usernames are sent in a **single API request** (batches of 100), so rate limits are not an issue for normal use.

### Menu commands

- `all` — copy **all** results to the clipboard
- `1`, `2`, `3`... — copy that line's userID
- `new` — look up another batch of usernames (no need to restart)
- `q` — quit

## Build from source

Requires [Go](https://go.dev/dl/) 1.21+:

```
go build -o username-to-userid.exe .
```
