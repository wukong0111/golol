package champions

import (
	"testing"

	"golol/internal/ddragon"
)

func TestApplyKitsFillsScaling(t *testing.T) {
	cat := sampleRoster(t)
	raw := []byte(`{
		"Aatrox":{
			"key":"Aatrox",
			"abilities":{
				"P":[{"name":"Deathbringer Stance","cooldown":{"modifiers":[{"values":[22,10],"units":["",""]}]},
					"effects":[{"description":"deal bonus magic damage equal to 4% : 8% (based on level) of the target's maximum health"}]}],
				"Q":[{
					"name":"The Darkin Blade",
					"effects":[{
						"leveling":[{
							"attribute":"Physical Damage",
							"modifiers":[
								{"values":[10,25,40,55,70],"units":["","","","",""]},
								{"values":[60,67.5,75,82.5,90],"units":["% AD","% AD","% AD","% AD","% AD"]}
							]
						}]
					}]
				}],
				"W":[{
					"name":"Infernal Chains",
					"effects":[{
						"leveling":[{
							"attribute":"Slow",
							"modifiers":[{"values":[25,25,25,25,25],"units":["%","%","%","%","%"]}]
						}]
					}]
				}]
			}
		}
	}`)
	if err := cat.ApplyKits(raw); err != nil {
		t.Fatal(err)
	}
	aatrox, ok := cat.Get("Aatrox")
	if !ok {
		t.Fatal("missing Aatrox")
	}
	if aatrox.Passive.Cooldown != "22 / 10" {
		t.Fatalf("passive cd: %q", aatrox.Passive.Cooldown)
	}
	if len(aatrox.Passive.Scaling) != 1 {
		t.Fatalf("passive scaling: %+v", aatrox.Passive.Scaling)
	}
	if aatrox.Passive.Scaling[0].Label != "Daño mágico" {
		t.Fatalf("passive label: %q", aatrox.Passive.Scaling[0].Label)
	}
	if aatrox.Passive.Scaling[0].Text != "4% – 8% de la vida máxima del objetivo" {
		t.Fatalf("passive text: %q", aatrox.Passive.Scaling[0].Text)
	}
	if len(aatrox.Spells) == 0 || len(aatrox.Spells[0].Scaling) != 1 {
		t.Fatalf("q scaling: %+v", aatrox.Spells)
	}
	row := aatrox.Spells[0].Scaling[0]
	if row.Label != "Daño físico" {
		t.Fatalf("label: %q", row.Label)
	}
	if row.Text != "10 / 25 / 40 / 55 / 70 (+ 60 / 67.5 / 75 / 82.5 / 90% DA)" {
		t.Fatalf("text: %q", row.Text)
	}

	ahri, ok := cat.Get("Ahri")
	if !ok {
		t.Fatal("missing Ahri")
	}
	if len(ahri.Spells) > 0 && len(ahri.Spells[0].Scaling) != 0 {
		t.Fatalf("ahri should stay without kits: %+v", ahri.Spells[0].Scaling)
	}
}

func TestApplyKitsMultipleForms(t *testing.T) {
	rawRoster := []byte(`{
		"type":"champion","version":"16.16.1",
		"data":{"Gnar":{"id":"Gnar","name":"Gnar","tags":["Fighter"],"image":{"full":"Gnar.png"},
			"spells":[{"id":"GnarQ","name":"Bumerán / Peñascazo","description":"q","image":{"full":"GnarQ.png"}}]}}
	}`)
	cat, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, rawRoster)
	if err != nil {
		t.Fatal(err)
	}
	if err := cat.ApplyKits([]byte(`{
		"Gnar":{"key":"Gnar","abilities":{"Q":[
			{"name":"Boomerang Throw","effects":[{"leveling":[{"attribute":"Physical Damage","modifiers":[{"values":[5,15],"units":["",""]}]}]}]},
			{"name":"Boulder Toss","effects":[{"leveling":[{"attribute":"Physical Damage","modifiers":[{"values":[20,40],"units":["",""]}]}]}]}
		]}}
	}`)); err != nil {
		t.Fatal(err)
	}
	gnar, _ := cat.Get("Gnar")
	if len(gnar.Spells[0].Forms) != 2 {
		t.Fatalf("forms: %+v", gnar.Spells[0].Forms)
	}
	if gnar.Spells[0].Forms[0].Name != "Boomerang Throw" || gnar.Spells[0].Forms[1].Name != "Boulder Toss" {
		t.Fatalf("form names: %+v", gnar.Spells[0].Forms)
	}
	if len(gnar.Spells[0].Scaling) != 0 {
		t.Fatal("single-form scaling should stay empty when there are forms")
	}
}

