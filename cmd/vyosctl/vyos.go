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
	NatDynRuleRe  = regexp.MustCompile(`(?ms)^(\d+).+to\s+(\d{0,3}\.\d{0,3}\.\d{0,3}\.\d{0,3}/\d+).*DYNAMIC$`)
	NatRuleStatRe = regexp.MustCompile(`(?m)^(\d+)\s+\d+\w?\s+(\d+)(\w)?.*$`)
)

type VyosResp struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Error   string `json:"error"`
}

type VyosAPI struct {
	Addr string
	Key  string
}

func (vapi *VyosAPI) GetNatRules() (rules map[string][]int) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	res, err := client.PostForm(vapi.Addr+"/show", url.Values{
		"key":  []string{vapi.Key},
		"data": []string{`{"cmd": "nat source rules"}`},
	})
	eh(err)

	resp := new(VyosResp)
	eh(json.NewDecoder(res.Body).Decode(&resp))
	if !resp.Success {
		eh(errors.New("vyos resp isn't success"), resp.Error)
	}

	rules = make(map[string][]int)

	dataStrings := strings.Split(resp.Data, "\n\n")[1:]
	for _, s := range dataStrings {
		match := NatDynRuleRe.FindStringSubmatch(s)
		if match == nil {
			continue
		}
		ruleNum, err := strconv.Atoi(match[1])
		eh(err)

		if _, ok := rules[match[2]]; !ok {
			rules[match[2]] = []int{ruleNum}
		} else {
			rules[match[2]] = append(rules[match[2]], ruleNum)
		}
	}
	return
}

func foundInArrInt(s *[]int, v int) bool {
	for _, i := range *s {
		if i == v {
			return true
		}
	}
	return false
}

func parseKMGTuint(digits string, multiplier string) uint64 {
	d, err := strconv.ParseUint(digits, 10, 64)
	eh(err)
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
	return d
}

func (vapi *VyosAPI) GetNatRulesTop(rules []int) (maxBytesRuleNum int) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	res, err := client.PostForm(vapi.Addr+"/show", url.Values{
		"key":  []string{vapi.Key},
		"data": []string{`{"cmd": "nat source statistics"}`},
	})
	eh(err)

	resp := new(VyosResp)
	eh(json.NewDecoder(res.Body).Decode(&resp))
	if !resp.Success {
		eh(errors.New("vyos resp isn't success"), resp.Error)
	}

	var maxBytes uint64 = 0
	dataStrings := strings.Split(resp.Data, "\n")[2:]
	for _, s := range dataStrings {
		match := NatRuleStatRe.FindStringSubmatch(s)
		if match == nil {
			continue
		}

		ruleNum, err := strconv.Atoi(match[1])
		eh(err)

		if !foundInArrInt(&rules, ruleNum) {
			continue
		}

		bytes := parseKMGTuint(match[2], match[3])

		if bytes > maxBytes || maxBytes == 0 {
			maxBytes = bytes
			maxBytesRuleNum = ruleNum
		}
	}
	return
}

func (vapi *VyosAPI) SetRuleTarget(ruleNum int, target string) {
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	data := fmt.Sprintf(
		`{"op": "set", "path": ["nat", "source", "rule", "%d", "translation", "address"], "value": "%s"}`,
		ruleNum, target)

	LOG.Println(data)

	res, err := client.PostForm(vapi.Addr+"/configure", url.Values{"key": []string{vapi.Key}, "data": []string{data}})
	eh(err)

	resp := new(VyosResp)
	eh(json.NewDecoder(res.Body).Decode(&resp))
	if !resp.Success {
		eh(errors.New("vyos resp isn't success"), resp.Error)
	}
}
