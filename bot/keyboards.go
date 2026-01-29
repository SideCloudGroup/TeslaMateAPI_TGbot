package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GetMainMenu 获取主菜单键盘
func GetMainMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
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
	)
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
