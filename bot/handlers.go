package bot

import (
	"fmt"
	"strings"

	"teslamate-bot/client"
)

// Handler 处理器结构
type Handler struct {
	client *client.Client
}

// NewHandler 创建新的处理器
func NewHandler(tmClient *client.Client) *Handler {
	return &Handler{
		client: tmClient,
	}
}

// HandleStart 处理/start命令
func (h *Handler) HandleStart() string {
	return "🚗 欢迎使用Tesla车辆监控Bot\n\n" +
		"请选择您要查看的信息："
}

// HandleHelp 处理/help命令
func (h *Handler) HandleHelp() string {
	return "📖 可用命令：\n\n" +
		"/start - 显示主菜单\n" +
		"/info - 查看车辆详细信息\n" +
		"/status - 查看车辆当前状态\n" +
		"/battery - 查看电池健康度\n" +
		"/charge - 查看最新充电记录\n" +
		"/drive - 查看最近一次驾驶信息\n" +
		"/help - 显示帮助信息"
}

// HandleInfo 处理车辆信息请求
func (h *Handler) HandleInfo() (string, error) {
	car, err := h.client.GetCarDetails()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"📋 车辆详细信息\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"🚗 名称: %s\n"+
			"📱 型号: Model %s %s\n"+
			"🔢 VIN: %s\n"+
			"🎨 颜色: %s\n"+
			"🛞 轮毂: %s\n"+
			"📊 效率: %.2f kWh/km\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"📈 统计数据:\n"+
			"  🔌 总充电次数: %d\n"+
			"  🚙 总行驶次数: %d\n"+
			"  📲 系统更新次数: %d",
		car.Name,
		car.CarDetails.Model,
		car.CarDetails.TrimBadging,
		car.CarDetails.VIN,
		car.CarExterior.ExteriorColor,
		car.CarExterior.WheelType,
		car.CarDetails.Efficiency,
		car.TeslaMateStats.TotalCharges,
		car.TeslaMateStats.TotalDrives,
		car.TeslaMateStats.TotalUpdates,
	), nil
}

// HandleStatus 处理车辆状态请求
func (h *Handler) HandleStatus() (string, error) {
	statusResp, err := h.client.GetCarStatus()
	if err != nil {
		return "", err
	}

	status := statusResp.Data.Status
	units := statusResp.Data.Units

	// 格式化车辆状态
	stateEmoji := "🔴"
	if status.State == "online" {
		stateEmoji = "🟢"
	} else if status.State == "asleep" {
		stateEmoji = "🟡"
	}

	// 充电状态
	chargingStatus := "未充电"
	if status.ChargingDetails.PluggedIn {
		if status.ChargingDetails.ChargingState != "" {
			chargingStatus = fmt.Sprintf("充电中 (%.1f kW)", float64(status.ChargingDetails.ChargerPower))
		} else {
			chargingStatus = "已插入，未充电"
		}
	}

	// 车门/车窗状态
	doorStatus := "🔒 已锁定"
	if !status.CarStatusInfo.Locked {
		doorStatus = "🔓 未锁定"
	}

	windowStatus := "已关闭"
	if status.CarStatusInfo.WindowsOpen {
		windowStatus = "⚠️ 开启"
	}

	sentryStatus := "关闭"
	if status.CarStatusInfo.SentryMode {
		sentryStatus = "✅ 开启"
	}

	return fmt.Sprintf(
		"🚗 %s (Model %s)\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"%s 车辆状态: %s\n"+
			"🔋 电量: %d%% (%.2f %s)\n"+
			"🔌 充电: %s\n"+
			"🌡️ 车内温度: %.1f°%s\n"+
			"🌡️ 车外温度: %.1f°%s\n"+
			"%s\n"+
			"🪟 车窗: %s\n"+
			"🚨 哨兵模式: %s\n"+
			"📏 里程: %.2f %s\n"+
			"⏰ 状态更新: %s",
		status.DisplayName,
		status.CarDetails.Model,
		stateEmoji,
		status.State,
		status.BatteryDetails.BatteryLevel,
		status.BatteryDetails.RatedBatteryRange,
		units.UnitOfLength,
		chargingStatus,
		status.ClimateDetails.InsideTemp,
		units.UnitOfTemperature,
		status.ClimateDetails.OutsideTemp,
		units.UnitOfTemperature,
		doorStatus,
		windowStatus,
		sentryStatus,
		status.Odometer,
		units.UnitOfLength,
		formatDateTime(status.StateSince),
	), nil
}

