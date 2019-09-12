package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	conf *config
)

type config struct {
	IP struct {
		Type   string `yaml:"type"`
		Config struct {
			URL string `yaml:"url"`
		} `yaml:"config"`
	} `yaml:"ip"`
	DNS struct {
		Type   string `yaml:"type"`
		Config struct {
			Token         string `yaml:"token"`
			ZoneID        string `yaml:"zone_id"`
			DNSRecordName string `yaml:"dns_record_name"`
		} `yaml:"config"`
	} `yaml:"dns"`
}

func main() {
	conf = parseConfig()
	ip := getIP()
	updateDNS(ip)
}

func parseConfig() *config {
	b, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	c := &config{}
	if err := yaml.Unmarshal(b, c); err != nil {
		panic(err)
	}
	return c
}

func getIP() string {
	c := &http.Client{
		Timeout: time.Minute,
	}

	// Contact public IP provider
	resp, err := c.Get(conf.IP.Config.URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Parse IP from the body
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		return scanner.Text()
	}
	panic("cannot parse ip address")
}

func updateDNS(ip string) {
	apiURL := "https://api.cloudflare.com/client/v4"

	c := &http.Client{
		Timeout: time.Minute,
	}

	// Load existing DNS record to be able to update it later
	type resultResp struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	type listResp struct {
		Success bool `json:"success"`
		Result  []resultResp
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/zones/%s/dns_records?type=A&name=%s",
			apiURL, conf.DNS.Config.ZoneID, url.QueryEscape(conf.DNS.Config.DNSRecordName),
		),
		nil,
	)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+conf.DNS.Config.Token)

	r, err := c.Do(req)
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	lr := &listResp{}
	if err := json.Unmarshal(b, lr); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", lr)

	if !lr.Success {
		panic("cannot get dns entry id (invalid credentials?)")
	}

	if len(lr.Result) != 1 {
		panic("cannot get dns entry id (no or multiple records found?)")
	}
	lrr := lr.Result[0]

	// Update the DNS record
	type updateReq struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	type updateResp struct {
		Success bool `json:"success"`
		Result  resultResp
	}

	ureq := &updateReq{
		Type:    lrr.Type,
		Name:    lrr.Name,
		Content: ip,
	}
	js, err := json.Marshal(ureq)
	if err != nil {
		panic(err)
	}

	req, err = http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"%s/zones/%s/dns_records/%s",
			apiURL, conf.DNS.Config.ZoneID, lrr.ID,
		),
		bytes.NewReader(js),
	)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+conf.DNS.Config.Token)

	r, err = c.Do(req)
	if err != nil {
		panic(err)
	}

	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	uresp := &updateResp{}
	if err := json.Unmarshal(b, uresp); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", uresp)

	if !uresp.Success {
		panic("cannot update dns entry")
	}
}
