package network

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/network/resources"
)

func TestDNSFetch(t *testing.T) {
	// Start a local mock DNS server
	server := &dns.Server{Addr: "127.0.0.1:0", Net: "udp"}

	dns.HandleFunc("example.com.", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		if len(r.Question) == 0 {
			w.WriteMsg(m)
			return
		}

		switch r.Question[0].Qtype {
		case dns.TypeA:
			rr, _ := dns.NewRR("example.com. 3600 IN A 1.2.3.4")
			m.Answer = append(m.Answer, rr)
		case dns.TypeMX:
			rr, _ := dns.NewRR("example.com. 3600 IN MX 10 mail.example.com.")
			m.Answer = append(m.Answer, rr)
		case dns.TypeTXT:
			rr, _ := dns.NewRR("example.com. 3600 IN TXT \"v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQKBAQUAA4GNADCBiQKBgQD\"")
			m.Answer = append(m.Answer, rr)
		}
		w.WriteMsg(m)
	})

	pc, err := net.ListenPacket("udp", "127.0.0.1:0") //nolint:noctx // Test code
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	server.PacketConn = pc
	addr := pc.LocalAddr().String()

	go func() {
		server.ActivateAndServe()
	}()
	defer server.Shutdown()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test Fetch
	r := resources.NewDNS()
	config := map[string]string{
		"domain":     "example.com",
		"nameserver": addr,
	}
	asset := core.Asset{
		FQDN:   "example.com",
		Config: config,
	}

	dnsResources, err := r.Fetch(context.Background(), asset)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(dnsResources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(dnsResources))
	}

	res := dnsResources[0]
	records := res["records"].([]map[string]interface{})

	// Check A Record
	foundA := false
	for _, rec := range records {
		if rec["type"] == "A" {
			foundA = true
			if rec["data"] != "1.2.3.4" {
				t.Errorf("Expected A record 1.2.3.4, got %v", rec["data"])
			}
		}
	}
	if !foundA {
		t.Error("Did not find A record")
	}

	// Check MX Record
	foundMX := false
	for _, rec := range records {
		if rec["type"] == "MX" {
			foundMX = true
			if rec["data"] != "mail.example.com" {
				t.Errorf("Expected MX mail.example.com, got %v", rec["data"])
			}
			if rec["preference"] != uint16(10) {
				t.Errorf("Expected MX preference 10, got %v", rec["preference"])
			}
		}
	}
	if !foundMX {
		t.Error("Did not find MX record")
	}

	// Check DKIM (TXT)
	foundDKIM := false
	for _, rec := range records {
		if rec["type"] == "TXT" {
			if dkim, ok := rec["dkim"].(map[string]string); ok {
				foundDKIM = true
				if dkim["v"] != "DKIM1" {
					t.Errorf("Expected DKIM v=DKIM1, got %v", dkim["v"])
				}
			}
		}
	}
	if !foundDKIM {
		t.Error("Did not find DKIM record or failed to parse")
	}
	// Check Params
	if params, ok := res["params"].(map[string]map[string]interface{}); ok {
		if txtParams, ok := params["TXT"]; ok {
			if rData, ok := txtParams["rData"].([]string); ok {
				foundSPF := false
				for _, d := range rData {
					if d == "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQKBAQUAA4GNADCBiQKBgQD" {
						foundSPF = true // referencing the DKIM record which is consistent with the mock
					}
				}
				if !foundSPF {
					t.Errorf("Expected params['TXT']['rData'] to contain the TXT record")
				}
			} else {
				t.Errorf("Expected params['TXT']['rData'] to be []string")
			}
		} else {
			t.Errorf("Expected params['TXT'] to exist")
		}
	} else {
		t.Errorf("Expected 'params' to be map[string]map[string]interface{}, got mixed or missing")
	}
}

// TestDNS_Name tests the resource name.
func TestDNS_Name(t *testing.T) {
	r := resources.NewDNS()
	if got := r.Name(); got != "dns" {
		t.Errorf("DNS.Name() = %q, want %q", got, "dns")
	}
}

// TestDNS_Fetch_MissingDomain tests error handling for missing domain.
func TestDNS_Fetch_MissingDomain(t *testing.T) {
	r := resources.NewDNS()
	asset := core.Asset{
		Config: map[string]string{},
	}

	_, err := r.Fetch(context.Background(), asset)
	if err == nil {
		t.Error("Expected error for missing domain")
	}
}
