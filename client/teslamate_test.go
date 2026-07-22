package client

import (
	"testing"

	"teslamate-bot/models"
)

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
