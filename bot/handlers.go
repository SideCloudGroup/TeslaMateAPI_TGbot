package bot

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"

	"teslamate-bot/client"
	"teslamate-bot/models"
)

var localLoc *time.Location

func init() {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil || loc == nil {
		localLoc = time.UTC
	} else {
		localLoc = loc
	}
}

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
func (h *Handler) HandleStart(carName string) string {
	text := "🚗 欢迎使用 Tesla 车辆监控 Bot\n"
	if carName != "" {
		text += fmt.Sprintf("当前车辆：%s\n", carName)
	}
	text += "\n请选择您要查看的信息："
	return text
}

// HandleHelp 处理/help命令
func (h *Handler) HandleHelp() string {
	return "📖 可用命令：\n\n" +
		"/start - 显示主菜单\n" +
		"/cars - 查看并切换车辆\n" +
		"/info - 查看车辆详细信息\n" +
		"/status - 查看车辆当前状态\n" +
		"/battery - 查看电池健康度\n" +
		"/charge - 查看最新充电记录\n" +
		"/drive - 查看最近一次驾驶信息\n" +
		"/help - 显示帮助信息"
}

// HandleCars 获取车辆列表及展示文本
func (h *Handler) HandleCars(selectedID int) (string, []models.Car, error) {
	cars, err := h.client.GetCars()
	if err != nil {
		return "", nil, err
	}

	var b strings.Builder
	b.WriteString("🚗 车辆列表\n")
	b.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	for _, car := range cars {
		marker := "  "
		if car.CarID == selectedID {
			marker = "✅"
		}
		b.WriteString(fmt.Sprintf("%s %s (Model %s)\n", marker, car.Name, car.CarDetails.Model))
		b.WriteString(fmt.Sprintf("   ID: %d | 颜色: %s\n", car.CarID, car.CarExterior.ExteriorColor))
	}
	b.WriteString("\n请选择要查看的车辆：")
	return b.String(), cars, nil
}

// CarDisplayName 根据车辆列表解析显示名称
func (h *Handler) CarDisplayName(cars []models.Car, carID int) string {
	for _, car := range cars {
		if car.CarID == carID {
			return car.Name
		}
	}
	return fmt.Sprintf("车辆 #%d", carID)
}

// maskVIN 对 VIN 打码：保留前 4 位与后 4 位，中间用 * 替换
func maskVIN(vin string) string {
	vin = strings.TrimSpace(vin)
	if vin == "" {
		return ""
	}
	runes := []rune(vin)
	n := len(runes)
	if n <= 8 {
		return strings.Repeat("*", n)
	}
	return string(runes[:4]) + strings.Repeat("*", n-8) + string(runes[n-4:])
}

// HandleInfo 处理车辆信息请求
func (h *Handler) HandleInfo(carID int) (string, error) {
	car, err := h.client.GetCarDetails(carID)
	if err != nil {
		return "", err
	}

	version := "未知"
	if statusResp, statusErr := h.client.GetCarStatus(carID); statusErr == nil {
		if v := statusResp.Data.Status.CarVersions.Version; v != "" {
			version = v
		}
	}

	return fmt.Sprintf(
		"📋 车辆详细信息\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"🚗 名称: %s\n"+
			"📱 型号: Model %s %s\n"+
			"🔢 VIN: %s\n"+
			"📲 版本: %s\n"+
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
		maskVIN(car.CarDetails.VIN),
		version,
		car.CarExterior.ExteriorColor,
		car.CarExterior.WheelType,
		car.CarDetails.Efficiency,
		car.TeslaMateStats.TotalCharges,
		car.TeslaMateStats.TotalDrives,
		car.TeslaMateStats.TotalUpdates,
	), nil
}

