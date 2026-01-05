package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/juliankoehn/kspec/core"
	"github.com/juliankoehn/kspec/provider"
)

func main() {
	// cnquery run "query" --resource file --config path=.
	resourceType := flag.String("resource", "", "Resource type (e.g., file, github_repo)")
	configStr := flag.String("config", "", "Config key=value pairs comma separated (e.g. path=./,owner=foo)")

	flag.Parse()
	query := flag.Arg(0)

	if *resourceType == "" || query == "" {
		log.Fatal("Usage: cnquery run <query> --resource <type> --config <k=v>")
	}

	// Parse config
	config := make(map[string]string)
	if *configStr != "" {
		pairs := strings.Split(*configStr, ",")
		for _, p := range pairs {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				config[kv[0]] = kv[1]
			}
		}
	}

	// Initialize Registry with ALL providers
	// Pass config to providers (e.g. if one of the k=v was 'token')
	registry, err := provider.InitProviders(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to init providers: %v", err)
	}

	spec, ok := registry[*resourceType]
	if !ok {
		log.Fatalf("Unknown resource type: %s", *resourceType)
	}

	ctx := context.Background()

	// Create Asset
	asset := core.Asset{
		Type:   *resourceType,
		Config: config,
	}
	// Fallback/heuristic: if config has domain, set FQDN
	if d, ok := config["domain"]; ok {
		asset.FQDN = d
	}
	if p, ok := config["path"]; ok {
		asset.Name = p
	}

	resources, err := spec.Fetch(ctx, asset)
	if err != nil {
		log.Fatalf("Failed to fetch: %v", err)
	}

	eval, err := core.NewEvaluator(registry)
	if err != nil {
		log.Fatalf("Failed to init evaluator: %v", err)
	}

	for _, res := range resources {
		passed, err := eval.Evaluate(query, res, config, nil, asset)

		result := map[string]interface{}{
			"resource": res,
			"passed":   passed,
			"error":    nil,
		}
		if err != nil {
			result["error"] = err.Error()
		}

		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
	}
}
