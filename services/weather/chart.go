package weather

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	windScaleMax      = 120.0
	humidityMaxPct    = 100.0
	windShowThreshold = 30.0

	chartFontAxis    = 15.0
	chartFontHour    = 16.0
	chartFontPoint   = 17.0
	chartFontCurrent = 16.0
	chartFontSun     = 15.0
)

type ChartOptions struct {
	// SparseLabels shows temp values only on first, last, hottest and coldest points.
	SparseLabels bool
	// TempOnly draws only the temperature series (no wind/humidity).
	TempOnly bool
}

func IsRainCode(code int) bool {
	switch {
	case code >= 51 && code <= 67:
		return true
	case code >= 80 && code <= 82:
		return true
	case code >= 95:
		return true
	default:
		return false
	}
}

func hasRainInHourly(hourly []HourlyPoint) bool {
	for _, h := range hourly {
		if IsRainCode(h.WeatherCode) {
			return true
		}
	}
	return false
}

func hasStrongWind(hourly []HourlyPoint) bool {
	for _, h := range hourly {
		if h.WindSpeedKmh > windShowThreshold {
			return true
		}
	}
	return false
}

func sparseLabelIndexes(temps []float64) map[int]bool {
	out := map[int]bool{}
	if len(temps) == 0 {
		return out
	}
	out[0] = true
	out[len(temps)-1] = true

	minIdx, maxIdx := 0, 0
	for i, v := range temps {
		if v < temps[minIdx] {
			minIdx = i
		}
		if v > temps[maxIdx] {
			maxIdx = i
		}
	}
	if minIdx != 0 && minIdx != len(temps)-1 {
		out[minIdx] = true
	}
	if maxIdx != 0 && maxIdx != len(temps)-1 {
		out[maxIdx] = true
	}
	return out
}

func BuildHourlyChart(hourly []HourlyPoint, sun SunResponse, now time.Time, loc *time.Location, current CurrentResponse) string {
	return BuildHourlyChartWithOptions(hourly, sun, now, loc, current, ChartOptions{})
}

