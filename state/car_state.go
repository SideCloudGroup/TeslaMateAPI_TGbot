package state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// DefaultStatePath 选车状态文件路径（固定，不可配置）
const DefaultStatePath = "data/car_state.json"

type fileData struct {
	Selections map[string]int `json:"selections"`
}

// CarStateStore 按 chat 持久化车辆选择
type CarStateStore struct {
	path      string
	mu        sync.RWMutex
	byChat    map[int64]int
	defaultID int
}

// NewCarStateStore 加载或创建选车状态文件
func NewCarStateStore(defaultCarID int) (*CarStateStore, error) {
	s := &CarStateStore{
		path:      DefaultStatePath,
		byChat:    make(map[int64]int),
		defaultID: defaultCarID,
	}

	if err := ensureParentDir(s.path); err != nil {
		return nil, fmt.Errorf("创建选车状态目录失败: %w", err)
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, s.persist()
		}
		return nil, fmt.Errorf("读取选车状态文件失败: %w", err)
	}

	var fd fileData
	if err := json.Unmarshal(data, &fd); err != nil {
		log.Printf("解析选车状态文件失败，将以空状态启动: %v", err)
		return s, nil
	}

	if fd.Selections != nil {
		for chatStr, carID := range fd.Selections {
			chatID, err := strconv.ParseInt(chatStr, 10, 64)
			if err != nil {
				log.Printf("跳过无效 chat_id: %s", chatStr)
				continue
			}
			s.byChat[chatID] = carID
		}
	}

	return s, nil
}

// DefaultID 返回配置的默认车辆 ID
func (s *CarStateStore) DefaultID() int {
	return s.defaultID
}

// Get 获取 chat 已保存的车辆 ID
func (s *CarStateStore) Get(chatID int64) (carID int, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	carID, ok = s.byChat[chatID]
	return carID, ok
}

// Set 保存 chat 的车辆选择并落盘
func (s *CarStateStore) Set(chatID int64, carID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byChat[chatID] = carID
	return s.persistLocked()
}

// Clear 清除 chat 的车辆选择
func (s *CarStateStore) Clear(chatID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byChat, chatID)
	return s.persistLocked()
}

func (s *CarStateStore) persist() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.persistLocked()
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func (s *CarStateStore) persistLocked() error {
	if err := ensureParentDir(s.path); err != nil {
		return fmt.Errorf("创建选车状态目录失败: %w", err)
	}

	fd := fileData{Selections: make(map[string]int, len(s.byChat))}
	for chatID, carID := range s.byChat {
		fd.Selections[strconv.FormatInt(chatID, 10)] = carID
	}

	data, err := json.MarshalIndent(fd, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化选车状态失败: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		log.Printf("写入选车状态文件失败: %v", err)
		return err
	}

	return nil
}
