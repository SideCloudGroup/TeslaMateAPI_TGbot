package bot

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"teslamate-bot/client"
	"teslamate-bot/state"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot Telegram Bot结构
type Bot struct {
	api              *tgbotapi.BotAPI
	handler          *Handler
	whitelistChatIDs map[int64]bool
	carState         *state.CarStateStore
}

// NewBot 创建新的Bot实例
func NewBot(token string, whitelistChatIDs []int64, apiEndpoint, httpProxy string, tmClient *client.Client, carState *state.CarStateStore) (*Bot, error) {
	endpoint := tgbotapi.APIEndpoint
	if apiEndpoint != "" {
		endpoint = apiEndpoint + "/bot%s/%s"
		log.Printf("使用自定义Telegram API: %s", apiEndpoint)
	} else {
		log.Println("使用默认Telegram API")
	}

	httpClient, err := newTelegramHTTPClient(httpProxy)
	if err != nil {
		return nil, err
	}
	if httpProxy != "" {
		log.Printf("使用HTTP代理连接Telegram API: %s", redactProxyURL(httpProxy))
	}

	botAPI, err := tgbotapi.NewBotAPIWithClient(token, endpoint, httpClient)
	if err != nil {
		if apiEndpoint != "" {
			return nil, fmt.Errorf("初始化Telegram Bot失败（自定义API: %s）: %w", apiEndpoint, err)
		}
		return nil, fmt.Errorf("初始化Telegram Bot失败: %w", err)
	}

	whitelist := make(map[int64]bool)
	for _, chatID := range whitelistChatIDs {
		whitelist[chatID] = true
	}

	log.Printf("已授权使用 Bot: %s", botAPI.Self.UserName)

	return &Bot{
		api:              botAPI,
		handler:          NewHandler(tmClient),
		whitelistChatIDs: whitelist,
		carState:         carState,
	}, nil
}

func newTelegramHTTPClient(proxyURL string) (*http.Client, error) {
	if proxyURL == "" {
		return &http.Client{}, nil
	}

	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("解析HTTP代理URL失败: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}, nil
}

func redactProxyURL(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return proxyURL
	}
	if u.User != nil {
		u.User = url.UserPassword("***", "***")
	}
	return u.String()
}

func (b *Bot) registerCommands() error {
	cfg := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "开始使用 / 主菜单"},
		tgbotapi.BotCommand{Command: "help", Description: "查看帮助与可用命令"},
		tgbotapi.BotCommand{Command: "cars", Description: "查看并切换车辆"},
		tgbotapi.BotCommand{Command: "info", Description: "车辆信息"},
		tgbotapi.BotCommand{Command: "status", Description: "当前状态"},
		tgbotapi.BotCommand{Command: "battery", Description: "电池健康"},
		tgbotapi.BotCommand{Command: "charge", Description: "最新充电"},
		tgbotapi.BotCommand{Command: "drive", Description: "最近驾驶"},
	)
	_, err := b.api.Request(cfg)
	return err
}

// Start 启动Bot
func (b *Bot) Start() error {
	if err := b.registerCommands(); err != nil {
		log.Printf("注册 Telegram 指令失败（不影响运行）: %v", err)
	} else {
		log.Println("已注册 Telegram 指令")
	}
	log.Println("开始接收消息...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		}

		if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}

	return nil
}

func (b *Bot) isAuthorized(chatID int64) bool {
	return b.whitelistChatIDs[chatID]
}

func (b *Bot) resolveCarID(chatID int64) (int, error) {
	if id, ok := b.carState.Get(chatID); ok {
		if err := b.validateCarID(id); err != nil {
			_ = b.carState.Clear(chatID)
			return 0, err
		}
		return id, nil
	}

	defaultID := b.carState.DefaultID()
	if defaultID > 0 {
		if err := b.validateCarID(defaultID); err != nil {
			return 0, err
		}
		return defaultID, nil
	}

	cars, err := b.handler.client.GetCars()
	if err != nil {
		return 0, fmt.Errorf("无可用车辆: %w", err)
	}
	if len(cars) == 0 {
		return 0, fmt.Errorf("无可用车辆")
	}

	if len(cars) == 1 {
		_ = b.carState.Set(chatID, cars[0].CarID)
	}
	return cars[0].CarID, nil
}

func (b *Bot) validateCarID(carID int) error {
	cars, err := b.handler.client.GetCars()
	if err != nil {
		return err
	}
	for _, car := range cars {
		if car.CarID == carID {
			return nil
		}
	}
	return fmt.Errorf("车辆 #%d 不存在，请使用 /cars 重新选择", carID)
}

func (b *Bot) getCarName(carID int) string {
	cars, err := b.handler.client.GetCars()
	if err != nil {
		return fmt.Sprintf("车辆 #%d", carID)
	}
	return b.handler.CarDisplayName(cars, carID)
}

