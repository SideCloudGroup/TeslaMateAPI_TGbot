package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"teslamate-bot/models"

	"github.com/valyala/fasthttp"
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

// Client TeslaMate API客户端
type Client struct {
	baseURL    string
	apiKey     string
	carID      int
	headers    map[string]string
	httpClient *fasthttp.Client
}

// NewClient 创建新的TeslaMate API客户端
func NewClient(baseURL, apiKey string, carID, timeout int, headers map[string]string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		carID:   carID,
		headers: headers,
		httpClient: &fasthttp.Client{
			ReadTimeout:  time.Duration(timeout) * time.Second,
			WriteTimeout: time.Duration(timeout) * time.Second,
		},
	}
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(method, path string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	// 构建完整URL
	url := c.baseURL + path
	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	// 设置认证头（api_key 为空时不发送）
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// 设置自定义请求头
	if c.headers != nil {
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}
	}

	// 执行请求
	if err := c.httpClient.Do(req, resp); err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	// 检查状态码
	statusCode := resp.StatusCode()
	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", statusCode, string(resp.Body()))
	}

	// 复制响应体
	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	return body, nil
}

// GetCarDetails 获取车辆详细信息
func (c *Client) GetCarDetails() (*models.Car, error) {
	path := fmt.Sprintf("/api/v1/cars/%d", c.carID)
	body, err := c.doRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("获取车辆详情失败: %w", err)
	}

	var response models.CarResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析车辆详情失败: %w", err)
	}

	if len(response.Data.Cars) == 0 {
		return nil, fmt.Errorf("未找到车辆信息")
	}

	return &response.Data.Cars[0], nil
}

// GetCarStatus 获取车辆当前状态
func (c *Client) GetCarStatus() (*models.StatusResponse, error) {
	path := fmt.Sprintf("/api/v1/cars/%d/status", c.carID)
	body, err := c.doRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("获取车辆状态失败: %w", err)
	}

	var response models.StatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析车辆状态失败: %w", err)
	}

	return &response, nil
}

// GetBatteryHealth 获取电池健康度
func (c *Client) GetBatteryHealth() (*models.BatteryHealthResponse, error) {
	path := fmt.Sprintf("/api/v1/cars/%d/battery-health", c.carID)
	body, err := c.doRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("获取电池健康度失败: %w", err)
	}

	var response models.BatteryHealthResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析电池健康度失败: %w", err)
	}

	return &response, nil
}

// GetLatestCharge 获取最新充电记录
func (c *Client) GetLatestCharge() (*models.Charge, error) {
	path := fmt.Sprintf("/api/v1/cars/%d/charges", c.carID)
	body, err := c.doRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("获取充电记录失败: %w", err)
	}

	var response models.ChargesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析充电记录失败: %w", err)
	}

	if len(response.Data.Charges) == 0 {
		return nil, fmt.Errorf("暂无充电记录")
	}

	// 返回最新的充电记录（第一条）
	return &response.Data.Charges[0], nil
}

func (c *Client) getDrives(startDate, endDate string) (*models.DrivesResponse, error) {
	path := fmt.Sprintf(
		"/api/v1/cars/%d/drives?startDate=%s&endDate=%s",
		c.carID,
		url.QueryEscape(startDate),
		url.QueryEscape(endDate),
	)
	body, err := c.doRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("获取驾驶记录失败: %w", err)
	}

	var response models.DrivesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析驾驶记录失败: %w", err)
	}

	return &response, nil
}

// GetLatestDrive 获取最近一次驾驶记录（默认 7 天内最后一条）
func (c *Client) GetLatestDrive() (*models.Drive, *models.Units, error) {
	startDate := time.Now().Add(-7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	endDate := time.Now().UTC().Format(time.RFC3339)
	response, err := c.getDrives(startDate, endDate)
	if err != nil {
		return nil, nil, err
	}

	if len(response.Data.Drives) == 0 {
		return nil, nil, fmt.Errorf("7天内暂无驾驶记录")
	}

	return &response.Data.Drives[0], &response.Data.Units, nil
}

// GetTodayDriveDistance 返回今日（本地时区）行程总里程与次数
func (c *Client) GetTodayDriveDistance() (float64, int, *models.Units, error) {
	now := time.Now().In(localLoc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, localLoc)
	startDate := startOfDay.Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	response, err := c.getDrives(startDate, endDate)
	if err != nil {
		return 0, 0, nil, err
	}

	var totalDistance float64
	for _, drive := range response.Data.Drives {
		totalDistance += drive.OdometerDetails.OdometerDistance
	}

	return totalDistance, len(response.Data.Drives), &response.Data.Units, nil
}
