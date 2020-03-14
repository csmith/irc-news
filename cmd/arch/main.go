package main

import (
	"encoding/json"
	"fmt"
	"github.com/csmith/ircplugins"
	"log"
	"net/http"
	"net/url"
)

func main() {
	client, err := ircplugins.NewClient()
	if err != nil {
		log.Fatalf("Failed to connec to RPC: %v\n", err)
	}

	defer ignoreError(client.Close())

	err = client.ListenForCommands(map[string]ircplugins.CommandHandler{
		"!arch": searchPackages,
	})
	if err != nil {
		log.Printf("Error listening for commands: %v\n", err)
	}
}

func searchPackages(command ircplugins.Command) {
	respond := func(name string, retriever func(string) ([]Package, error)) {
		packages, err := retriever(command.Arguments)
		if err != nil {
			_ = command.Reply(fmt.Sprintf("Error retrieving packages from %s: %s", name, err.Error()))
			return
		}

		for _, p := range packages {
			_ = command.Reply(fmt.Sprintf("[%s] %s (version %s): %s", name, p.Name, p.Version, p.Description))
		}

		if len(packages) == 0 {
			_ = command.Reply(fmt.Sprintf("[%s] No results found", name))
		}
	}

	go respond("AUR", aurPackages)
	go respond("Main", normalPackages)
}

type Package struct {
	Name        string
	Description string
	Version     string
}

func normalPackages(name string) ([]Package, error) {
	return packages(
		fmt.Sprintf("https://www.archlinux.org/packages/search/json/?q=%s", url.QueryEscape(name)),
		"pkgname",
		"pkgdesc",
		"pkgver",
	)
}

func aurPackages(name string) ([]Package, error) {
	return packages(
		fmt.Sprintf("https://aur.archlinux.org/rpc/?v=5&type=search&arg=%s", url.QueryEscape(name)),
		"Name",
		"Description",
		"Version",
	)
}

func packages(url, nameKey, descriptionKey, versionKey string) (packages []Package, err error) {
	type Result struct {
		Results []map[string]interface{} `json:"results"`
	}

	res, err := http.Get(url)
	if err != nil {
		return
	}

	defer ignoreError(res.Body.Close())

	result := &Result{}
	if err = json.NewDecoder(res.Body).Decode(result); err != nil {
		return
	}

	for i, p := range result.Results {
		if i >= 5 {
			break
		}

		packages = append(packages, Package{
			Name:        p[nameKey].(string),
			Description: p[descriptionKey].(string),
			Version:     p[versionKey].(string),
		})
	}
	return
}

func ignoreError (_ error) {}
