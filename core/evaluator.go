package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
)

type Evaluator struct {
	Env      *cel.Env
	Registry map[string]ResourceSpec // Access to providers
}

func NewEvaluator(registry map[string]ResourceSpec) (*Evaluator, error) {
	// We register a custom 'dns' function.
	// It takes a string (domain) and returns a map (resource).

	// Function declaration
	dnsFuncOpts := []cel.EnvOption{
		cel.Function("dns",
			cel.Overload("dns_string",
				[]*cel.Type{cel.StringType},
				cel.DynType, // Returns a dynamic map/object
				cel.UnaryBinding(func(arg ref.Val) ref.Val {
					// This is executed at runtime.
					domainStr, ok := arg.Value().(string)
					if !ok {
						return types.NewErr("dns() argument must be a string")
					}

					if registry == nil {
						return types.NewErr("registry not available")
					}

					dnsProvider, ok := registry["dns"]
					if !ok {
						return types.NewErr("dns provider not found")
					}

					// Fetch!
					ctx := context.TODO() // We don't have ctx here.
					res, err := dnsProvider.Fetch(ctx, Asset{FQDN: domainStr})
					if err != nil {
						return types.NewErr("dns fetch failed: %s", err.Error())
					}
					if len(res) == 0 {
						return types.NullValue
					}

					// Return the first resource as a Map
					return types.NewDynamicMap(types.DefaultTypeAdapter, res[0])
				}),
			),
		),
		cel.Function("reverseIP",
			cel.Overload("reverseIP_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(arg ref.Val) ref.Val {
					ipStr, ok := arg.Value().(string)
					if !ok {
						return types.NewErr("reverseIP() argument must be a string")
					}
					parts := strings.Split(ipStr, ".")
					if len(parts) != 4 {
						return types.NewErr("invalid IPv4 address: %s", ipStr)
					}
					// Reverse
					return types.String(fmt.Sprintf("%s.%s.%s.%s.in-addr.arpa", parts[3], parts[2], parts[1], parts[0]))
				}),
			),
		),
	}

	// Default env options + generic extension
	envOpts := append([]cel.EnvOption{
		cel.Variable("resource", cel.MapType(cel.StringType, cel.DynType)),  // The resource being verified
		cel.Variable("props", cel.MapType(cel.StringType, cel.DynType)),     // Props map injected
		cel.Variable("config", cel.MapType(cel.StringType, cel.StringType)), // Config map injected
		cel.Variable("asset", cel.MapType(cel.StringType, cel.DynType)),     // Asset context injected
		cel.Variable("empty", cel.NullType),                                 // Empty constant
		cel.Variable("regex", cel.MapType(cel.StringType, cel.StringType)),  // Regex patterns injected
		ext.Strings(),  // String extensions (contains, split, etc.)
		ext.Lists(),    // List extensions (filter, map, slice, etc.)
		ext.Encoders(), // Encoding extensions (base64, etc.)
	}, dnsFuncOpts...)

	env, err := cel.NewEnv(envOpts...)
	if err != nil {
		return nil, err
	}

	return &Evaluator{Env: env, Registry: registry}, nil
}

func (e *Evaluator) Evaluate(expression string, resource Resource, config map[string]string, props map[string]interface{}, asset Asset) (bool, error) {
	ast, issues := e.Env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("compile error: %s", issues.Err())
	}

	prg, err := e.Env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("program construction error: %s", err)
	}

	// Activation
	input := map[string]interface{}{
		"resource": resource,
		"config":   config,
		"props":    props,
		"asset": map[string]interface{}{
			"id":     asset.ID,
			"name":   asset.Name,
			"fqdn":   asset.FQDN,
			"type":   asset.Type,
			"config": asset.Config,
		},
		"empty": nil,
		"regex": map[string]interface{}{
			"ipv4": "^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$",
			"ipv6": "^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$",
		},
	}

	out, _, err := prg.Eval(input)
	if err != nil {
		return false, fmt.Errorf("evaluation error: %s", err)
	}

	if out == types.True {
		return true, nil
	}
	if out == types.False {
		return false, nil
	}

	// If it returns something else, try boolean conversion or error
	if b, ok := out.Value().(bool); ok {
		return b, nil
	}

	return false, fmt.Errorf("expression did not return a boolean, got: %v", out.Value())
}