// HandleStatus 处理车辆状态请求
func (h *Handler) HandleStatus(carID int) (string, error) {
	statusResp, err := h.client.GetCarStatus(carID)
	if err != nil {
		carName := ""
		if car, detailErr := h.client.GetCarDetails(carID); detailErr == nil {
			carName = car.Name
		}
		return h.buildStatusUnavailableText(carID, carName), nil
	}

	status := statusResp.Data.Status
	car := statusResp.Data.Car
	units := statusResp.Data.Units

	if isRealtimeStatusEmpty(status) {
		carName := status.DisplayName
		if carName == "" {
			carName = car.CarName
		}
		return h.buildStatusUnavailableText(carID, carName), nil
	}

	displayName := status.DisplayName
	if displayName == "" {
		displayName = car.CarName
	}

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
			chargingStatus = fmt.Sprintf("充电中 (%.1f kW)", status.ChargingDetails.ChargerPower)
		} else {
			chargingStatus = "已插入，未充电"
		}
	}

	windowStatus := "已关闭"
	if status.CarStatusInfo.WindowsOpen {
		windowStatus = "⚠️ 开启"
	}

	sentryStatus := "关闭"
	if status.CarStatusInfo.SentryMode {
		sentryStatus = "✅ 开启"
	}

	speedLine := ""
	if isDriving(status) {
		speed := int(status.DrivingDetails.Speed)
		shift := status.DrivingDetails.ShiftState
		if shift != "" {
			speedLine = fmt.Sprintf("🚀 当前速度: %d %s/h (%s)\n", speed, html.EscapeString(units.UnitOfLength), html.EscapeString(shift))
		} else {
			speedLine = fmt.Sprintf("🚀 当前速度: %d %s/h\n", speed, html.EscapeString(units.UnitOfLength))
		}
	}

	todayDriveLine := ""
	if todayDistance, todayCount, todayUnits, err := h.client.GetTodayDriveDistance(carID); err == nil {
		lengthUnit := units.UnitOfLength
		if todayUnits != nil && todayUnits.UnitOfLength != "" {
			lengthUnit = todayUnits.UnitOfLength
		}
		todayDriveLine = fmt.Sprintf("📅 今日行驶: %.2f %s (%d 次)\n", todayDistance, html.EscapeString(lengthUnit), todayCount)
	}

	diagram := html.EscapeString(buildCarDiagram(status, units))

	return fmt.Sprintf(
		"🚗 %s (Model %s)\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"%s 车辆状态: %s\n"+
			"%s"+
			"<pre>%s</pre>\n"+
			"🔋 电量: %d%% (%.2f %s)\n"+
			"🔌 充电: %s\n"+
			"🌡️ 车内温度: %.1f°%s\n"+
			"🌡️ 车外温度: %.1f°%s\n"+
			"🪟 车窗: %s\n"+
			"🚨 哨兵模式: %s\n"+
			"%s"+
			"📏 里程: %.2f %s\n"+
			"⏰ 状态更新: %s",
		html.EscapeString(displayName),
		html.EscapeString(status.CarDetails.Model),
		stateEmoji,
		html.EscapeString(status.State),
		speedLine,
		diagram,
		status.BatteryDetails.BatteryLevel,
		status.BatteryDetails.RatedBatteryRange,
		html.EscapeString(units.UnitOfLength),
		html.EscapeString(chargingStatus),
		status.ClimateDetails.InsideTemp,
		html.EscapeString(units.UnitOfTemperature),
		status.ClimateDetails.OutsideTemp,
		html.EscapeString(units.UnitOfTemperature),
		windowStatus,
		sentryStatus,
		todayDriveLine,
		status.Odometer,
		html.EscapeString(units.UnitOfLength),
		html.EscapeString(formatDateTimeLocal(status.StateSince)),
	), nil
}

func isRealtimeStatusEmpty(status models.CarStatus) bool {
	if status.State != "" || status.DisplayName != "" {
		return false
	}
	if int(status.BatteryDetails.BatteryLevel) > 0 {
		return false
	}
	if status.ClimateDetails.InsideTemp != 0 || status.ClimateDetails.OutsideTemp != 0 {
		return false
	}
	if status.Odometer > 0 {
		return false
	}
	return true
}

func (h *Handler) buildStatusUnavailableText(carID int, carName string) string {
	if carName == "" {
		carName = fmt.Sprintf("车辆 #%d", carID)
	}
	text := fmt.Sprintf("🚗 %s\n━━━━━━━━━━━━━━━━━━━━\n暂无最后状态数据", html.EscapeString(carName))
	if todayDistance, todayCount, todayUnits, err := h.client.GetTodayDriveDistance(carID); err == nil && todayCount > 0 {
		lengthUnit := "km"
		if todayUnits != nil && todayUnits.UnitOfLength != "" {
			lengthUnit = todayUnits.UnitOfLength
		}
		text += fmt.Sprintf("\n📅 今日行驶: %.2f %s (%d 次)", todayDistance, html.EscapeString(lengthUnit), todayCount)
	}
	return text
}

func isDriving(status models.CarStatus) bool {
	shift := status.DrivingDetails.ShiftState
	if shift == "D" || shift == "R" || shift == "N" {
		return true
	}
	return status.State == "online" && status.DrivingDetails.Speed > 0
}

// HandleBattery 处理电池健康度请求
func (h *Handler) HandleBattery(carID int) (string, error) {
	batteryResp, err := h.client.GetBatteryHealth(carID)
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
func (h *Handler) HandleCharge(carID int) (string, error) {
	charge, err := h.client.GetLatestCharge(carID)
	if err != nil {
		return "", err
	}

	// 解析日期时间（转为本地时区显示）
	startDate, startTime := splitDateTimeLocal(charge.StartDate)
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
func (h *Handler) HandleDrive(carID int) (string, error) {
	drive, units, err := h.client.GetLatestDrive(carID)
	if err != nil {
		return "", err
	}

	startDate, startTime := splitDateTimeLocal(drive.StartDate)
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

// formatDateTimeLocal 将 API 时间字符串转为本地时区并格式化显示
func formatDateTimeLocal(datetime string) string {
	if datetime == "" {
		return "未知"
	}
	t, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		if len(datetime) >= 19 {
			return datetime[:19]
		}
		return datetime
	}
	if t.Year() < 2000 {
		return "未知"
	}
	return t.In(localLoc).Format("2006-01-02 15:04:05")
}

// splitDateTimeLocal 将 API 时间字符串转为本地时区，返回日期与时间部分
func splitDateTimeLocal(datetime string) (dateStr, timeStr string) {
	t, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return splitDateTime(datetime)
	}
	local := t.In(localLoc)
	return local.Format("2006-01-02"), local.Format("15:04:05")
}

// splitDateTime 分割日期和时间（用于解析失败回退）
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

// extractTime 提取时间部分（本地时区）
func extractTime(datetime string) string {
	_, timeStr := splitDateTimeLocal(datetime)
	return timeStr
}
