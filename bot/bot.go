package bot

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pplul/teslamateapi-tgbot/client"
)

// Bot Telegram Bot结构
type Bot struct {
	api              *tgbotapi.BotAPI
	handler          *Handler
	whitelistChatIDs map[int64]bool
}

// NewBot 创建新的Bot实例
func NewBot(token string, whitelistChatIDs []int64, apiEndpoint string, tmClient *client.Client) (*Bot, error) {
	var botAPI *tgbotapi.BotAPI
	var err error

	// 初始化Telegram Bot API
	if apiEndpoint != "" {
		// 使用自定义API端点
		botAPI, err = tgbotapi.NewBotAPIWithAPIEndpoint(token, apiEndpoint+"/bot%s/%s")
		if err != nil {
			return nil, fmt.Errorf("初始化Telegram Bot失败（自定义API: %s）: %w", apiEndpoint, err)
		}
		log.Printf("使用自定义Telegram API: %s", apiEndpoint)
	} else {
		// 使用默认Telegram API
		botAPI, err = tgbotapi.NewBotAPI(token)
		if err != nil {
			return nil, fmt.Errorf("初始化Telegram Bot失败: %w", err)
		}
		log.Println("使用默认Telegram API")
	}

	// 创建白名单映射
	whitelist := make(map[int64]bool)
	for _, chatID := range whitelistChatIDs {
		whitelist[chatID] = true
	}

	log.Printf("已授权使用 Bot: %s", botAPI.Self.UserName)

	return &Bot{
		api:              botAPI,
		handler:          NewHandler(tmClient),
		whitelistChatIDs: whitelist,
	}, nil
}

// Start 启动Bot
func (b *Bot) Start() error {
	log.Println("开始接收消息...")

	// 配置更新
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// 获取更新通道
	updates := b.api.GetUpdatesChan(u)

	// 处理更新
	for update := range updates {
		// 处理消息
		if update.Message != nil {
			b.handleMessage(update.Message)
		}

		// 处理回调查询
		if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

// isAuthorized 检查用户是否在白名单中
func (b *Bot) isAuthorized(chatID int64) bool {
	return b.whitelistChatIDs[chatID]
}

// handleMessage 处理文本消息
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// 检查白名单
	if !b.isAuthorized(message.Chat.ID) {
		log.Printf("未授权访问尝试: ChatID=%d, User=%s", message.Chat.ID, message.From.UserName)
		// 不做任何响应，直接返回
		return
	}

	// 处理命令
	if message.IsCommand() {
		b.handleCommand(message)
		return
	}
}

// handleCommand 处理命令
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	chatID := message.Chat.ID

	log.Printf("收到命令: %s, ChatID=%d", command, chatID)

	switch command {
	case "start":
		text := b.handler.HandleStart()
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = GetMainMenu()
		b.api.Send(msg)

	case "help":
		text := b.handler.HandleHelp()
		msg := tgbotapi.NewMessage(chatID, text)
		b.api.Send(msg)

	case "info":
		b.sendInfo(chatID)

	case "status":
		b.sendStatus(chatID)

	case "battery":
		b.sendBattery(chatID)

	case "charge":
		b.sendCharge(chatID)

	default:
		msg := tgbotapi.NewMessage(chatID, "❓ 未知命令，请使用 /help 查看可用命令")
		b.api.Send(msg)
	}
}

// handleCallbackQuery 处理回调查询
func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// 检查白名单
	if !b.isAuthorized(query.Message.Chat.ID) {
		// 不做任何响应，直接返回
		return
	}

	data := query.Data
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	log.Printf("收到回调: %s, ChatID=%d", data, chatID)

	// 先回应回调查询
	callback := tgbotapi.NewCallback(query.ID, "")
	b.api.Request(callback)

	// 处理不同的回调
	switch {
	case data == "info":
		b.sendInfo(chatID)

	case data == "status":
		b.sendStatus(chatID)

	case data == "battery":
		b.sendBattery(chatID)

	case data == "charge":
		b.sendCharge(chatID)

	case data == "back_main":
		text := b.handler.HandleStart()
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetMainMenu()
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case strings.HasPrefix(data, "refresh_"):
		refreshType := strings.TrimPrefix(data, "refresh_")
		b.handleRefresh(chatID, messageID, refreshType)

	default:
		b.api.Request(tgbotapi.NewCallback(query.ID, "❓ 未知操作"))
	}
}

// handleRefresh 处理刷新操作
func (b *Bot) handleRefresh(chatID int64, messageID int, refreshType string) {
	switch refreshType {
	case "info":
		text, err := b.handler.HandleInfo()
		if err != nil {
			text = fmt.Sprintf("❌ 获取车辆信息失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("info")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case "status":
		text, err := b.handler.HandleStatus()
		if err != nil {
			text = fmt.Sprintf("❌ 获取车辆状态失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("status")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case "battery":
		text, err := b.handler.HandleBattery()
		if err != nil {
			text = fmt.Sprintf("❌ 获取电池健康度失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("battery")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case "charge":
		text, err := b.handler.HandleCharge()
		if err != nil {
			text = fmt.Sprintf("❌ 获取充电记录失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("charge")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)
	}
}

// sendInfo 发送车辆信息
func (b *Bot) sendInfo(chatID int64) {
	text, err := b.handler.HandleInfo()
	if err != nil {
		text = fmt.Sprintf("❌ 获取车辆信息失败: %v", err)
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = GetRefreshMenu("info")
	b.api.Send(msg)
}

// sendStatus 发送车辆状态
func (b *Bot) sendStatus(chatID int64) {
	text, err := b.handler.HandleStatus()
	if err != nil {
		text = fmt.Sprintf("❌ 获取车辆状态失败: %v", err)
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = GetRefreshMenu("status")
	b.api.Send(msg)
}

// sendBattery 发送电池健康度
func (b *Bot) sendBattery(chatID int64) {
	text, err := b.handler.HandleBattery()
	if err != nil {
		text = fmt.Sprintf("❌ 获取电池健康度失败: %v", err)
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = GetRefreshMenu("battery")
	b.api.Send(msg)
}

// sendCharge 发送最新充电记录
func (b *Bot) sendCharge(chatID int64) {
	text, err := b.handler.HandleCharge()
	if err != nil {
		text = fmt.Sprintf("❌ 获取充电记录失败: %v", err)
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = GetRefreshMenu("charge")
	b.api.Send(msg)
}
