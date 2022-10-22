package main

import "time"

func MoveFromTo(v *VyosAPI, from *Uplink, to *Uplink) {
	if to.IsDown() {
		LOG.Printf("nat %s is down (current %d < zero-level %d) - do nothing\n", to.Nat, to.Current, to.Lz)
		ZeroCounts(from, to)
		return
	}

	LOG.Printf("get rules with nat %s...\n", from.Nat)
	rules := v.GetNatRules()
	LOG.Println(len(rules[from.Nat]), "rules found")

	if len(rules[from.Nat]) == 0 {
		ZeroCounts(from, to)
		return
	}

	//LOG.Println("get nat top rule ...")
	ruleTop := v.GetNatRulesTop(rules[from.Nat])
	LOG.Printf("top rule # %d\n", ruleTop)
	if ruleTop == 0 {
		LOG.Println("ERROR: top rule number is zero")
		ZeroCounts(from, to)
		return
	}

	LOG.Printf("move rule # %d to %s\n", ruleTop, to.Nat)
	v.SetRuleTarget(ruleTop, to.Nat)
	ZeroCounts(from, to)
}

func main() {
	LOG.Println("-- start --")
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
		CalcLimits(Beeline, TTK)
		LOG.Printf("Beeline [%d:%d, %d:%d]: %5d, TTK [%d:%d, %d:%d]: %5d\n",
			Beeline.L0.Limit, Beeline.L0.OverflowCount, Beeline.L1.Limit, Beeline.L1.OverflowCount, Beeline.Current,
			TTK.L0.Limit, TTK.L0.OverflowCount, TTK.L1.Limit, TTK.L1.OverflowCount, TTK.Current)

		if Beeline.L0.Overflowed && !TTK.L1.Overflowed {
			LOG.Println("Beeline -> TTK")
			MoveFromTo(v, Beeline, TTK)
		}

		if (TTK.L0.Overflowed && !Beeline.L0.Overflowed) || (TTK.L1.Overflowed && !Beeline.L1.Overflowed) {
			LOG.Println("TTK -> Beeline")
			MoveFromTo(v, TTK, Beeline)
		}

		//LOG.Printf("waiting %v ...\n", CFG.CheckPeriod.Duration)
		time.Sleep(CFG.CheckPeriod.Duration - time.Second)
	}
}
