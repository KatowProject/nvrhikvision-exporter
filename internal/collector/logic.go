package collector

import (
	"github.com/KatowProject/nvrhikvision-exporter/internal/models"
)

func isSeagate(rawAttrs map[int]int64) bool {
	// Seagate disks commonly expose attributes 18, 187, 188, and 195.
	_, has18 := rawAttrs[18]
	_, has187 := rawAttrs[187]
	_, has195 := rawAttrs[195]

	return has18 || (has187 && has195)
}

// decodeTemperature extracts the actual temperature from multiple sources.
// IDs 194 and 190 use vendor packed format (Seagate/WD):
// byte 0 (least significant) = current temperature in Celsius.
// Priority: Temperature field from XML -> ID 194 -> ID 190
func decodeTemperature(smart models.SmartTestStatus, rawAttrs map[int]int64) int {
	if smart.Temperature > 0 && smart.Temperature < 100 {
		return smart.Temperature
	}
	if raw, ok := rawAttrs[194]; ok && raw > 0 {
		t := int(raw & 0xFF)
		if t > 0 && t < 100 {
			return t
		}
	}
	if raw, ok := rawAttrs[190]; ok && raw > 0 {
		t := int(raw & 0xFF)
		if t > 0 && t < 100 {
			return t
		}
	}
	return 0
}

