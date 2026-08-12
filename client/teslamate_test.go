package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"teslamate-bot/models"
)

func TestGetLatestChargeLoadsDetailsAndElectricalStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/cars/7/charges":
			if got := r.URL.Query().Get("show"); got != "1" {
				t.Errorf("show query = %q, want 1", got)
			}
			_, _ = w.Write([]byte(`{"data":{"charges":[{"charge_id":42}]}}`))
		case "/api/v1/cars/7/charges/42":
			_, _ = w.Write([]byte(`{"data":{"charge":{"charge_id":42,"charge_details":[{"charger_details":{"charger_voltage":230,"charger_actual_current":12,"charger_power":3}},{"charger_details":{"charger_voltage":240,"charger_actual_current":16,"charger_power":5}}]}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "", 5, nil)
	charge, err := client.GetLatestCharge(7)
	if err != nil {
		t.Fatalf("GetLatestCharge() error = %v", err)
	}
	if charge.ChargeID != 42 {
		t.Errorf("charge ID = %d, want 42", charge.ChargeID)
	}
	if charge.ElectricalStats.AverageVoltage != 235 || charge.ElectricalStats.MaximumPower != 5 {
		t.Errorf("electrical stats = %#v", charge.ElectricalStats)
	}
}

func TestLatestDriveAtLeastSkipsShorterDrives(t *testing.T) {
	drives := []models.Drive{
		{DriveID: 3, OdometerDetails: models.DriveOdometerDetails{OdometerDistance: 0.1}},
		{DriveID: 2, OdometerDetails: models.DriveOdometerDetails{OdometerDistance: 0.49}},
		{DriveID: 1, OdometerDetails: models.DriveOdometerDetails{OdometerDistance: 0.5}},
	}

	got := latestDriveAtLeast(drives, "km", minimumLatestDriveDistanceKM)
	if got == nil || got.DriveID != 1 {
		t.Fatalf("latestDriveAtLeast() = %#v, want drive ID 1", got)
	}
}

func TestLatestDriveAtLeastReturnsNilWhenAllAreShort(t *testing.T) {
	drives := []models.Drive{
		{DriveID: 2, OdometerDetails: models.DriveOdometerDetails{OdometerDistance: 0.1}},
		{DriveID: 1, OdometerDetails: models.DriveOdometerDetails{OdometerDistance: 0.49}},
	}

	if got := latestDriveAtLeast(drives, "km", minimumLatestDriveDistanceKM); got != nil {
		t.Fatalf("latestDriveAtLeast() = %#v, want nil", got)
	}
}

func TestLatestDriveAtLeastUsesKilometreThresholdForMiles(t *testing.T) {
	drives := []models.Drive{
		{DriveID: 2, OdometerDetails: models.DriveOdometerDetails{OdometerDistance: 0.3}},
		{DriveID: 1, OdometerDetails: models.DriveOdometerDetails{OdometerDistance: 0.32}},
	}

	got := latestDriveAtLeast(drives, "mi", minimumLatestDriveDistanceKM)
	if got == nil || got.DriveID != 1 {
		t.Fatalf("latestDriveAtLeast() = %#v, want drive ID 1", got)
	}
}

func TestDriveDistanceInKM(t *testing.T) {
	tests := []struct {
		name     string
		distance float64
		unit     string
		want     float64
	}{
		{name: "kilometres", distance: 0.5, unit: "km", want: 0.5},
		{name: "miles", distance: 1, unit: "mi", want: 1.609344},
		{name: "empty unit defaults to kilometres", distance: 0.5, want: 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := driveDistanceInKM(tt.distance, tt.unit); got != tt.want {
				t.Fatalf("driveDistanceInKM(%v, %q) = %v, want %v", tt.distance, tt.unit, got, tt.want)
			}
		})
	}
}

func TestSummarizeChargeElectricalStats(t *testing.T) {
	details := []models.ChargeDetail{
		{ChargerDetails: models.ChargeChargerDetails{ChargerVoltage: 0, ChargerActualCurrent: 0, ChargerPower: 0}},
		{ChargerDetails: models.ChargeChargerDetails{ChargerVoltage: 220, ChargerActualCurrent: 10, ChargerPower: 2}},
		{ChargerDetails: models.ChargeChargerDetails{ChargerVoltage: 240, ChargerActualCurrent: 20, ChargerPower: 6}},
	}

	got := summarizeChargeElectricalStats(details)
	if got.AverageVoltage != 230 || got.MaximumVoltage != 240 {
		t.Errorf("voltage stats = avg %v, max %v; want avg 230, max 240", got.AverageVoltage, got.MaximumVoltage)
	}
	if got.AverageCurrent != 15 || got.MaximumCurrent != 20 {
		t.Errorf("current stats = avg %v, max %v; want avg 15, max 20", got.AverageCurrent, got.MaximumCurrent)
	}
	if got.AveragePower != 4 || got.MaximumPower != 6 {
		t.Errorf("power stats = avg %v, max %v; want avg 4, max 6", got.AveragePower, got.MaximumPower)
	}
}

func TestSummarizeChargeElectricalStatsWithNoActiveSamples(t *testing.T) {
	got := summarizeChargeElectricalStats(nil)
	if got != (models.ChargeElectricalStats{}) {
		t.Fatalf("summarizeChargeElectricalStats(nil) = %#v, want zero stats", got)
	}
}

func TestSummarizeChargeElectricalStatsDetectsDCCharging(t *testing.T) {
	details := []models.ChargeDetail{
		{FastChargerInfo: models.ChargeFastChargerInfo{FastChargerPresent: false}},
		{FastChargerInfo: models.ChargeFastChargerInfo{FastChargerPresent: true}},
	}

	got := summarizeChargeElectricalStats(details)
	if !got.IsDC {
		t.Fatal("IsDC = false, want true when any sample reports a fast charger")
	}
}
