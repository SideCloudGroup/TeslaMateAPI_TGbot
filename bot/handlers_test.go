package bot

import (
	"strings"
	"testing"

	"teslamate-bot/models"
)

func TestFormatChargeElectricalDetailsHidesVoltageAndCurrentForDC(t *testing.T) {
	got := formatChargeElectricalDetails(models.ChargeElectricalStats{
		IsDC:           true,
		AverageVoltage: 400,
		AverageCurrent: 120,
		AveragePower:   48,
		MaximumPower:   75,
	})

	if strings.Contains(got, "电压") || strings.Contains(got, "电流") {
		t.Fatalf("DC details should hide voltage and current, got %q", got)
	}
	if !strings.Contains(got, "功率") {
		t.Fatalf("DC details should retain power, got %q", got)
	}
}

func TestFormatChargeElectricalDetailsShowsVoltageAndCurrentForAC(t *testing.T) {
	got := formatChargeElectricalDetails(models.ChargeElectricalStats{
		AverageVoltage: 230,
		AverageCurrent: 16,
		AveragePower:   3.7,
	})

	for _, label := range []string{"电压", "电流", "功率"} {
		if !strings.Contains(got, label) {
			t.Fatalf("AC details should contain %q, got %q", label, got)
		}
	}
}
