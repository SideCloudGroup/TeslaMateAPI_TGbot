package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pplul/teslamateapi-tgbot/bot"
	"github.com/pplul/teslamateapi-tgbot/client"
	"github.com/pplul/teslamateapi-tgbot/config"
)

var (
	version = "1.0.0"
)

func main() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 定义命令行参数
	var (
		configPath  = flag.String("config", "", "配置文件路径 (默认: ./config.toml)")
		showVersion = flag.Bool("version", false, "显示版本信息")
		showHelp    = flag.Bool("help", false, "显示帮助信息")
	)

	flag.StringVar(configPath, "c", "", "配置文件路径（简写）")
	flag.BoolVar(showVersion, "v", false, "显示版本信息（简写）")
	flag.BoolVar(showHelp, "h", false, "显示帮助信息（简写）")

	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("Tesla TeslaMate Telegram Bot v%s\n", version)
		fmt.Println("基于TeslaMate API的Telegram Bot")
		os.Exit(0)
	}

	// 显示帮助信息
	if *showHelp {
		fmt.Println("Tesla TeslaMate Telegram Bot - 使用说明")
		fmt.Println()
		fmt.Println("用法:")
		fmt.Printf("  %s [选项]\n", os.Args[0])
		fmt.Println()
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("示例:")
		fmt.Printf("  %s                              # 使用默认配置文件\n", os.Args[0])
		fmt.Printf("  %s -config /path/to/config.toml # 指定配置文件\n", os.Args[0])
		fmt.Printf("  %s -c config.toml               # 使用简写\n", os.Args[0])
		fmt.Printf("  %s -version                     # 显示版本\n", os.Args[0])
		os.Exit(0)
	}

	// 确定配置文件路径
	if *configPath == "" {
		// 获取可执行文件所在目录
		exePath, err := os.Executable()
		if err != nil {
			log.Fatalf("获取可执行文件路径失败: %v", err)
		}
		exeDir := filepath.Dir(exePath)
		*configPath = filepath.Join(exeDir, "config.toml")
	}

	log.Printf("正在加载配置文件: %s", *configPath)
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	log.Println("配置文件加载成功")

	// 初始化TeslaMate API客户端
	tmClient := client.NewClient(
		cfg.TeslaMate.APIURL,
		cfg.TeslaMate.APIKey,
		cfg.TeslaMate.CarID,
		cfg.TeslaMate.Timeout,
	)
	log.Printf("TeslaMate API客户端初始化完成 (CarID: %d)", cfg.TeslaMate.CarID)

	// 初始化Telegram Bot
	tgBot, err := bot.NewBot(
		cfg.Telegram.BotToken,
		cfg.Telegram.WhitelistChatIDs,
		cfg.Telegram.APIEndpoint,
		tmClient,
	)
	if err != nil {
		log.Fatalf("初始化Telegram Bot失败: %v", err)
	}

	log.Printf("已授权 %d 个会话使用Bot", len(cfg.Telegram.WhitelistChatIDs))

	// 启动Bot
	log.Println("Tesla Telegram Bot 启动成功!")
	if err := tgBot.Start(); err != nil {
		log.Fatalf("Bot运行错误: %v", err)
	}
}