// HandleBattery 处理电池健康度请求
func (h *Handler) HandleBattery() (string, error) {
	batteryResp, err := h.client.GetBatteryHealth()
	if err != nil {
		return "", err
	}

	battery := batteryResp.Data.BatteryHealth
	units := batteryResp.Data.Units

	// 健康度emoji
	healthEmoji := "💚"
	if battery.BatteryHealthPercentage < 95 {
		healthEmoji = "💛"
	}
	if battery.BatteryHealthPercentage < 90 {
		healthEmoji = "🧡"
	}
	if battery.BatteryHealthPercentage < 85 {
		healthEmoji = "❤️"
	}

	return fmt.Sprintf(
		"🔋 电池健康度\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"%s 健康度: %.2f%%\n"+
			"📊 当前容量: %.2f kWh\n"+
			"📊 最大容量: %.2f kWh\n"+
			"📏 当前续航: %.2f %s\n"+
			"📏 最大续航: %.2f %s\n"+
			"⚡ 额定效率: %.0f Wh/km",
		healthEmoji,
		battery.BatteryHealthPercentage,
		battery.CurrentCapacity,
		battery.MaxCapacity,
		battery.CurrentRange,
		units.UnitOfLength,
		battery.MaxRange,
		units.UnitOfLength,
		battery.RatedEfficiency,
	), nil
}

// HandleCharge 处理最新充电记录请求
func (h *Handler) HandleCharge() (string, error) {
	charge, err := h.client.GetLatestCharge()
	if err != nil {
		return "", err
	}

	// 解析日期时间
	startDate, startTime := splitDateTime(charge.StartDate)
	endTime := extractTime(charge.EndDate)

	return fmt.Sprintf(
		"🔌 最新充电记录\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"📅 日期: %s\n"+
			"🕐 开始: %s\n"+
			"🕐 结束: %s\n"+
			"⏱️ 时长: %s\n"+
			"⚡ 充入电量: %.2f kWh\n"+
			"🔋 电量变化: %d%% → %d%%\n"+
			"📏 续航增加: %.0f km → %.0f km\n"+
			"💰 费用: ¥%.2f\n"+
			"🌡️ 平均温度: %.0f°C",
		startDate,
		startTime,
		endTime,
		charge.DurationStr,
		charge.ChargeEnergyAdded,
		charge.BatteryDetails.StartBatteryLevel,
		charge.BatteryDetails.EndBatteryLevel,
		charge.RangeRated.StartRange,
		charge.RangeRated.EndRange,
		charge.Cost,
		charge.OutsideTempAvg,
	), nil
}

// HandleDrive 处理最近一次驾驶信息请求
func (h *Handler) HandleDrive() (string, error) {
	drive, units, err := h.client.GetLatestDrive()
	if err != nil {
		return "", err
	}

	startDate, startTime := splitDateTime(drive.StartDate)
	endTime := extractTime(drive.EndDate)

	return fmt.Sprintf(
		"🚗 最近一次驾驶\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"📅 日期: %s\n"+
			"🕐 开始: %s\n"+
			"🕐 结束: %s\n"+
			"⏱️ 时长: %s\n"+
			"📏 里程: %.2f %s\n"+
			"📊 表显: %.2f → %.2f %s\n"+
			"🔋 电量: %d%% → %d%%\n"+
			"📏 续航: %.0f → %.0f %s\n"+
			"⚡ 能耗: %.2f kWh (%.0f Wh/%s)\n"+
			"🌡️ 车外/车内: %.1f°%s / %.1f°%s\n"+
			"🚀 最高速度: %.0f %s/h | 平均: %.0f %s/h",
		startDate,
		startTime,
		endTime,
		drive.DurationStr,
		drive.OdometerDetails.OdometerDistance,
		units.UnitOfLength,
		drive.OdometerDetails.OdometerStart,
		drive.OdometerDetails.OdometerEnd,
		units.UnitOfLength,
		drive.BatteryDetails.StartBatteryLevel,
		drive.BatteryDetails.EndBatteryLevel,
		drive.RangeRated.StartRange,
		drive.RangeRated.EndRange,
		units.UnitOfLength,
		drive.EnergyConsumedNet,
		drive.ConsumptionNet,
		units.UnitOfLength,
		drive.OutsideTempAvg,
		units.UnitOfTemperature,
		drive.InsideTempAvg,
		units.UnitOfTemperature,
		drive.SpeedMax,
		units.UnitOfLength,
		drive.SpeedAvg,
		units.UnitOfLength,
	), nil
}

// formatDateTime 格式化日期时间
func formatDateTime(datetime string) string {
	// 简单处理，只取日期和时间部分
	if len(datetime) >= 19 {
		return datetime[:19]
	}
	return datetime
}

// splitDateTime 分割日期和时间
func splitDateTime(datetime string) (string, string) {
	if len(datetime) < 19 {
		return datetime, ""
	}
	parts := strings.Split(datetime[:19], "T")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return datetime, ""
}

// extractTime 提取时间部分
func extractTime(datetime string) string {
	_, time := splitDateTime(datetime)
	return time
}
