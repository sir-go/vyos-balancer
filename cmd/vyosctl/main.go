package main

import (
	"flag"
	"time"
)

// MoveFromTo moves the NAT rule with the biggest amount of bytes from one Uplink to the other
func MoveFromTo(v *VyosAPI, from *Uplink, to *Uplink) error {
	if to.IsDown() {
		LOG.Printf("nat %s is down (current %d < zero-level %d) - do nothing\n", to.Nat, to.Current, to.Lz)
		ZeroCounts(from, to)
		return nil
	}

	LOG.Printf("get rules with nat %s...\n", from.Nat)
	rules, err := v.GetNatRules()
	if err != nil {
		return err
	}
	LOG.Println(len(rules[from.Nat]), "rules found")

	if len(rules[from.Nat]) == 0 {
		ZeroCounts(from, to)
		return nil
	}

	ruleTop, err := v.GetNatRulesTop(rules[from.Nat])
	if err != nil {
		return err
	}

	LOG.Printf("top rule # %d\n", ruleTop)
	if ruleTop == 0 {
		LOG.Println("top rule number is zero")
		ZeroCounts(from, to)
		return nil
	}

	LOG.Printf("move rule # %d to %s\n", ruleTop, to.Nat)
	if err = v.SetRuleTarget(ruleTop, to.Nat); err != nil {
		return err
	}
	ZeroCounts(from, to)

	return nil
}

func main() {
	LOG.Println("-- start --")

	fCfgPath := flag.String("c", "config.toml", "path to conf file")
	flag.Parse()
	CFG = LoadConfig(*fCfgPath)
	LOG.Println("conf: ", CFG.Path)

	LOG.Println("VyOS url: ", CFG.Vyos.Addr)
	v := &VyosAPI{Addr: CFG.Vyos.Addr, Key: CFG.Vyos.Key}

	LOG.Println("All bw meters are in Mbps")

	Beeline = CFG.Beeline
	TTK = CFG.TTK

	LOG.Printf("run checks loop with period: %v ...\n", CFG.CheckPeriod.Duration)
	for {
		//LOG.Println("get bw ...")
		GetDevBW(Beeline, TTK)
		CalcLimits(CFG.OverflowCount, Beeline, TTK)
		LOG.Printf("Beeline [%d:%d, %d:%d]: %5d, TTK [%d:%d, %d:%d]: %5d\n",
			Beeline.L0.Limit, Beeline.L0.OverflowCount, Beeline.L1.Limit, Beeline.L1.OverflowCount, Beeline.Current,
			TTK.L0.Limit, TTK.L0.OverflowCount, TTK.L1.Limit, TTK.L1.OverflowCount, TTK.Current)

		if Beeline.L0.Overflowed && !TTK.L1.Overflowed {
			LOG.Println("Beeline -> TTK")
			if err := MoveFromTo(v, Beeline, TTK); err != nil {
				LOG.Panic(err)
			}
		}

		if (TTK.L0.Overflowed && !Beeline.L0.Overflowed) || (TTK.L1.Overflowed && !Beeline.L1.Overflowed) {
			LOG.Println("TTK -> Beeline")
			if err := MoveFromTo(v, TTK, Beeline); err != nil {
				LOG.Panic(err)
			}
		}

		//LOG.Printf("waiting %v ...\n", CFG.CheckPeriod.Duration)
		time.Sleep(CFG.CheckPeriod.Duration - time.Second)
	}
}
