package growthref

// Approximate WHO 2006 child growth P3/P50/P97 tables (0–60 months).
// Values are for product reference only, not clinical diagnosis.
// Age is stored as completed months; converted to days as month*30.4375.

type Sex string

const (
	SexBoy  Sex = "boy"
	SexGirl Sex = "girl"
)

type Metric string

const (
	MetricWeight Metric = "weight" // kg
	MetricHeight Metric = "height" // cm
	MetricHead   Metric = "head"   // cm
)

type Point struct {
	AgeDays int     `json:"age_days"`
	P3      float64 `json:"p3"`
	P50     float64 `json:"p50"`
	P97     float64 `json:"p97"`
}

type monthRow struct {
	Month int
	P3    float64
	P50   float64
	P97   float64
}

func ParseSex(gender string) Sex {
	switch gender {
	case "female", "girl", "f", "女":
		return SexGirl
	default:
		return SexBoy
	}
}

func Curves(sex Sex, metric Metric, maxAgeDays int) []Point {
	rows := whoTable(sex, metric)
	if len(rows) == 0 {
		return nil
	}
	if maxAgeDays <= 0 {
		maxAgeDays = 60 * 30
	}
	out := make([]Point, 0, len(rows))
	for _, r := range rows {
		age := int(float64(r.Month)*30.4375 + 0.5)
		if age > maxAgeDays+60 {
			break
		}
		out = append(out, Point{AgeDays: age, P3: r.P3, P50: r.P50, P97: r.P97})
	}
	return out
}

// Percentile estimates percentile from P3/P50/P97 bands via linear interpolation.
func Percentile(sex Sex, metric Metric, ageDays int, value float64) *float64 {
	p3, p50, p97, ok := interpolateBand(sex, metric, ageDays)
	if !ok {
		return nil
	}
	var pct float64
	switch {
	case value <= p3:
		pct = 3
	case value >= p97:
		pct = 97
	case value <= p50:
		if p50 == p3 {
			pct = 50
		} else {
			pct = 3 + (value-p3)/(p50-p3)*47
		}
	default:
		if p97 == p50 {
			pct = 50
		} else {
			pct = 50 + (value-p50)/(p97-p50)*47
		}
	}
	return &pct
}

