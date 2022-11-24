package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	// NatDynRuleRe  - regexp for matching dynamic NAT rules
	NatDynRuleRe = regexp.MustCompile(`(?ms)^(\d+).+to\s+(\d{0,3}\.\d{0,3}\.\d{0,3}\.\d{0,3}/\d+).*DYNAMIC$`)

	// NatRuleStatRe  - regexp for matching the rule state
	NatRuleStatRe = regexp.MustCompile(`(?m)^(\d+)\s+\d+\w?\s+(\d+)(\w)?.*$`)
)

// VyosResp stores VyOS API response
type VyosResp struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Error   string `json:"error"`
}

// VyosAPI client
type VyosAPI struct {
	Addr string
	Key  string
}

// GetNatRules fetches all NAT rules via VyOS API
func (api *VyosAPI) GetNatRules() (rules map[string][]int, err error) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	res, err := client.PostForm(api.Addr+"/show", url.Values{
		"key":  []string{api.Key},
		"data": []string{`{"cmd": "nat source rules"}`},
	})
	if err != nil {
		return nil, err
	}

	resp := new(VyosResp)
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, errors.New("vyos resp isn't success, " + resp.Error)
	}

	rules = make(map[string][]int)

	dataStrings := strings.Split(resp.Data, "\n\n")[1:]
	for _, s := range dataStrings {
		match := NatDynRuleRe.FindStringSubmatch(s)
		if match == nil {
			continue
		}
		ruleNum, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, err
		}

		if _, ok := rules[match[2]]; !ok {
			rules[match[2]] = []int{ruleNum}
		} else {
			rules[match[2]] = append(rules[match[2]], ruleNum)
		}
	}
	return
}

// foundInArrInt show is an Int value in the array
func foundInArrInt(s []int, v int) bool {
	for _, i := range s {
		if i == v {
			return true
		}
	}
	return false
}

// parseKMGTUint converts a string to uint64 considering a multiplier mark
func parseKMGTUint(digits string, multiplier string) (uint64, error) {
	d, err := strconv.ParseUint(digits, 10, 64)
	if err != nil {
		return 0, err
	}
	switch strings.ToLower(multiplier) {
	case "k":
		d <<= 10
	case "m":
		d <<= 20
	case "g":
		d <<= 30
	case "t":
		d <<= 40
	}
	return d, nil
}

// GetNatRulesTop returns an ID of given
func (api *VyosAPI) GetNatRulesTop(rules []int) (maxBytesRuleNum int, err error) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	res, err := client.PostForm(api.Addr+"/show", url.Values{
		"key":  []string{api.Key},
		"data": []string{`{"cmd": "nat source statistics"}`},
	})
	if err != nil {
		return 0, err
	}

	resp := new(VyosResp)
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return 0, err
	}
	if !resp.Success {
		return 0, errors.New("vyos resp isn't success, " + resp.Error)
	}

	var maxBytes uint64 = 0
	dataStrings := strings.Split(resp.Data, "\n")[2:]
	for _, s := range dataStrings {
		match := NatRuleStatRe.FindStringSubmatch(s)
		if match == nil {
			continue
		}

		ruleNum, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, err
		}

		if !foundInArrInt(rules, ruleNum) {
			continue
		}

		bytes, err := parseKMGTUint(match[2], match[3])
		if err != nil {
			return 0, err
		}

		if bytes > maxBytes || maxBytes == 0 {
			maxBytes = bytes
			maxBytesRuleNum = ruleNum
		}
	}
	return maxBytesRuleNum, nil
}

// SetRuleTarget changes a target address for certain NAT rule
func (api *VyosAPI) SetRuleTarget(ruleNum int, target string) error {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	data := fmt.Sprintf(
		`{"op": "set", "path": ["nat", "source", "rule", "%d", "translation", "address"], "value": "%s"}`,
		ruleNum, target)

	LOG.Println(data)

	res, err := client.PostForm(api.Addr+"/configure", url.Values{"key": []string{api.Key}, "data": []string{data}})
	if err != nil {
		return err
	}

	resp := new(VyosResp)
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return err
	}
	if !resp.Success {
		return errors.New("vyos resp isn't success, " + resp.Error)
	}
	return nil
}