func TestExtractLevelScales(t *testing.T) {
	ahri := extractLevelScales("she consumes them to heal herself for 35 : 95 (based on level) (+ 20% AP). takedown heal herself for 75 : 165 (based on level) (+ 30% AP)")
	if len(ahri) != 2 {
		t.Fatalf("ahri rows: %+v", ahri)
	}
	if ahri[0].Label != "Curación" || ahri[0].Text != "35 – 95 (+ 20% PH)" {
		t.Fatalf("ahri 0: %+v", ahri[0])
	}
	if ahri[1].Text != "75 – 165 (+ 30% PH)" {
		t.Fatalf("ahri 1: %+v", ahri[1])
	}

	teemo := extractLevelScales("he gains 20% / 40% / 60% / 80% (based on level) bonus attack speed for 5 seconds")
	if len(teemo) != 1 || teemo[0].Label != "Velocidad de ataque adicional" {
		t.Fatalf("teemo: %+v", teemo)
	}
	if teemo[0].Text != "20% / 40% / 60% / 80%" {
		t.Fatalf("teemo text: %q", teemo[0].Text)
	}

	darius := extractLevelScales("dealt[ 13 : 30 (based on level) (+ 30% bonus AD) total physical damage over the duration, ][ 3.25 : 7.5 (based on level) (+ 7.5% bonus AD) physical damage every 1.25 seconds")
	if len(darius) != 2 {
		t.Fatalf("darius: %+v", darius)
	}
	if darius[0].Label != "Daño físico" || darius[0].Text != "13 – 30 (+ 30% DA adicional)" {
		t.Fatalf("darius 0: %+v", darius[0])
	}
}

func TestExtractGwenWFormulas(t *testing.T) {
	desc := "Active: Gwen summons the Hallowed Mist upon her current location, lasting for 4 seconds. While inside the mist, Gwen becomes ghosted, gains 22 (+ 7% AP) bonus armor and bonus magic resistance and is untargetable."
	rows := extractDescriptionScales(desc)
	got := map[string]string{}
	for _, row := range rows {
		got[row.Label] = row.Text
	}
	if got["Resistencias"] != "22 (+ 7% PH)" {
		t.Fatalf("resists: %+v", rows)
	}
	if got["Duración"] != "4 s" {
		t.Fatalf("duration: %+v", rows)
	}
	ahri := extractDescriptionScales("heal herself for 35 : 95 (based on level) (+ 20% AP)")
	if len(ahri) != 1 || ahri[0].Text != "35 – 95 (+ 20% PH)" {
		t.Fatalf("should not duplicate formula: %+v", ahri)
	}
}

func TestFormatValues(t *testing.T) {
	if got := formatValues([]float64{40, 40, 40, 40, 40}); got != "40" {
		t.Fatalf("equal: %q", got)
	}
	if got := formatValues([]float64{22, 21.29, 20, 10, 9, 8, 7, 6, 5, 4, 3, 2}); got != "22 – 2" {
		t.Fatalf("long: %q", got)
	}
	if got := formatNum(67.5); got != "67.5" {
		t.Fatalf("num: %q", got)
	}
}

func TestTranslateAttr(t *testing.T) {
	if got := translateAttr("First Cast Damage"); got != "Daño (1.er lanzamiento)" {
		t.Fatalf("got %q", got)
	}
	if got := translateAttr("First Sweetspot Damage"); got != "Daño de punto dulce (1.er lanzamiento)" {
		t.Fatalf("sweetspot: %q", got)
	}
	if got := translateAttr("Physical Damage"); got != "Daño físico" {
		t.Fatalf("exact: %q", got)
	}
}

func TestApplyKitsBadJSON(t *testing.T) {
	cat := sampleRoster(t)
	if err := cat.ApplyKits([]byte(`{`)); err == nil {
		t.Fatal("expected error")
	}
}