func interpolateBand(sex Sex, metric Metric, ageDays int) (p3, p50, p97 float64, ok bool) {
	rows := whoTable(sex, metric)
	if len(rows) == 0 {
		return 0, 0, 0, false
	}
	ages := make([]int, len(rows))
	for i, r := range rows {
		ages[i] = int(float64(r.Month)*30.4375 + 0.5)
	}
	if ageDays <= ages[0] {
		r := rows[0]
		return r.P3, r.P50, r.P97, true
	}
	last := rows[len(rows)-1]
	if ageDays >= ages[len(ages)-1] {
		return last.P3, last.P50, last.P97, true
	}
	for i := 1; i < len(rows); i++ {
		if ageDays <= ages[i] {
			a0, a1 := float64(ages[i-1]), float64(ages[i])
			t := (float64(ageDays) - a0) / (a1 - a0)
			r0, r1 := rows[i-1], rows[i]
			return lerp(r0.P3, r1.P3, t), lerp(r0.P50, r1.P50, t), lerp(r0.P97, r1.P97, t), true
		}
	}
	return last.P3, last.P50, last.P97, true
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// WHO 2006 approximations (key months).
func whoTable(sex Sex, metric Metric) []monthRow {
	if sex == SexGirl {
		switch metric {
		case MetricWeight:
			return []monthRow{
				{0, 2.4, 3.2, 4.2}, {1, 3.2, 4.2, 5.5}, {2, 3.9, 5.1, 6.6}, {3, 4.5, 5.8, 7.5},
				{4, 5.0, 6.4, 8.2}, {5, 5.4, 6.9, 8.8}, {6, 5.7, 7.3, 9.3}, {9, 6.5, 8.2, 10.5},
				{12, 7.0, 8.9, 11.5}, {18, 8.1, 10.2, 13.2}, {24, 9.0, 11.5, 14.8},
				{36, 10.8, 13.9, 18.1}, {48, 12.3, 16.1, 21.5}, {60, 13.7, 18.2, 24.9},
			}
		case MetricHeight:
			return []monthRow{
				{0, 45.4, 49.1, 52.9}, {1, 49.8, 53.7, 57.6}, {2, 53.0, 57.1, 61.1}, {3, 55.6, 59.8, 64.0},
				{4, 57.8, 62.1, 66.4}, {5, 59.6, 64.0, 68.5}, {6, 61.2, 65.7, 70.3}, {9, 65.3, 70.1, 74.9},
				{12, 68.9, 74.0, 79.2}, {18, 74.9, 80.7, 86.5}, {24, 80.0, 86.4, 92.9},
				{36, 88.4, 95.1, 101.8}, {48, 94.9, 102.7, 110.6}, {60, 100.7, 109.4, 118.1},
			}
		case MetricHead:
			return []monthRow{
				{0, 31.5, 33.9, 36.2}, {1, 34.2, 36.5, 38.9}, {2, 35.8, 38.3, 40.7}, {3, 37.1, 39.5, 42.0},
				{6, 39.5, 42.2, 44.9}, {9, 41.2, 43.8, 46.5}, {12, 42.4, 45.0, 47.7},
				{18, 44.1, 46.6, 49.2}, {24, 45.1, 47.6, 50.1}, {36, 46.3, 48.8, 51.3},
				{48, 47.1, 49.6, 52.1}, {60, 47.7, 50.2, 52.7},
			}
		}
	}
	// boy
	switch metric {
	case MetricWeight:
		return []monthRow{
			{0, 2.5, 3.3, 4.4}, {1, 3.4, 4.5, 5.8}, {2, 4.3, 5.6, 7.1}, {3, 5.0, 6.4, 8.0},
			{4, 5.6, 7.0, 8.7}, {5, 6.0, 7.5, 9.3}, {6, 6.4, 7.9, 9.8}, {9, 7.1, 8.9, 11.0},
			{12, 7.7, 9.6, 12.0}, {18, 8.8, 10.9, 13.7}, {24, 9.7, 12.2, 15.3},
			{36, 11.3, 14.3, 18.3}, {48, 12.7, 16.3, 21.2}, {60, 14.1, 18.3, 24.2},
		}
	case MetricHeight:
		return []monthRow{
			{0, 46.1, 49.9, 53.7}, {1, 50.8, 54.7, 58.6}, {2, 54.4, 58.4, 62.4}, {3, 57.3, 61.4, 65.5},
			{4, 59.7, 63.9, 68.0}, {5, 61.7, 65.9, 70.1}, {6, 63.3, 67.6, 71.9}, {9, 67.5, 72.0, 76.5},
			{12, 71.0, 75.7, 80.5}, {18, 76.9, 82.3, 87.7}, {24, 81.7, 87.8, 93.9},
			{36, 88.7, 96.1, 103.5}, {48, 94.9, 103.3, 111.7}, {60, 100.7, 110.0, 119.2},
		}
	case MetricHead:
		return []monthRow{
			{0, 32.1, 34.5, 36.9}, {1, 35.1, 37.3, 39.6}, {2, 36.9, 39.1, 41.5}, {3, 38.3, 40.5, 42.9},
			{6, 40.9, 43.3, 45.8}, {9, 42.6, 45.0, 47.4}, {12, 43.8, 46.1, 48.5},
			{18, 45.4, 47.6, 49.9}, {24, 46.3, 48.4, 50.6}, {36, 47.4, 49.5, 51.6},
			{48, 48.1, 50.2, 52.3}, {60, 48.6, 50.7, 52.8},
		}
	}
	return nil
}
