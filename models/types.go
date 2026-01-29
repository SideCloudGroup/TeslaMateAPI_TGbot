package models

// CarResponse 车辆详情响应
type CarResponse struct {
	Data struct {
		Cars []Car `json:"cars"`
	} `json:"data"`
}

// Car 车辆信息
type Car struct {
	CarID            int              `json:"car_id"`
	Name             string           `json:"name"`
	CarDetails       CarDetails       `json:"car_details"`
	CarExterior      CarExterior      `json:"car_exterior"`
	CarSettings      CarSettings      `json:"car_settings"`
	TeslaMateDetails TeslaMateDetails `json:"teslamate_details"`
	TeslaMateStats   TeslaMateStats   `json:"teslamate_stats"`
}

// CarDetails 车辆详细信息
type CarDetails struct {
	EID         int64   `json:"eid"`
	VID         int64   `json:"vid"`
	VIN         string  `json:"vin"`
	Model       string  `json:"model"`
	TrimBadging string  `json:"trim_badging"`
	Efficiency  float64 `json:"efficiency"`
}

// CarExterior 车辆外观
type CarExterior struct {
	ExteriorColor string `json:"exterior_color"`
	SpoilerType   string `json:"spoiler_type"`
	WheelType     string `json:"wheel_type"`
}

// CarSettings 车辆设置
type CarSettings struct {
	SuspendMin          int  `json:"suspend_min"`
	SuspendAfterIdleMin int  `json:"suspend_after_idle_min"`
	ReqNotUnlocked      bool `json:"req_not_unlocked"`
	FreeSupercharging   bool `json:"free_supercharging"`
	UseStreamingAPI     bool `json:"use_streaming_api"`
}

// TeslaMateDetails TeslaMate记录详情
type TeslaMateDetails struct {
	InsertedAt string `json:"inserted_at"`
	UpdatedAt  string `json:"updated_at"`
}

// TeslaMateStats TeslaMate统计
type TeslaMateStats struct {
	TotalCharges int `json:"total_charges"`
	TotalDrives  int `json:"total_drives"`
	TotalUpdates int `json:"total_updates"`
}

// StatusResponse 车辆状态响应
type StatusResponse struct {
	Data struct {
		Car    StatusCar `json:"car"`
		Status CarStatus `json:"status"`
		Units  Units     `json:"units"`
	} `json:"data"`
}

// StatusCar 状态中的车辆信息
type StatusCar struct {
	CarID   int    `json:"car_id"`
	CarName string `json:"car_name"`
}

// CarStatus 车辆状态详情
type CarStatus struct {
	DisplayName     string           `json:"display_name"`
	State           string           `json:"state"`
	StateSince      string           `json:"state_since"`
	Odometer        float64          `json:"odometer"`
	CarStatusInfo   CarStatusInfo    `json:"car_status"`
	CarDetails      StatusCarDetails `json:"car_details"`
	CarExterior     CarExterior      `json:"car_exterior"`
	CarGeodata      CarGeodata       `json:"car_geodata"`
	CarVersions     CarVersions      `json:"car_versions"`
	DrivingDetails  DrivingDetails   `json:"driving_details"`
	ClimateDetails  ClimateDetails   `json:"climate_details"`
	BatteryDetails  BatteryDetails   `json:"battery_details"`
	ChargingDetails ChargingDetails  `json:"charging_details"`
	TPMSDetails     TPMSDetails      `json:"tpms_details"`
}

// CarStatusInfo 车辆状态信息
type CarStatusInfo struct {
	Healthy                bool `json:"healthy"`
	Locked                 bool `json:"locked"`
	SentryMode             bool `json:"sentry_mode"`
	WindowsOpen            bool `json:"windows_open"`
	DoorsOpen              bool `json:"doors_open"`
	DriverFrontDoorOpen    bool `json:"driver_front_door_open"`
	DriverRearDoorOpen     bool `json:"driver_rear_door_open"`
	PassengerFrontDoorOpen bool `json:"passenger_front_door_open"`
	PassengerRearDoorOpen  bool `json:"passenger_rear_door_open"`
	TrunkOpen              bool `json:"trunk_open"`
	FrunkOpen              bool `json:"frunk_open"`
	IsUserPresent          bool `json:"is_user_present"`
	CenterDisplayState     int  `json:"center_display_state"`
}

// StatusCarDetails 状态中的车辆详情
type StatusCarDetails struct {
	Model       string `json:"model"`
	TrimBadging string `json:"trim_badging"`
}

