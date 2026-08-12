package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sync"
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

const carsCacheTTL = 60 * time.Second

const minimumLatestDriveDistanceKM = 0.5

// Client TeslaMate API客户端
type Client struct {
	baseURL    string
	apiKey     string
	headers    map[string]string
	httpClient *fasthttp.Client

	carsMu        sync.RWMutex
	carsCache     []models.Car
	carsCacheTime time.Time
}

// NewClient 创建新的TeslaMate API客户端
func NewClient(baseURL, apiKey string, timeout int, headers map[string]string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
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

	url := c.baseURL + path
	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	if c.headers != nil {
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}
	}

	if err := c.httpClient.Do(req, resp); err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	statusCode := resp.StatusCode()
	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", statusCode, string(resp.Body()))
	}

	body := make([]byte, len(resp.Body()))
	copy(body, resp.Body())

	if err := checkAPIError(body); err != nil {
		return nil, err
	}

	return body, nil
}

func checkAPIError(body []byte) error {
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil
	}
	if errResp.Error != "" {
		return fmt.Errorf("API错误: %s", errResp.Error)
	}
	return nil
}

// GetCars 获取车辆列表（带 60s 内存缓存）
func (c *Client) GetCars() ([]models.Car, error) {
	c.carsMu.RLock()
	if c.carsCache != nil && time.Since(c.carsCacheTime) < carsCacheTTL {
		cars := make([]models.Car, len(c.carsCache))
		copy(cars, c.carsCache)
		c.carsMu.RUnlock()
		return cars, nil
	}
	c.carsMu.RUnlock()

	body, err := c.doRequest("GET", "/api/v1/cars")
	if err != nil {
		return nil, fmt.Errorf("获取车辆列表失败: %w", err)
	}

	var response models.CarResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析车辆列表失败: %w", err)
	}

	cars := response.Data.Cars
	if len(cars) == 0 {
		return nil, fmt.Errorf("未找到可用车辆")
	}

	c.carsMu.Lock()
	c.carsCache = make([]models.Car, len(cars))
	copy(c.carsCache, cars)
	c.carsCacheTime = time.Now()
	c.carsMu.Unlock()

	result := make([]models.Car, len(cars))
	copy(result, cars)
	return result, nil
}

// GetCarDetails 获取车辆详细信息
func (c *Client) GetCarDetails(carID int) (*models.Car, error) {
	path := fmt.Sprintf("/api/v1/cars/%d", carID)
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
func (c *Client) GetCarStatus(carID int) (*models.StatusResponse, error) {
	path := fmt.Sprintf("/api/v1/cars/%d/status", carID)
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
func (c *Client) GetBatteryHealth(carID int) (*models.BatteryHealthResponse, error) {
	path := fmt.Sprintf("/api/v1/cars/%d/battery-health", carID)
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
func (c *Client) GetLatestCharge(carID int) (*models.Charge, error) {
	path := fmt.Sprintf("/api/v1/cars/%d/charges?show=1", carID)
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

	chargeID := response.Data.Charges[0].ChargeID
	detailPath := fmt.Sprintf("/api/v1/cars/%d/charges/%d", carID, chargeID)
	detailBody, err := c.doRequest("GET", detailPath)
	if err != nil {
		return nil, fmt.Errorf("获取充电详情失败: %w", err)
	}

	var detailResponse models.ChargeDetailsResponse
	if err := json.Unmarshal(detailBody, &detailResponse); err != nil {
		return nil, fmt.Errorf("解析充电详情失败: %w", err)
	}

	charge := detailResponse.Data.Charge
	charge.ElectricalStats = summarizeChargeElectricalStats(charge.ChargeDetails)
	return &charge, nil
}

func summarizeChargeElectricalStats(details []models.ChargeDetail) models.ChargeElectricalStats {
	var stats models.ChargeElectricalStats
	var voltageCount, currentCount, powerCount int

	for _, detail := range details {
		charger := detail.ChargerDetails
		if charger.ChargerVoltage > 0 {
			stats.AverageVoltage += charger.ChargerVoltage
			voltageCount++
			if charger.ChargerVoltage > stats.MaximumVoltage {
				stats.MaximumVoltage = charger.ChargerVoltage
			}
		}
		if charger.ChargerActualCurrent > 0 {
			stats.AverageCurrent += charger.ChargerActualCurrent
			currentCount++
			if charger.ChargerActualCurrent > stats.MaximumCurrent {
				stats.MaximumCurrent = charger.ChargerActualCurrent
			}
		}
		if charger.ChargerPower > 0 {
			stats.AveragePower += charger.ChargerPower
			powerCount++
			if charger.ChargerPower > stats.MaximumPower {
				stats.MaximumPower = charger.ChargerPower
			}
		}
	}

	if voltageCount > 0 {
		stats.AverageVoltage /= float64(voltageCount)
	}
	if currentCount > 0 {
		stats.AverageCurrent /= float64(currentCount)
	}
	if powerCount > 0 {
		stats.AveragePower /= float64(powerCount)
	}

	return stats
}

func (c *Client) getDrives(carID int, startDate, endDate string) (*models.DrivesResponse, error) {
	path := fmt.Sprintf(
		"/api/v1/cars/%d/drives?startDate=%s&endDate=%s",
		carID,
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

// GetLatestDrive 获取最近一次里程不少于 0.5 km 的驾驶记录（默认查询最近 7 天）
func (c *Client) GetLatestDrive(carID int) (*models.Drive, *models.Units, error) {
	startDate := time.Now().Add(-7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	endDate := time.Now().UTC().Format(time.RFC3339)
	response, err := c.getDrives(carID, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}

	if len(response.Data.Drives) == 0 {
		return nil, nil, fmt.Errorf("7天内暂无驾驶记录")
	}

	drive := latestDriveAtLeast(response.Data.Drives, response.Data.Units.UnitOfLength, minimumLatestDriveDistanceKM)
	if drive != nil {
		return drive, &response.Data.Units, nil
	}

	return nil, nil, fmt.Errorf("7天内暂无里程不少于0.5 km的驾驶记录")
}

func latestDriveAtLeast(drives []models.Drive, unitOfLength string, minimumDistanceKM float64) *models.Drive {
	for i := range drives {
		if driveDistanceInKM(drives[i].OdometerDetails.OdometerDistance, unitOfLength) >= minimumDistanceKM {
			return &drives[i]
		}
	}
	return nil
}

func driveDistanceInKM(distance float64, unitOfLength string) float64 {
	if unitOfLength == "mi" {
		return distance * 1.609344
	}
	return distance
}

// GetTodayDriveDistance 返回今日（本地时区）行程总里程与次数
func (c *Client) GetTodayDriveDistance(carID int) (float64, int, *models.Units, error) {
	now := time.Now().In(localLoc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, localLoc)
	startDate := startOfDay.Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	response, err := c.getDrives(carID, startDate, endDate)
	if err != nil {
		return 0, 0, nil, err
	}

	var totalDistance float64
	for _, drive := range response.Data.Drives {
		totalDistance += drive.OdometerDetails.OdometerDistance
	}

	return totalDistance, len(response.Data.Drives), &response.Data.Units, nil
}
