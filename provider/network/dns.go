package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/kopexa-grc/kspec/core"
)

// DNSResource provides access to DNS record information for a domain.
type DNSResource struct{}

// Name returns the resource type name.
func (r *DNSResource) Name() string {
	return "dns"
}

// Fetch retrieves DNS records for the domain specified in the asset.
func (r *DNSResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	domainName := asset.FQDN
	if domainName == "" {
		domainName = asset.Config["domain"]
	}
	if domainName == "" {
		return nil, fmt.Errorf("missing 'domain' or FQDN in asset for dns resource")
	}
	config := asset.Config

	var nameservers []string

	if ns, ok := config["nameserver"]; ok {
		if !strings.Contains(ns, ":") {
			ns = net.JoinHostPort(ns, "53")
		}
		nameservers = []string{ns}
	} else {
		dnsConfig, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err == nil && len(dnsConfig.Servers) > 0 {
			for _, s := range dnsConfig.Servers {
				nameservers = append(nameservers, net.JoinHostPort(s, dnsConfig.Port))
			}
		} else {
			nameservers = []string{"8.8.8.8:53", "1.1.1.1:53"}
		}
	}

	c := new(dns.Client)
	c.Timeout = 5 * time.Second

	var records []map[string]interface{}
	var mxRecords []map[string]interface{}
	// params map: Type -> { "rData": []string, ... }
	params := make(map[string]map[string]interface{})

	// Pre-initialize params for all types to avoid "no such key" errors
	for _, t := range []uint16{dns.TypeA, dns.TypeAAAA, dns.TypeMX, dns.TypeTXT, dns.TypeNS, dns.TypeCNAME, dns.TypePTR} {
		typeName := dns.TypeToString[t]
		params[typeName] = map[string]interface{}{
			"rData": []string{},
		}
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	typesToQuery := []uint16{
		dns.TypeA,
		dns.TypeAAAA,
		dns.TypeMX,
		dns.TypeTXT,
		dns.TypeNS,
		dns.TypeCNAME,
		dns.TypePTR,
	}

	for _, t := range typesToQuery {
		wg.Add(1)
		go func(t uint16) {
			defer wg.Done()

			m := new(dns.Msg)
			m.SetQuestion(dns.Fqdn(domainName), t)
			m.RecursionDesired = true

			var r *dns.Msg
			var err error

			for _, ns := range nameservers {
				if ctx.Err() != nil {
					return
				}
				r, _, err = c.ExchangeContext(ctx, m, ns)
				if err == nil && r != nil {
					break
				}
			}

			if err != nil || r == nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			for _, ans := range r.Answer {
				typeName := dns.TypeToString[ans.Header().Rrtype]

				rec := map[string]interface{}{
					"name":  strings.TrimSuffix(ans.Header().Name, "."),
					"ttl":   ans.Header().Ttl,
					"class": dns.ClassToString[ans.Header().Class],
					"type":  typeName,
				}

				// Params initialization (already done, but ensure safety if type unexpected?)
				if _, ok := params[typeName]; !ok {
					params[typeName] = map[string]interface{}{
						"rData": []string{},
					}
				}

				var dataVal string

				switch v := ans.(type) {
				case *dns.A:
					dataVal = v.A.String()
					rec["data"] = dataVal
				case *dns.AAAA:
					dataVal = v.AAAA.String()
					rec["data"] = dataVal
				case *dns.MX:
					dataVal = strings.TrimSuffix(v.Mx, ".")
					rec["data"] = dataVal
					rec["preference"] = v.Preference

					mxRec := map[string]interface{}{
						"name":       dataVal,
						"preference": v.Preference,
						"ttl":        v.Hdr.Ttl,
					}
					mxRecords = append(mxRecords, mxRec)
				case *dns.TXT:
					dataVal = strings.Join(v.Txt, "")
					rec["data"] = dataVal
					if strings.Contains(dataVal, "v=DKIM1") {
						dkimData := parseDKIM(dataVal)
						rec["dkim"] = dkimData
					}
				case *dns.NS:
					dataVal = strings.TrimSuffix(v.Ns, ".")
					rec["data"] = dataVal
				case *dns.CNAME:
					dataVal = strings.TrimSuffix(v.Target, ".")
					rec["data"] = dataVal
				case *dns.PTR:
					dataVal = strings.TrimSuffix(v.Ptr, ".")
					rec["data"] = dataVal
				default:
					dataVal = ""
				}

				// Append to params rData
				if pData, ok := params[typeName]["rData"].([]string); ok {
					params[typeName]["rData"] = append(pData, dataVal)
				}

				records = append(records, rec)
			}
		}(t)
	}

	wg.Wait()

	resource := core.Resource{
		"name":    domainName,
		"records": records,
		"mx":      mxRecords,
		"params":  params,
	}

	return []core.Resource{resource}, nil
}

func parseDKIM(txt string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(txt, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			result[key] = val
		}
	}
	return result
}
