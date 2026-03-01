// set-fly-secrets reads .env and runs fly secrets set for TALO_ACCESS_KEY and SESSION_SECRET.
// Handles special characters in .env values (quotes, =, etc.) that break shell parsing.
//
// Usage (from services/vreco or project root):
//   go run ./scripts/set-fly-secrets
//
// Ensure .env exists with TALO_ACCESS_KEY and SESSION_SECRET (32+ chars).
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	cwd, _ := os.Getwd()

	// Load .env from cwd or parent (handles running from services/vreco or project root)
	_ = godotenv.Load(".env", "../.env")

	accessKey := os.Getenv("TALO_ACCESS_KEY")
	sessionSecret := os.Getenv("SESSION_SECRET")

	if accessKey == "" || sessionSecret == "" {
		fmt.Fprintf(os.Stderr, "Error: TALO_ACCESS_KEY and SESSION_SECRET must be set in .env\n")
		os.Exit(1)
	}

	if len(sessionSecret) < 32 {
		fmt.Fprintf(os.Stderr, "Error: SESSION_SECRET must be at least 32 characters\n")
		os.Exit(1)
	}

	// Run fly from directory containing fly.toml
	flyDir := cwd
	if _, err := os.Stat("fly.toml"); os.IsNotExist(err) {
		if _, err := os.Stat("services/vreco/fly.toml"); err == nil {
			flyDir = filepath.Join(cwd, "services", "vreco")
		}
	}
	_ = os.Chdir(flyDir)

	cmd := exec.Command("fly", "secrets", "set",
		"TALO_ACCESS_KEY="+accessKey,
		"SESSION_SECRET="+sessionSecret,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