func BuildHourlyChartWithOptions(hourly []HourlyPoint, sun SunResponse, now time.Time, loc *time.Location, current CurrentResponse, opts ChartOptions) string {
	if len(hourly) == 0 {
		return ""
	}

	showHumidity := !opts.TempOnly && (IsRainCode(current.WeatherCode) || hasRainInHourly(hourly))
	showWind := !opts.TempOnly && (current.WindSpeedKmh > windShowThreshold || hasStrongWind(hourly))

	const width = 800.0
	height := 190.0
	if opts.TempOnly {
		height = 150.0
	}
	const padL = 12.0
	padR := 12.0
	if showWind || showHumidity {
		padR = 48.0
	}
	const padT = 18.0
	const padB = 28.0
	chartW := width - padL - padR
	chartH := height - padT - padB
	xLabelY := padT + chartH + 16

	start, _ := time.Parse(time.RFC3339, hourly[0].Time)
	end, _ := time.Parse(time.RFC3339, hourly[len(hourly)-1].Time)
	if end.Before(start) {
		end = start.Add(11 * time.Hour)
	}
	duration := end.Sub(start)
	if duration <= 0 {
		duration = time.Hour
	}

	var temps, winds, humids []float64
	for _, h := range hourly {
		temps = append(temps, h.TemperatureC)
		winds = append(winds, h.WindSpeedKmh)
		humids = append(humids, float64(h.HumidityPct))
	}

	tempMin, tempMax := paddedTempRange(temps)
	labelIdx := map[int]bool{}
	if opts.SparseLabels {
		labelIdx = sparseLabelIndexes(temps)
	}

	xFor := func(t time.Time) float64 {
		if t.Before(start) {
			return padL
		}
		if !t.Before(end) && t.Equal(end) {
			return padL + chartW
		}
		ratio := float64(t.Sub(start)) / float64(duration)
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		return padL + ratio*chartW
	}

	yFor := func(v, minV, maxV float64) float64 {
		if maxV-minV < 0.001 {
			return padT + chartH/2
		}
		ratio := (v - minV) / (maxV - minV)
		return padT + chartH - ratio*chartH
	}

	yTemp := func(v float64) float64 { return yFor(v, tempMin, tempMax) }
	yWind := func(v float64) float64 { return yFor(v, 0, windScaleMax) }
	yHumid := func(v float64) float64 { return yFor(v, 0, humidityMaxPct) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" preserveAspectRatio="xMidYMid meet" role="img" style="display:block;width:100%%;height:auto;">`, width, height)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="#fff"/>`)
	fmt.Fprintf(&b, `<clipPath id="plot"><rect x="%.0f" y="%.0f" width="%.0f" height="%.0f"/></clipPath>`, padL, padT, chartW, chartH)
	// Monochrome e-paper: stroke/fill only pure black (no greys).
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="none" stroke="#000" stroke-width="1.5"/>`, padL, padT, chartW, chartH)

	for i := 0; i <= 4; i++ {
		v := tempMin + (tempMax-tempMin)*float64(i)/4
		y := yTemp(v)
		fmt.Fprintf(&b, `<line x1="%.0f" y1="%.1f" x2="%.0f" y2="%.1f" stroke="#000" stroke-width="1"/>`, padL, y, padL+chartW, y)
	}

	axisX := padL + chartW + 4
	if showWind {
		for i := 0; i <= 4; i++ {
			v := windScaleMax * float64(i) / 4
			y := yWind(v)
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="%.0f" fill="#000" text-anchor="start">%.0f</text>`, axisX, y+4, chartFontAxis, v)
		}
		fmt.Fprintf(&b, `<text x="%.0f" y="12" font-size="%.0f" fill="#000">km/h</text>`, axisX, chartFontAxis)
		axisX += 28
	}
	if showHumidity {
		for i := 0; i <= 4; i++ {
			v := humidityMaxPct * float64(i) / 4
			y := yHumid(v)
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="%.0f" fill="#000" text-anchor="start">%.0f%%</text>`, axisX, y+4, chartFontAxis, v)
		}
		fmt.Fprintf(&b, `<text x="%.0f" y="12" font-size="%.0f" fill="#000">%%</text>`, axisX, chartFontAxis)
	}

	drawLine := func(values []float64, yFn func(float64) float64, stroke string, strokeWidth float64) {
		if len(values) == 0 {
			return
		}
		b.WriteString(`<g clip-path="url(#plot)">`)
		fmt.Fprintf(&b, `<polyline fill="none" stroke="%s" stroke-width="%.1f" points="`, stroke, strokeWidth)
		for i, h := range hourly {
			t, _ := time.Parse(time.RFC3339, h.Time)
			x := xFor(t)
			y := yFn(values[i])
			fmt.Fprintf(&b, "%.1f,%.1f ", x, y)
		}
		b.WriteString(`"/></g>`)
	}

	drawLine(temps, yTemp, "#000", 2.5)
	if showWind {
		drawLine(winds, yWind, "#000", 2)
	}
	if showHumidity {
		drawLine(humids, yHumid, "#000", 2)
	}

	for i, h := range hourly {
		t, _ := time.Parse(time.RFC3339, h.Time)
		x := xFor(t)
		yt := yTemp(temps[i])

		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.2" fill="#000"/>`, x, yt)
		showLabel := !opts.SparseLabels || labelIdx[i]
		if showLabel {
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="%.0f" fill="#000" text-anchor="middle" font-weight="700">%.0f°</text>`, x, yt-10, chartFontPoint, temps[i])
		}

		if i%2 == 0 || i == len(hourly)-1 {
			label := t.In(loc).Format("15h")
			fmt.Fprintf(&b, `<text x="%.1f" y="%.0f" font-size="%.0f" fill="#000" text-anchor="middle" font-weight="600">%s</text>`, x, xLabelY, chartFontHour, label)
		}
	}

	x0 := xFor(start)
	if showWind {
		yw := yWind(winds[0])
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3" fill="#000"/>`, x0, yw)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="%.0f" fill="#000" text-anchor="start" font-weight="700">%.0f km/h</text>`, x0+6, yw-8, chartFontCurrent, winds[0])
	}
	if showHumidity {
		yh := yHumid(humids[0])
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3" fill="#000"/>`, x0, yh)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="%.0f" fill="#000" text-anchor="start" font-weight="700">%.0f%%</text>`, x0+6, yh-8, chartFontCurrent, humids[0])
	}

	drawSunLine := func(ts, label string) {
		if ts == "" {
			return
		}
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return
		}
		if t.Before(start) || t.After(end) {
			return
		}
		x := xFor(t)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.0f" x2="%.1f" y2="%.0f" stroke="#000" stroke-width="2" stroke-dasharray="5 3"/>`, x, padT, x, padT+chartH)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.0f" font-size="%.0f" fill="#000" font-weight="700">%s</text>`, x+3, padT+14, chartFontSun, label)
	}

	drawSunLine(sun.NextSunrise, "↑ lever")
	drawSunLine(sun.NextSunset, "↓ coucher")

	b.WriteString(`</svg>`)
	return b.String()
}

func paddedTempRange(values []float64) (float64, float64) {
	minV, maxV := minMax(values)
	padding := math.Max(2, (maxV-minV)*0.1)
	return math.Floor(minV-padding), math.Ceil(maxV+padding)
}

func minMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}
	minV, maxV := values[0], values[0]
	for _, v := range values[1:] {
		minV = math.Min(minV, v)
		maxV = math.Max(maxV, v)
	}
	if minV == maxV {
		maxV = minV + 1
	}
	return minV, maxV
}