// CarGeodata 车辆地理位置
type CarGeodata struct {
	Geofence  string   `json:"geofence"`
	Location  Location `json:"location"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
}

// Location 位置信息
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// CarVersions 车辆版本信息
type CarVersions struct {
	Version         string `json:"version"`
	UpdateAvailable bool   `json:"update_available"`
	UpdateVersion   string `json:"update_version"`
}

// DrivingDetails 驾驶详情
type DrivingDetails struct {
	ActiveRoute            ActiveRoute `json:"active_route"`
	ActiveRouteDestination string      `json:"active_route_destination"`
	ActiveRouteLatitude    float64     `json:"active_route_latitude"`
	ActiveRouteLongitude   float64     `json:"active_route_longitude"`
	ShiftState             string      `json:"shift_state"`
	Power                  int         `json:"power"`
	Speed                  int         `json:"speed"`
	Heading                int         `json:"heading"`
	Elevation              int         `json:"elevation"`
}

// ActiveRoute 活动路线
type ActiveRoute struct {
	Destination         string   `json:"destination"`
	EnergyAtArrival     int      `json:"energy_at_arrival"`
	DistanceToArrival   int      `json:"distance_to_arrival"`
	MinutesToArrival    int      `json:"minutes_to_arrival"`
	TrafficMinutesDelay int      `json:"traffic_minutes_delay"`
	Location            Location `json:"location"`
}

// ClimateDetails 气候详情
type ClimateDetails struct {
	IsClimateOn       bool    `json:"is_climate_on"`
	InsideTemp        float64 `json:"inside_temp"`
	OutsideTemp       float64 `json:"outside_temp"`
	IsPreconditioning bool    `json:"is_preconditioning"`
	ClimateKeeperMode string  `json:"climate_keeper_mode"`
}

// BatteryDetails 电池详情
type BatteryDetails struct {
	EstBatteryRange    float64 `json:"est_battery_range"`
	RatedBatteryRange  float64 `json:"rated_battery_range"`
	IdealBatteryRange  float64 `json:"ideal_battery_range"`
	BatteryLevel       int     `json:"battery_level"`
	UsableBatteryLevel int     `json:"usable_battery_level"`
}

// ChargingDetails 充电详情
type ChargingDetails struct {
	PluggedIn                  bool    `json:"plugged_in"`
	ChargingState              string  `json:"charging_state"`
	ChargeEnergyAdded          float64 `json:"charge_energy_added"`
	ChargeLimitSOC             int     `json:"charge_limit_soc"`
	ChargePortDoorOpen         bool    `json:"charge_port_door_open"`
	ChargerActualCurrent       int     `json:"charger_actual_current"`
	ChargerPhases              int     `json:"charger_phases"`
	ChargerPower               int     `json:"charger_power"`
	ChargerVoltage             int     `json:"charger_voltage"`
	ChargeCurrentRequest       int     `json:"charge_current_request"`
	ChargeCurrentRequestMax    int     `json:"charge_current_request_max"`
	ScheduledChargingStartTime string  `json:"scheduled_charging_start_time"`
	TimeToFullCharge           float64 `json:"time_to_full_charge"`
}

// TPMSDetails 胎压监测详情
type TPMSDetails struct {
	TPMSPressureFL    float64 `json:"tpms_pressure_fl"`
	TPMSPressureFR    float64 `json:"tpms_pressure_fr"`
	TPMSPressureRL    float64 `json:"tpms_pressure_rl"`
	TPMSPressureRR    float64 `json:"tpms_pressure_rr"`
	TPMSSoftWarningFL bool    `json:"tpms_soft_warning_fl"`
	TPMSSoftWarningFR bool    `json:"tpms_soft_warning_fr"`
	TPMSSoftWarningRL bool    `json:"tpms_soft_warning_rl"`
	TPMSSoftWarningRR bool    `json:"tpms_soft_warning_rr"`
}

// BatteryHealthResponse 电池健康度响应
type BatteryHealthResponse struct {
	Data struct {
		Car           StatusCar     `json:"car"`
		BatteryHealth BatteryHealth `json:"battery_health"`
		Units         Units         `json:"units"`
	} `json:"data"`
}

// BatteryHealth 电池健康度
type BatteryHealth struct {
	MaxRange                float64 `json:"max_range"`
	CurrentRange            float64 `json:"current_range"`
	MaxCapacity             float64 `json:"max_capacity"`
	CurrentCapacity         float64 `json:"current_capacity"`
	RatedEfficiency         float64 `json:"rated_efficiency"`
	BatteryHealthPercentage float64 `json:"battery_health_percentage"`
}

// ChargesResponse 充电记录响应
type ChargesResponse struct {
	Data struct {
		Car     StatusCar `json:"car"`
		Charges []Charge  `json:"charges"`
		Units   Units     `json:"units"`
	} `json:"data"`
}

// Charge 充电记录
type Charge struct {
	ChargeID          int                  `json:"charge_id"`
	StartDate         string               `json:"start_date"`
	EndDate           string               `json:"end_date"`
	Address           string               `json:"address"`
	ChargeEnergyAdded float64              `json:"charge_energy_added"`
	ChargeEnergyUsed  float64              `json:"charge_energy_used"`
	Cost              float64              `json:"cost"`
	DurationMin       int                  `json:"duration_min"`
	DurationStr       string               `json:"duration_str"`
	BatteryDetails    ChargeBatteryDetails `json:"battery_details"`
	RangeIdeal        ChargeRange          `json:"range_ideal"`
	RangeRated        ChargeRange          `json:"range_rated"`
	OutsideTempAvg    float64              `json:"outside_temp_avg"`
	Odometer          float64              `json:"odometer"`
	Latitude          float64              `json:"latitude"`
	Longitude         float64              `json:"longitude"`
}

// ChargeBatteryDetails 充电电池详情
type ChargeBatteryDetails struct {
	StartBatteryLevel int `json:"start_battery_level"`
	EndBatteryLevel   int `json:"end_battery_level"`
}

// ChargeRange 充电续航范围
type ChargeRange struct {
	StartRange float64 `json:"start_range"`
	EndRange   float64 `json:"end_range"`
}

// Units 单位信息
type Units struct {
	UnitOfLength      string `json:"unit_of_length"`
	UnitOfTemperature string `json:"unit_of_temperature"`
	UnitOfPressure    string `json:"unit_of_pressure,omitempty"`
}

// DrivesResponse 驾驶记录响应
type DrivesResponse struct {
	Data struct {
		Car    DrivesCar `json:"car"`
		Drives []Drive   `json:"drives"`
		Units  Units     `json:"units"`
	} `json:"data"`
}

// DrivesCar 驾驶列表中的车辆信息
type DrivesCar struct {
	CarID   int    `json:"car_id"`
	CarName string `json:"car_name"`
}

// Drive 驾驶记录
type Drive struct {
	DriveID           int                  `json:"drive_id"`
	StartDate         string               `json:"start_date"`
	EndDate           string               `json:"end_date"`
	StartAddress      string               `json:"start_address"`
	EndAddress        string               `json:"end_address"`
	OdometerDetails   DriveOdometerDetails `json:"odometer_details"`
	DurationMin       int                  `json:"duration_min"`
	DurationStr       string               `json:"duration_str"`
	SpeedMax          float64              `json:"speed_max"`
	SpeedAvg          float64              `json:"speed_avg"`
	PowerMax          float64              `json:"power_max"`
	PowerMin          float64              `json:"power_min"`
	BatteryDetails    DriveBatteryDetails  `json:"battery_details"`
	RangeIdeal        DriveRange           `json:"range_ideal"`
	RangeRated        DriveRange           `json:"range_rated"`
	OutsideTempAvg    float64              `json:"outside_temp_avg"`
	InsideTempAvg     float64              `json:"inside_temp_avg"`
	EnergyConsumedNet float64              `json:"energy_consumed_net"`
	ConsumptionNet    float64              `json:"consumption_net"`
}

// DriveOdometerDetails 驾驶里程详情
type DriveOdometerDetails struct {
	OdometerStart    float64 `json:"odometer_start"`
	OdometerEnd      float64 `json:"odometer_end"`
	OdometerDistance float64 `json:"odometer_distance"`
}

// DriveBatteryDetails 驾驶电池详情
type DriveBatteryDetails struct {
	StartUsableBatteryLevel int  `json:"start_usable_battery_level"`
	StartBatteryLevel       int  `json:"start_battery_level"`
	EndUsableBatteryLevel   int  `json:"end_usable_battery_level"`
	EndBatteryLevel         int  `json:"end_battery_level"`
	ReducedRange            bool `json:"reduced_range"`
	IsSufficientlyPrecise   bool `json:"is_sufficiently_precise"`
}

// DriveRange 驾驶续航范围
type DriveRange struct {
	StartRange float64 `json:"start_range"`
	EndRange   float64 `json:"end_range"`
	RangeDiff  float64 `json:"range_diff"`
}
