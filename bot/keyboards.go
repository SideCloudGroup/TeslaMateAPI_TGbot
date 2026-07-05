package bot

import (
	"fmt"

	"teslamate-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GetMainMenu 获取主菜单键盘
func GetMainMenu(showCarSwitch bool) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 车辆信息", "info"),
			tgbotapi.NewInlineKeyboardButtonData("⚡ 当前状态", "status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔋 电池健康", "battery"),
			tgbotapi.NewInlineKeyboardButtonData("🔌 最新充电", "charge"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚗 最近驾驶", "drive"),
		),
	}
	if showCarSwitch {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 切换车辆", "cars"),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// GetRefreshMenu 获取刷新菜单（带返回按钮）
func GetRefreshMenu(refreshType string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 刷新", "refresh_"+refreshType),
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "back_main"),
		),
	)
}

// GetCarSelectMenu 获取车辆选择键盘
func GetCarSelectMenu(cars []models.Car, selectedID int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, car := range cars {
		label := car.Name
		if car.CarID == selectedID {
			label = "✅ " + label
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				label,
				fmt.Sprintf("select_car_%d", car.CarID),
			),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "back_main"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