func (b *Bot) showCarSwitch() bool {
	cars, err := b.handler.client.GetCars()
	if err != nil {
		return true
	}
	return len(cars) > 1
}

func (b *Bot) carSelectErrorText() string {
	return "❌ 无法确定当前车辆，请使用 /cars 选择车辆"
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	if !b.isAuthorized(message.Chat.ID) {
		log.Printf("未授权访问尝试: ChatID=%d, User=%s", message.Chat.ID, message.From.UserName)
		return
	}

	if message.IsCommand() {
		b.handleCommand(message)
		return
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	chatID := message.Chat.ID

	log.Printf("收到命令: %s, ChatID=%d", command, chatID)

	switch command {
	case "start":
		b.sendMainMenu(chatID, 0)

	case "help":
		text := b.handler.HandleHelp()
		msg := tgbotapi.NewMessage(chatID, text)
		b.api.Send(msg)

	case "cars":
		b.sendCars(chatID, 0)

	case "info":
		b.sendInfo(chatID, 0)

	case "status":
		b.sendStatus(chatID, 0)

	case "battery":
		b.sendBattery(chatID, 0)

	case "charge":
		b.sendCharge(chatID, 0)

	case "drive":
		b.sendDrive(chatID, 0)

	default:
		if b.isGroupChat(message.Chat) {
			return
		}
		msg := tgbotapi.NewMessage(chatID, "❓ 未知命令，请使用 /help 查看可用命令")
		b.api.Send(msg)
	}
}

func (b *Bot) isGroupChat(chat *tgbotapi.Chat) bool {
	return chat.Type == "group" || chat.Type == "supergroup"
}

func (b *Bot) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	if !b.isAuthorized(query.Message.Chat.ID) {
		return
	}

	data := query.Data
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	log.Printf("收到回调: %s, ChatID=%d", data, chatID)

	switch {
	case strings.HasPrefix(data, "select_car_"):
		carIDStr := strings.TrimPrefix(data, "select_car_")
		carID, err := strconv.Atoi(carIDStr)
		if err != nil {
			callback := tgbotapi.NewCallback(query.ID, "❌ 无效的车辆")
			b.api.Request(callback)
			return
		}
		if err := b.validateCarID(carID); err != nil {
			callback := tgbotapi.NewCallback(query.ID, "❌ 车辆不存在")
			b.api.Request(callback)
			return
		}
		if err := b.carState.Set(chatID, carID); err != nil {
			log.Printf("保存选车状态失败: %v", err)
		}
		carName := b.getCarName(carID)
		callback := tgbotapi.NewCallback(query.ID, fmt.Sprintf("已切换到 %s", carName))
		b.api.Request(callback)
		b.sendMainMenu(chatID, messageID)

	case data == "cars":
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		b.sendCars(chatID, messageID)

	case data == "info":
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		b.sendInfo(chatID, messageID)

	case data == "status":
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		b.sendStatus(chatID, messageID)

	case data == "battery":
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		b.sendBattery(chatID, messageID)

	case data == "charge":
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		b.sendCharge(chatID, messageID)

	case data == "drive":
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		b.sendDrive(chatID, messageID)

	case data == "back_main":
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		b.sendMainMenu(chatID, messageID)

	case strings.HasPrefix(data, "refresh_"):
		callback := tgbotapi.NewCallback(query.ID, "")
		b.api.Request(callback)
		refreshType := strings.TrimPrefix(data, "refresh_")
		b.handleRefresh(chatID, messageID, refreshType)

	default:
		b.api.Request(tgbotapi.NewCallback(query.ID, "❓ 未知操作"))
	}
}

func (b *Bot) sendMainMenu(chatID int64, messageID int) {
	carID, err := b.resolveCarID(chatID)
	carName := ""
	if err == nil {
		carName = b.getCarName(carID)
	}
	text := b.handler.HandleStart(carName)
	menu := GetMainMenu(b.showCarSwitch())

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ReplyMarkup = &menu
		b.api.Send(edit)
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = menu
	b.api.Send(msg)
}

func (b *Bot) sendCars(chatID int64, messageID int) {
	carID, _ := b.resolveCarID(chatID)
	text, cars, err := b.handler.HandleCars(carID)
	if err != nil {
		text = fmt.Sprintf("❌ 获取车辆列表失败: %v", err)
		if messageID > 0 {
			edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
			b.api.Send(edit)
		} else {
			msg := tgbotapi.NewMessage(chatID, text)
			b.api.Send(msg)
		}
		return
	}

	if len(cars) == 1 && carID == 0 {
		_ = b.carState.Set(chatID, cars[0].CarID)
		carID = cars[0].CarID
	}

	menu := GetCarSelectMenu(cars, carID)
	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ReplyMarkup = &menu
		b.api.Send(edit)
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = menu
	b.api.Send(msg)
}

