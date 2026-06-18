// username-to-userid: converts Roblox usernames into userIDs.
//
// Uses the official public Roblox API (no API key, no browser):
//   POST https://users.roblox.com/v1/usernames/users
//
// Usage:
//   1) Paste usernames (one per line), finish with an empty line or Ctrl+Z+Enter (Win) / Ctrl+D (Unix)
//   2) Or via file:        username-to-userid.exe < list.txt
//   3) Or as arguments:    username-to-userid.exe AceFaaam DashBl0xRBX
//
// After the results, a menu lets you:
//   - "all"   -> copy ALL results to the clipboard
//   - number  -> copy just that line's userID
//   - "q"     -> quit
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

// result per username, in input order.
type result struct {
	Username string
	UserID   int64 // 0 = not found
}

// stdin is a single shared scanner. Multiple bufio.Scanner on os.Stdin would
// lose buffered lines to read-ahead, so the whole program reuses this one.
var stdin = bufio.NewScanner(os.Stdin)

func main() {
	usernames := readUsernames()
	if len(usernames) == 0 {
		fmt.Println("No usernames provided.")
		return
	}

	for {
		results, err := lookup(usernames)
		if err != nil {
			fmt.Println("Lookup error:", err)
			os.Exit(1)
		}

		printResults(results)
		if !interactiveMenu(results) {
			return // user quit
		}
		// user chose "new": read another batch of usernames and loop
		usernames = promptUsernames()
		if len(usernames) == 0 {
			return
		}
	}
}

// promptUsernames: reads a fresh batch of usernames from stdin (one per line).
func promptUsernames() []string {
	fmt.Println("\nPaste new usernames (one per line). Empty line to finish:")
	return readLines()
}

// readUsernames: reads from arguments OR stdin (one per line, stops on empty line).
func readUsernames() []string {
	if len(os.Args) > 1 {
		return cleanList(os.Args[1:])
	}
	fmt.Println("Paste usernames (one per line). Empty line to finish:")
	return readLines()
}

// readLines: reads lines from the shared stdin scanner until an empty line/EOF.
func readLines() []string {
	var raw []string
	for stdin.Scan() {
		line := strings.TrimSpace(stdin.Text())
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

// lookup: queries the API in batches of 100 (API limit) and preserves input order.
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
			// the API matches by real name (case-insensitive); also store by requested username
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
		val := "NOT FOUND"
		if r.UserID != 0 {
			val = strconv.FormatInt(r.UserID, 10)
		}
		fmt.Printf("[%d] userid -> %s = %s\n", i+1, r.Username, val)
	}
	fmt.Println()
}

// interactiveMenu: returns true if the user wants to enter a new batch of usernames,
// false if they want to quit.
func interactiveMenu(results []result) bool {
	fmt.Println("Commands: [all] copy all | [number] copy one | [new] add more usernames | [q] quit")
	for {
		fmt.Print("> ")
		if !stdin.Scan() {
			return false
		}
		cmd := strings.TrimSpace(strings.ToLower(stdin.Text()))
		switch {
		case cmd == "":
			continue
		case cmd == "q":
			return false
		case cmd == "new" || cmd == "n":
			return true // signal caller to read a new batch
		case cmd == "all":
			var b strings.Builder
			for _, r := range results {
				if r.UserID != 0 {
					fmt.Fprintf(&b, "userid -> %s = %d\n", r.Username, r.UserID)
				} else {
					fmt.Fprintf(&b, "userid -> %s = NOT FOUND\n", r.Username)
				}
			}
			copyClipboard(b.String())
			fmt.Println("Copied: all results.")
		default:
			n, err := strconv.Atoi(cmd)
			if err != nil || n < 1 || n > len(results) {
				fmt.Println("Invalid command.")
				continue
			}
			r := results[n-1]
			if r.UserID == 0 {
				fmt.Println("That username was not found.")
				continue
			}
			copyClipboard(strconv.FormatInt(r.UserID, 10))
			fmt.Printf("Copied: %d (%s)\n", r.UserID, r.Username)
		}
	}
}

// copyClipboard: uses each OS's native utility (no external dependencies).
func copyClipboard(s string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("clip.exe")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	cmd.Stdin = strings.NewReader(s)
	if err := cmd.Run(); err != nil {
		fmt.Println("(warning: could not copy to clipboard:", err, ")")
	}
}
