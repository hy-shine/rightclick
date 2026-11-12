// Command updatepack writes a signed update archive and manifest for one app
// target. The release script calls it after the Finder extension is embedded.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hy-shine/rightclick/internal/updatepack"
)

func main() {
	log.SetFlags(0)
	var (
		app        = flag.String("app", "", "path to the .app bundle")
		name       = flag.String("name", "RightClick", "archive name prefix")
		version    = flag.String("version", "", "version in the manifest")
		target     = flag.String("target", "", "update target, such as darwin-arm64")
		baseURL    = flag.String("base-url", "", "HTTPS directory that serves the archive")
		notes      = flag.String("notes", "", "release notes in Markdown")
		notesFile  = flag.String("notes-file", "", "Markdown changelog whose version section is used")
		appex      = flag.String("appex", "RightClickFinderSync.appex", "Finder extension bundle that must be present")
		privateKey = flag.String("private-key", "", "file with the base64 Ed25519 private key, or omit to use MYGO_UPDATER_PRIVATE_KEY")
		publicKey  = flag.String("public-key", "", "file with the base64 Ed25519 public key, or omit to use MYGO_UPDATER_PUBLIC_KEY")
		out        = flag.String("out", ".", "directory for the archive and manifest")
		date       = flag.String("date", "", "manifest date in RFC 3339; default now")
	)
	flag.Parse()
	if *app == "" || *version == "" || *target == "" || *baseURL == "" {
		flag.Usage()
		os.Exit(2)
	}
	privateText, err := readKeyFileOrEnv(*privateKey, "MYGO_UPDATER_PRIVATE_KEY")
	if err != nil {
		log.Fatal(err)
	}
	publicText, err := readKeyFileOrEnv(*publicKey, "MYGO_UPDATER_PUBLIC_KEY")
	if err != nil {
		log.Fatal(err)
	}
	priv, err := updatepack.ParsePrivateKey(privateText)
	if err != nil {
		log.Fatal(err)
	}
	pub, err := updatepack.ParsePublicKey(publicText)
	if err != nil {
		log.Fatal(err)
	}
	notesText := *notes
	if *notesFile != "" {
		data, err := os.ReadFile(*notesFile)
		if err != nil {
			log.Fatal(err)
		}
		notesText = updatepack.ReleaseNotes(string(data), *version)
		if notesText == "" {
			log.Fatalf("updatepack: %s has no notes for %s", *notesFile, *version)
		}
	}
	var manifestDate time.Time
	if *date != "" {
		manifestDate, err = time.Parse(time.RFC3339, *date)
		if err != nil {
			log.Fatalf("updatepack: invalid date: %v", err)
		}
	}
	result, err := updatepack.Create(updatepack.Options{
		App:        *app,
		Name:       *name,
		Version:    *version,
		Target:     *target,
		BaseURL:    *baseURL,
		Notes:      notesText,
		Appex:      *appex,
		PrivateKey: priv,
		PublicKey:  pub,
		OutDir:     *out,
		Date:       manifestDate,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("archive=%s\nmanifest=%s\nsize=%d\n", result.Archive, result.Manifest, result.Size)
}

func readKeyFileOrEnv(path, env string) (string, error) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	if value := os.Getenv(env); value != "" {
		return value, nil
	}
	return "", fmt.Errorf("updatepack: set %s or pass a key file", env)
}
