package network

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"

	"github.com/kopexa-grc/kspec/core"
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
	r := &DNSResource{}
	config := map[string]string{
		"domain":     "example.com",
		"nameserver": addr,
	}
	asset := core.Asset{
		FQDN:   "example.com",
		Config: config,
	}

	resources, err := r.Fetch(context.Background(), asset)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("Expected 1 resource, got %d", len(resources))
	}

	res := resources[0]
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

func TestParseDKIM(t *testing.T) {
	txt := "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQKBAQUAA4GNADCBiQKBgQD"
	parsed := parseDKIM(txt)

	if parsed["v"] != "DKIM1" {
		t.Errorf("Expected v=DKIM1, got %s", parsed["v"])
	}
	if parsed["k"] != "rsa" {
		t.Errorf("Expected k=rsa, got %s", parsed["k"])
	}
	if len(parsed["p"]) < 10 {
		t.Errorf("Expected valid p value, got %s", parsed["p"])
	}
}

// FuzzParseDKIM tests DKIM parsing with random TXT record input.
func FuzzParseDKIM(f *testing.F) {
	// Seed corpus with various DKIM-like strings
	seeds := []string{
		"v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQKBAQUAA4GNADCBiQKBgQD",
		"v=DKIM1; k=rsa; p=base64data; t=s; n=notes",
		"",
		"no-semicolons",
		";;;",
		"key=value",
		"key=",
		"=value",
		"key=value;",
		";key=value",
		"k1=v1;k2=v2;k3=v3",
		"  k1 = v1 ;  k2 = v2  ",
		"key=value=more",
		"v=DKIM1",
		"k=rsa; p=data",
		string(make([]byte, 1000)),
		"key=\x00\x01\x02",
		"日本語=テスト",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, txt string) {
		// Should never panic
		result := parseDKIM(txt)

		// Result should always be a valid map
		if result == nil {
			t.Error("parseDKIM() returned nil map")
		}
	})
}

// TestParseDKIM_EdgeCases tests edge cases in DKIM parsing.
func TestParseDKIM_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want map[string]string
	}{
		{
			name: "empty string",
			txt:  "",
			want: map[string]string{},
		},
		{
			name: "only semicolons",
			txt:  ";;;",
			want: map[string]string{},
		},
		{
			name: "no equals sign",
			txt:  "just-text",
			want: map[string]string{},
		},
		{
			name: "empty key",
			txt:  "=value",
			want: map[string]string{"": "value"},
		},
		{
			name: "empty value",
			txt:  "key=",
			want: map[string]string{"key": ""},
		},
		{
			name: "whitespace handling",
			txt:  "  k  =  v  ;  k2  =  v2  ",
			want: map[string]string{"k": "v", "k2": "v2"},
		},
		{
			name: "multiple equals in value",
			txt:  "key=value=more=stuff",
			want: map[string]string{"key": "value=more=stuff"},
		},
		{
			name: "duplicate keys (last wins)",
			txt:  "k=first; k=second",
			want: map[string]string{"k": "second"},
		},
		{
			name: "unicode values",
			txt:  "日本語=テスト",
			want: map[string]string{"日本語": "テスト"},
		},
		{
			name: "leading/trailing semicolons",
			txt:  ";k=v;",
			want: map[string]string{"k": "v"},
		},
		{
			name: "single key-value",
			txt:  "v=DKIM1",
			want: map[string]string{"v": "DKIM1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDKIM(tt.txt)
			if len(got) != len(tt.want) {
				t.Errorf("parseDKIM(%q) returned %d keys, want %d", tt.txt, len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseDKIM(%q)[%q] = %q, want %q", tt.txt, k, got[k], v)
				}
			}
		})
	}
}

// TestDNSResource_Name tests the resource name.
func TestDNSResource_Name(t *testing.T) {
	r := &DNSResource{}
	if got := r.Name(); got != "dns" {
		t.Errorf("DNSResource.Name() = %q, want %q", got, "dns")
	}
}

// TestDNSResource_Fetch_MissingDomain tests error handling for missing domain.
func TestDNSResource_Fetch_MissingDomain(t *testing.T) {
	r := &DNSResource{}
	asset := core.Asset{
		Config: map[string]string{},
	}

	_, err := r.Fetch(context.Background(), asset)
	if err == nil {
		t.Error("Expected error for missing domain")
	}
}
