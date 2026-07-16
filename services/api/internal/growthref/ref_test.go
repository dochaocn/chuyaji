package growthref

import "testing"

func TestPercentileMid(t *testing.T) {
	p := Percentile(SexBoy, MetricWeight, 30, 4.5)
	if p == nil {
		t.Fatal("expected percentile")
	}
	if *p < 3 || *p > 97 {
		t.Fatalf("percentile out of range: %v", *p)
	}
}

func TestCurvesNonEmpty(t *testing.T) {
	c := Curves(SexGirl, MetricHeight, 365)
	if len(c) < 3 {
		t.Fatalf("expected curve points, got %d", len(c))
	}
}