func (b *Bot) handleRefresh(chatID int64, messageID int, refreshType string) {
	carID, err := b.resolveCarID(chatID)
	if err != nil {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, b.carSelectErrorText())
		b.api.Send(edit)
		return
	}

	switch refreshType {
	case "info":
		text, err := b.handler.HandleInfo(carID)
		if err != nil {
			text = fmt.Sprintf("❌ 获取车辆信息失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("info")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case "status":
		text, err := b.handler.HandleStatus(carID)
		if err != nil {
			text = fmt.Sprintf("❌ 获取车辆状态失败: %s", html.EscapeString(err.Error()))
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = tgbotapi.ModeHTML
		menu := GetRefreshMenu("status")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case "battery":
		text, err := b.handler.HandleBattery(carID)
		if err != nil {
			text = fmt.Sprintf("❌ 获取电池健康度失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("battery")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case "charge":
		text, err := b.handler.HandleCharge(carID)
		if err != nil {
			text = fmt.Sprintf("❌ 获取充电记录失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("charge")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)

	case "drive":
		text, err := b.handler.HandleDrive(carID)
		if err != nil {
			text = fmt.Sprintf("❌ 获取驾驶信息失败: %v", err)
		}
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		menu := GetRefreshMenu("drive")
		edit.ReplyMarkup = &menu
		b.api.Send(edit)
	}
}

func (b *Bot) sendInfo(chatID int64, messageID int) {
	carID, err := b.resolveCarID(chatID)
	if err != nil {
		b.sendOrEditMessage(chatID, messageID, b.carSelectErrorText(), nil)
		return
	}
	text, err := b.handler.HandleInfo(carID)
	if err != nil {
		text = fmt.Sprintf("❌ 获取车辆信息失败: %v", err)
	}
	menu := GetRefreshMenu("info")
	b.sendOrEditMessage(chatID, messageID, text, &menu)
}

func (b *Bot) sendStatus(chatID int64, messageID int) {
	carID, err := b.resolveCarID(chatID)
	if err != nil {
		b.sendOrEditMessage(chatID, messageID, b.carSelectErrorText(), nil)
		return
	}
	text, err := b.handler.HandleStatus(carID)
	if err != nil {
		text = fmt.Sprintf("❌ 获取车辆状态失败: %s", html.EscapeString(err.Error()))
	}
	menu := GetRefreshMenu("status")
	b.sendOrEditMessageWithParseMode(chatID, messageID, text, &menu, tgbotapi.ModeHTML)
}

func (b *Bot) sendBattery(chatID int64, messageID int) {
	carID, err := b.resolveCarID(chatID)
	if err != nil {
		b.sendOrEditMessage(chatID, messageID, b.carSelectErrorText(), nil)
		return
	}
	text, err := b.handler.HandleBattery(carID)
	if err != nil {
		text = fmt.Sprintf("❌ 获取电池健康度失败: %v", err)
	}
	menu := GetRefreshMenu("battery")
	b.sendOrEditMessage(chatID, messageID, text, &menu)
}

func (b *Bot) sendCharge(chatID int64, messageID int) {
	carID, err := b.resolveCarID(chatID)
	if err != nil {
		b.sendOrEditMessage(chatID, messageID, b.carSelectErrorText(), nil)
		return
	}
	text, err := b.handler.HandleCharge(carID)
	if err != nil {
		text = fmt.Sprintf("❌ 获取充电记录失败: %v", err)
	}
	menu := GetRefreshMenu("charge")
	b.sendOrEditMessage(chatID, messageID, text, &menu)
}

func (b *Bot) sendDrive(chatID int64, messageID int) {
	carID, err := b.resolveCarID(chatID)
	if err != nil {
		b.sendOrEditMessage(chatID, messageID, b.carSelectErrorText(), nil)
		return
	}
	text, err := b.handler.HandleDrive(carID)
	if err != nil {
		text = fmt.Sprintf("❌ 获取驾驶信息失败: %v", err)
	}
	menu := GetRefreshMenu("drive")
	b.sendOrEditMessage(chatID, messageID, text, &menu)
}

func (b *Bot) sendOrEditMessage(chatID int64, messageID int, text string, menu *tgbotapi.InlineKeyboardMarkup) {
	b.sendOrEditMessageWithParseMode(chatID, messageID, text, menu, "")
}

func (b *Bot) sendOrEditMessageWithParseMode(chatID int64, messageID int, text string, menu *tgbotapi.InlineKeyboardMarkup, parseMode string) {
	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ReplyMarkup = menu
		if parseMode != "" {
			edit.ParseMode = parseMode
		}
		b.api.Send(edit)
		return
	}

	msg := tgbotapi.NewMessage(chatID, text)
	if menu != nil {
		msg.ReplyMarkup = *menu
	}
	if parseMode != "" {
		msg.ParseMode = parseMode
	}
	b.api.Send(msg)
}