// CalculateHDDHealth computes HDD health percentage (Sentinel style)
func CalculateHDDHealth(smart models.SmartTestStatus) float64 {
	health := 100.0

	// Map raw values, normalized values, and worst values
	rawAttrs := make(map[int]int64)
	normAttrs := make(map[int]int)
	worstAttrs := make(map[int]int)
	for _, res := range smart.TestResultList {
		rawAttrs[res.AttributeID] = res.RawValue
		normAttrs[res.AttributeID] = res.Value
		worstAttrs[res.AttributeID] = res.Worst
	}

	seagateDisk := isSeagate(rawAttrs)

	// =========================================================
	// 1. CRITICAL GROUP - Bad Sectors (Most dangerous)
	// =========================================================

	// ID 5 (Reallocated Sectors): Damaged sectors already moved to spare area.
	// Cap at 60% to avoid over-penalizing a still-usable disk.
	if rawAttrs[5] > 0 {
		penalty := float64(rawAttrs[5]) * 2.0
		if penalty > 60 {
			penalty = 60
		}
		health -= penalty
	}

	// ID 197 (Current Pending Sectors): Sectors waiting for reallocation - VERY CRITICAL.
	// Indicates actively failing disk areas.
	if rawAttrs[197] > 0 {
		health -= 15.0                         // Heavy penalty because the disk is unstable
		health -= float64(rawAttrs[197]) * 3.0 // -3% per pending sector
	}

	// ID 198 (Offline Uncorrectable): Sectors that cannot be read at all.
	if rawAttrs[198] > 0 {
		health -= 15.0
		health -= float64(rawAttrs[198]) * 3.0
	}

	// ID 187 (Reported Uncorrectable): Errors reported by controller (Seagate).
	if rawAttrs[187] > 0 {
		health -= 10.0
		health -= float64(rawAttrs[187]) * 2.5
	}

	// =========================================================
	// 2. MECHANICAL GROUP - Mechanical issues
	// =========================================================

	// ID 10 (Spin Retry Count): Spindle motor issue - indicator of mechanical failure.
	// Cap at 45% so one attribute does not immediately force health to 0.
	if rawAttrs[10] > 0 {
		penalty := 20.0 + float64(rawAttrs[10])*5.0
		if penalty > 45 {
			penalty = 45
		}
		health -= penalty
	}

	// ID 7 (Seek Error Rate): Use worst value because it is more conservative -
	// a disk that once dropped to 60 is riskier than one stable at 85.
	if !seagateDisk {
		seekWorst := worstAttrs[7]
		seekNorm := normAttrs[7]
		if seekWorst > 0 && seekWorst < seekNorm {
			// Was worse in the past than current value, use worst
			health -= float64(100-seekWorst) * 0.3
		} else if seekNorm > 0 && seekNorm < 100 {
			health -= float64(100-seekNorm) * 0.3
		}
	}

	// =========================================================
	// 3. AGE & WEAR GROUP - Normal degradation
	// =========================================================

	// ID 9 (Power On Hours): HDD age estimation.
	powerOnHours := rawAttrs[9]
	if powerOnHours > 43800 { // > 5 years (24*365*5)
		health -= 10.0
	} else if powerOnHours > 26280 { // > 3 years
		health -= 5.0
	} else if powerOnHours > 17520 { // > 2 years
		health -= 2.0
	}

	// ID 193 (Load Cycle Count): Relevant for laptop HDDs with head parking.
	// Hikvision NVRs usually use desktop/surveillance HDDs, so this is usually low.
	if rawAttrs[193] > 600000 {
		health -= 10.0
	} else if rawAttrs[193] > 300000 {
		health -= 5.0
	}

	// ID 4 (Start/Stop Count): For 24/7 NVRs, this should be very low.
	// High values mean frequent spin down/up or frequent unexpected NVR shutdowns.
	if rawAttrs[4] > 10000 {
		health -= 5.0
	} else if rawAttrs[4] > 5000 {
		health -= 2.0
	}

	// ID 241 (Total LBAs Written): Estimated total data written to disk.
	// 1 LBA = 512 bytes. NVR workloads are write-heavy, and surveillance disks usually handle about 300 TB.
	// 300TB = 300 * 1024^4 / 512 = ~644,245,094,400 LBA
	totalWrittenLBA := rawAttrs[241]
	if totalWrittenLBA > 644245094400 { // > ~300TB written
		health -= 10.0
	} else if totalWrittenLBA > 429496729600 { // > ~200TB written
		health -= 5.0
	} else if totalWrittenLBA > 214748364800 { // > ~100TB written
		health -= 2.0
	}

	// =========================================================
	// 4. POWER & STABILITY GROUP
	// =========================================================

	// ID 12 (Power Cycle Count): Frequent on/off cycles increase mechanical risk.
	// Cross-check with power-on hours to detect abnormal patterns.
	powerCycles := rawAttrs[12]
	if powerOnHours > 0 && powerCycles > 0 {
		// Average cycles per day
		cyclesPerDay := float64(powerCycles) / (float64(powerOnHours) / 24.0)
		if cyclesPerDay > 5 {
			health -= 10.0 // Very frequent shutdowns - possible power supply issue
		} else if cyclesPerDay > 3 {
			health -= 5.0 // Fairly frequent shutdowns
		}
	}

	// ID 192 (Power-off Retract Count): How often heads are retracted due to sudden power cuts.
	// Different from power cycle - this is specific to unexpected power loss events.
	if rawAttrs[192] > 100 {
		health -= 5.0
	} else if rawAttrs[192] > 50 {
		health -= 2.0
	}

	// =========================================================
	// 5. CONNECTION GROUP - Connection/cable issues
	// =========================================================

	// ID 199 (UDMA CRC Error): Data transmission errors, usually SATA cable issues.
	if rawAttrs[199] > 0 {
		penalty := float64(rawAttrs[199]) * 0.3
		if penalty > 15 {
			penalty = 15
		}
		health -= penalty
	}

	// =========================================================
	// 6. TEMPERATURE - Temperature monitoring
	// =========================================================

	// IDs 194 and 190 use a packed format and must be decoded before use.
	// HDD Sentinel: 45-50°C = optimal, >50°C starts degrading health.
	temp := decodeTemperature(smart, rawAttrs)
	if temp > 0 { // skip if temperature data is unavailable (no sensor / 0)
		if temp > 65 {
			health -= 20.0 // Too hot - high risk
		} else if temp > 55 {
			health -= 10.0 // Hot - medium risk
		} else if temp > 50 {
			health -= 5.0 // Slightly hot - low risk
		}
		// Too cold is also bad (condensation risk)
		if temp < 20 {
			health -= 5.0
		}
	}

	// =========================================================
	// 7. READ/WRITE ERROR GROUP
	// =========================================================

	// ID 1 (Raw Read Error Rate): Use worst value if it was ever worse,
	// because Seagate raw values use a packed format.
	if !seagateDisk {
		readWorst := worstAttrs[1]
		readNorm := normAttrs[1]
		if readWorst > 0 && readWorst < readNorm {
			health -= float64(100-readWorst) * 0.2
		} else if readNorm > 0 && readNorm < 100 {
			health -= float64(100-readNorm) * 0.2
		}
	}

	// Sanity check: health must be within 0..100
	if health < 0 {
		return 0
	}
	if health > 100 {
		return 100
	}
	return health
}

// CalculateMemoryPercent converts raw memory values to percentage.
// Some Hikvision NVRs report usage in MB instead of percent.
func CalculateMemoryPercent(usage, total float64) float64 {
	if total <= 0 {
		return 0
	}
	if usage > 100 {
		// Usage in MB/Bytes, calculate percent from total
		pct := (usage / total) * 100
		if pct > 100 {
			return 100
		}
		return pct
	}
	return usage
}
