package bot

import (
	"fmt"
	"strings"

	"teslamate-bot/models"
)

func hasTPMSData(tpms models.TPMSDetails) bool {
	return tpms.TPMSPressureFL > 0 || tpms.TPMSPressureFR > 0 ||
		tpms.TPMSPressureRL > 0 || tpms.TPMSPressureRR > 0
}

func formatTPMS(pressure float64, warning bool, available bool) string {
	if !available {
		return "   --"
	}
	s := fmt.Sprintf("%.2f", pressure)
	if warning {
		s = "!" + s
	}
	if len(s) < 5 {
		return fmt.Sprintf("%5s", s)
	}
	return s
}

// buildCarDiagram 生成等宽俯视车形图（含胎压、锁、充电口）
func buildCarDiagram(status models.CarStatus, units models.Units) string {
	tpms := status.TPMSDetails
	info := status.CarStatusInfo
	available := hasTPMSData(tpms)

	fl := formatTPMS(tpms.TPMSPressureFL, tpms.TPMSSoftWarningFL, available)
	fr := formatTPMS(tpms.TPMSPressureFR, tpms.TPMSSoftWarningFR, available)
	rl := formatTPMS(tpms.TPMSPressureRL, tpms.TPMSSoftWarningRL, available)
	rr := formatTPMS(tpms.TPMSPressureRR, tpms.TPMSSoftWarningRR, available)

	pressureUnit := units.UnitOfPressure
	if pressureUnit == "" {
		pressureUnit = "bar"
	}

	lock := "🔒"
	if !info.Locked {
		lock = "🔓"
	}
	// emoji 约占 2 列；电量固定 4 列，再补空格使车身内显示宽度约 10
	bat := fmt.Sprintf("%3d%%", int(status.BatteryDetails.BatteryLevel))
	center := lock + " " + bat + "   "

	plug := " "
	if status.ChargingDetails.PluggedIn {
		plug = "~"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%5s      %5s\n", fl, fr))
	b.WriteString("  *----------*\n")
	b.WriteString("  |          |\n")
	b.WriteString(fmt.Sprintf("  |%s|%s\n", center, plug))
	b.WriteString("  |          |\n")
	b.WriteString("  *----------*\n")
	b.WriteString(fmt.Sprintf("%5s      %5s\n", rl, rr))
	b.WriteString(fmt.Sprintf("      %s", pressureUnit))
	return b.String()
}
