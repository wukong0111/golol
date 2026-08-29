package items

import "testing"

func TestParseBonusesInfinityEdge(t *testing.T) {
	got := ParseBonuses(`<mainText><stats><attention>75</attention> de daño de ataque<br><attention>25%</attention> de probabilidad de impacto crítico<br><attention>30%</attention> de daño de impacto crítico</stats><br><br></mainText>`)
	want := []struct {
		amount  float64
		percent bool
		name    string
	}{
		{75, false, "de daño de ataque"},
		{25, true, "de probabilidad de impacto crítico"},
		{30, true, "de daño de impacto crítico"},
	}
	if len(got) != len(want) {
		t.Fatalf("len %d: %+v", len(got), got)
	}
	for i, w := range want {
		if got[i].Amount != w.amount || got[i].Percent != w.percent || got[i].Name != w.name {
			t.Fatalf("line %d: %+v", i, got[i])
		}
	}
	if got[0].Rank >= got[1].Rank || got[1].Rank >= got[2].Rank {
		t.Fatalf("tooltip order: ranks %d %d %d", got[0].Rank, got[1].Rank, got[2].Rank)
	}
}

func TestParseBonusesAbilityHasteAndOrnn(t *testing.T) {
	got := ParseBonuses(`<stats><attention>45</attention> de daño de ataque<br><attention>400</attention> de vida<br><attention>20</attention> de velocidad de habilidades</stats>`)
	if len(got) != 3 || got[0].Name != "de daño de ataque" || got[1].Amount != 400 || got[2].Amount != 20 || got[2].Percent {
		t.Fatalf("cleaver: %+v", got)
	}

	ornn := ParseBonuses(`<stats><ornnBonus>40</ornnBonus> de velocidad de habilidades<br><attention>45%</attention> de daño de impacto crítico</stats>`)
	if len(ornn) != 2 || ornn[0].Amount != 45 || !ornn[0].Percent || ornn[1].Amount != 40 || ornn[1].Name != "de velocidad de habilidades" {
		t.Fatalf("ornn: %+v", ornn)
	}
}

func TestParseBonusesSkipsZeroAndEmpty(t *testing.T) {
	if got := ParseBonuses(`<stats><attention>0%</attention> de probabilidad de impacto crítico<br><attention>10%</attention> de velocidad de ataque</stats>`); len(got) != 1 || got[0].Amount != 10 {
		t.Fatalf("zero crit: %+v", got)
	}
	if got := ParseBonuses(`<stats>AS</stats>`); len(got) != 0 {
		t.Fatalf("no number: %+v", got)
	}
	if got := ParseBonuses(`<passive>no stats</passive>`); got != nil {
		t.Fatalf("no block: %+v", got)
	}
}

func TestSumBonusesMergesAndOrders(t *testing.T) {
	a := ParseBonuses(`<stats><attention>15</attention> de armadura<br><attention>10%</attention> de velocidad de ataque</stats>`)
	b := ParseBonuses(`<stats><attention>25</attention> de armadura<br><attention>45</attention> de daño de ataque<br><attention>4%</attention> de velocidad de movimiento</stats>`)
	got := SumBonuses(a, b)
	if len(got) != 4 {
		t.Fatalf("len %d: %+v", len(got), got)
	}
	if got[0].Name != "de daño de ataque" || got[0].Amount != 45 {
		t.Fatalf("ad first: %+v", got[0])
	}
	if got[1].Name != "de armadura" || got[1].Amount != 40 {
		t.Fatalf("armor summed: %+v", got[1])
	}
	if !got[2].Percent || got[2].Amount != 10 {
		t.Fatalf("as: %+v", got[2])
	}
	if !got[3].Percent || got[3].Amount != 4 {
		t.Fatalf("ms: %+v", got[3])
	}
}

func TestSumBonusesKeepsFlatAndPercentApart(t *testing.T) {
	got := SumBonuses(
		ParseBonuses(`<stats><attention>45</attention> de velocidad de movimiento</stats>`),
		ParseBonuses(`<stats><attention>4%</attention> de velocidad de movimiento</stats>`),
	)
	if len(got) != 2 || got[0].Percent || !got[1].Percent {
		t.Fatalf("flat vs percent: %+v", got)
	}
	if got[0].Amount != 45 || got[1].Amount != 4 {
		t.Fatalf("amounts: %+v", got)
	}
}

func TestBonusIndexOmitsEmpty(t *testing.T) {
	idx := BonusIndex([]Item{
		{ID: "1029", Bonuses: ParseBonuses(`<stats><attention>15</attention> de armadura</stats>`)},
		{ID: "0", Name: "none"},
	})
	if len(idx) != 1 || len(idx["1029"]) != 1 || idx["1029"][0].Amount != 15 {
		t.Fatalf("%+v", idx)
	}
}
