// username-to-userid: converte usernames do Roblox em userIDs.
//
// Usa a API publica oficial da Roblox (sem API key, sem browser):
//   POST https://users.roblox.com/v1/usernames/users
//
// Uso:
//   1) Cole os usernames (um por linha), termine com uma linha vazia ou Ctrl+Z+Enter (Win) / Ctrl+D (Unix)
//   2) Ou passe por arquivo:   go run . < lista.txt
//   3) Ou como argumentos:     go run . AceFaaam DashBl0xRBX
//
// Depois do resultado, um menu permite:
//   - "all"  -> copia TODOS os resultados pro clipboard
//   - numero -> copia so o userid daquela linha
//   - "q"    -> sair
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const apiURL = "https://users.roblox.com/v1/usernames/users"

type apiResponse struct {
	Data []struct {
		RequestedUsername string `json:"requestedUsername"`
		ID                int64  `json:"id"`
		Name              string `json:"name"`
	} `json:"data"`
}

// resultado por username, na ordem de entrada.
type result struct {
	Username string
	UserID   int64 // 0 = nao encontrado
}

func main() {
	usernames := readUsernames()
	if len(usernames) == 0 {
		fmt.Println("Nenhum username informado.")
		return
	}

	results, err := lookup(usernames)
	if err != nil {
		fmt.Println("Erro na consulta:", err)
		os.Exit(1)
	}

	printResults(results)
	interactiveMenu(results)
}

// readUsernames: pega de argumentos OU do stdin (uma por linha, para na linha vazia).
func readUsernames() []string {
	if len(os.Args) > 1 {
		return cleanList(os.Args[1:])
	}
	fmt.Println("Cole os usernames (um por linha). Linha vazia para finalizar:")
	var raw []string
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			break
		}
		raw = append(raw, line)
	}
	return cleanList(raw)
}

func cleanList(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, u := range in {
		u = strings.TrimSpace(u)
		if u == "" || seen[strings.ToLower(u)] {
			continue
		}
		seen[strings.ToLower(u)] = true
		out = append(out, u)
	}
	return out
}

// lookup: consulta a API em lotes de 100 (limite da API) e preserva a ordem de entrada.
func lookup(usernames []string) ([]result, error) {
	found := map[string]int64{} // chave = lower(name)
	client := &http.Client{Timeout: 15 * time.Second}

	for start := 0; start < len(usernames); start += 100 {
		end := start + 100
		if end > len(usernames) {
			end = len(usernames)
		}
		batch := usernames[start:end]

		body, _ := json.Marshal(map[string]any{
			"usernames":          batch,
			"excludeBannedUsers": false,
		})
		req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		var ar apiResponse
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&ar)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}
		for _, d := range ar.Data {
			// a API casa por nome real (case-insensitive); guarda pelo username pedido tambem
			found[strings.ToLower(d.RequestedUsername)] = d.ID
			found[strings.ToLower(d.Name)] = d.ID
		}
	}

	out := make([]result, 0, len(usernames))
	for _, u := range usernames {
		out = append(out, result{Username: u, UserID: found[strings.ToLower(u)]})
	}
	return out, nil
}

func printResults(results []result) {
	fmt.Println()
	for i, r := range results {
		val := "NAO ENCONTRADO"
		if r.UserID != 0 {
			val = strconv.FormatInt(r.UserID, 10)
		}
		fmt.Printf("[%d] userid -> %s = %s\n", i+1, r.Username, val)
	}
	fmt.Println()
}

func interactiveMenu(results []result) {
	fmt.Println("Comandos: [all] copiar todos | [numero] copiar 1 | [q] sair")
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			return
		}
		cmd := strings.TrimSpace(strings.ToLower(sc.Text()))
		switch {
		case cmd == "q" || cmd == "":
			return
		case cmd == "all":
			var b strings.Builder
			for _, r := range results {
				if r.UserID != 0 {
					fmt.Fprintf(&b, "userid -> %s = %d\n", r.Username, r.UserID)
				} else {
					fmt.Fprintf(&b, "userid -> %s = NAO ENCONTRADO\n", r.Username)
				}
			}
			copyClipboard(b.String())
			fmt.Println("Copiado: todos os resultados.")
		default:
			n, err := strconv.Atoi(cmd)
			if err != nil || n < 1 || n > len(results) {
				fmt.Println("Comando invalido.")
				continue
			}
			r := results[n-1]
			if r.UserID == 0 {
				fmt.Println("Esse username nao foi encontrado.")
				continue
			}
			copyClipboard(strconv.FormatInt(r.UserID, 10))
			fmt.Printf("Copiado: %d (%s)\n", r.UserID, r.Username)
		}
	}
}

// copyClipboard: usa o utilitario nativo de cada SO (sem dependencias externas).
func copyClipboard(s string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-NoProfile", "-Command", "$input | Set-Clipboard")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	cmd.Stdin = strings.NewReader(s)
	if err := cmd.Run(); err != nil {
		fmt.Println("(aviso: nao consegui copiar pro clipboard:", err, ")")
	}
}
