package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
)

type service struct {
	Host string
	FQDN string
	IP   net.IP
	Port uint16
}

const (
	defaultTmpl = `{{range .}}{{range .}}{{.Host}}:{{.Port}}
{{end}}{{end}}`

	envTmpl = `{{range $name, $srvs := .}}{{range $i, $srv := .}}{{$name}}_ADDR{{$i}}={{.Host}}:{{.Port}}
{{$name}}_HOST{{$i}}={{.Host}}
{{$name}}_PORT{{$i}}={{.Port}}
{{end}}{{end}}`
)

func parseQ(q string) (string, string) {
	toks := strings.SplitN(q, "=", 2)
	if len(toks) == 1 {
		return toks[0], toks[0]
	}
	return toks[0], toks[1]
}

func loadTemplate(path string, useEnvTemplate bool) (string, error) {
	if path != "" {
		var raw []byte
		var err error

		if path == "-" {
			raw, err = ioutil.ReadAll(os.Stdin)
		} else {
			raw, err = ioutil.ReadFile(path)
		}

		if err != nil {
			return "", err
		}

		return string(raw), err
	}

	if useEnvTemplate {
		return envTmpl, nil
	}

	return defaultTmpl, nil
}

func main() {
	var templateBody string
	var tpl *template.Template

	var (
		tplPath  = flag.String("t", "", "Path to a template to use. If '-', reads from stdin. Defaults to a built-in template.")
		ns       = flag.String("ns", "", "Nameserver to use. Defaults to the configuration in /etc/resolv.conf.")
		env      = flag.Bool("env", false, "Default to the environment variable output template.")
		jsonMode = flag.Bool("json", false, "Output a JSON representation of the results, instead of rendering a template.")
		relax    = flag.Bool("relax", false, "Treat names that return no results as non-fatal.")
	)

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags ...] [query ... | NAME=query ...]\n", os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\n\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s _api._tcp.internal\n      List hostnames and TCP ports for the `api` service.\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -env ZK=_zk._tcp.service.consul MQ=_mq._tcp.service.consul\n      Look up the `zk` and `mq` services, printing them as environment variables `ZK_ADDR0=host:port`, `ZK_HOST0=host`, ...\n", os.Args[0])
		os.Exit(1)
	}

	if *ns == "" {
		cf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			log.Fatal(err)
		}

		*ns = fmt.Sprintf("%s:%s", cf.Servers[0], cf.Port)
	}

	templateBody, err := loadTemplate(*tplPath, *env)
	if err != nil {
		log.Fatalf("Error loading template: %s", err)
	}

	if !*jsonMode {
		var err error
		if tpl, err = template.New("output").Parse(templateBody); err != nil {
			log.Fatalf("Error parsing template: %s", err)
		}
	}

	records := make(map[string][]service)
	client := &dns.Client{Net: "tcp"}

	for _, q := range flag.Args() {
		alias, domain := parseQ(q)

		msg := &dns.Msg{}
		msg.SetQuestion(dns.Fqdn(domain), dns.TypeSRV)

		in, _, err := client.Exchange(msg, *ns)
		if err != nil {
			log.Fatalf("DNS error: %s", err)
		}

		if !*relax && len(in.Answer) == 0 {
			log.Fatalf("%s: No SRV records returned", domain)
		}

		srvs := make([]service, len(in.Answer))
		for i, answer := range in.Answer {
			if srv, ok := answer.(*dns.SRV); ok {
				host := srv.Target
				if dns.IsFqdn(host) {
					host = host[:len(host)-1]
				}
				srvs[i] = service{Host: host, FQDN: srv.Target, Port: srv.Port}
			}
		}

		for i, extra := range in.Extra {
			if rec, ok := extra.(*dns.A); ok {
				srvs[i].IP = rec.A
			}
		}

		records[alias] = srvs
	}

	if *jsonMode {
		if err := json.NewEncoder(os.Stdout).Encode(records); err != nil {
			log.Fatalf("Error encoding JSON: %s", err)
		}
		return
	}

	if err := tpl.Execute(os.Stdout, records); err != nil {
		log.Fatalf("Error rendering template: %s", err)
	}
}
