package helper

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/bytedance/gopkg/util/gopool"

	"github.com/gin-gonic/gin"
)

const (
	InitialScannerBufferSize = 64 << 10 // 64KB (64*1024)
	MaxScannerBufferSize     = 10 << 20 // 10MB (10*1024*1024)
	DefaultPingInterval      = 10 * time.Second
)

func isTerminalStreamData(data string) bool {
	if strings.HasPrefix(data, "[DONE]") {
		return true
	}
	var event struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return false
	}
	return event.Type == "message_stop"
}

func StreamScannerHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, dataHandler func(data string) bool) {

	if resp == nil || dataHandler == nil {
		return
	}

	// 确保响应体总是被关闭
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	var (
		stopChan   = make(chan bool, 3) // 增加缓冲区避免阻塞
		scanner    = bufio.NewScanner(resp.Body)
		pingTicker *time.Ticker
		writeMutex sync.Mutex     // Mutex to protect concurrent writes
		wg         sync.WaitGroup // 用于等待所有 goroutine 退出
	)

	generalSettings := operation_setting.GetGeneralSetting()
	pingEnabled := generalSettings.PingIntervalEnabled && !info.DisablePing
	pingInterval := time.Duration(generalSettings.PingIntervalSeconds) * time.Second
	if pingInterval <= 0 {
		pingInterval = DefaultPingInterval
	}

	if pingEnabled {
		pingTicker = time.NewTicker(pingInterval)
	}

	if common.DebugEnabled {
		// print timeout and ping interval for debugging
		println("relay timeout seconds:", common.RelayTimeout)
		println("ping interval seconds:", int64(pingInterval.Seconds()))
	}

	// 改进资源清理，确保所有 goroutine 正确退出
	defer func() {
		// 通知所有 goroutine 停止
		common.SafeSendBool(stopChan, true)

		if pingTicker != nil {
			pingTicker.Stop()
		}

		if resp.Body != nil {
			_ = resp.Body.Close()
		}

		// 等待所有 goroutine 退出，最多等待5秒
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			logger.LogError(c, "timeout waiting for goroutines to exit")
		}

		close(stopChan)
	}()

	scanner.Buffer(make([]byte, InitialScannerBufferSize), MaxScannerBufferSize)
	scanner.Split(bufio.ScanLines)
	SetEventStreamHeaders(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, "stop_chan", stopChan)

	// Handle ping data sending with improved error handling
	if pingEnabled && pingTicker != nil {
		wg.Add(1)
		gopool.Go(func() {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					logger.LogError(c, fmt.Sprintf("ping goroutine panic: %v", r))
					common.SafeSendBool(stopChan, true)
				}
				if common.DebugEnabled {
					println("ping goroutine exited")
				}
			}()

			for {
				select {
				case <-pingTicker.C:
					// 使用超时机制防止写操作阻塞
					done := make(chan error, 1)
					go func() {
						writeMutex.Lock()
						defer writeMutex.Unlock()
						done <- PingData(c)
					}()

					select {
					case err := <-done:
						if err != nil {
							logger.LogError(c, "ping data error: "+err.Error())
							return
						}
						if common.DebugEnabled {
							println("ping data sent")
						}
					case <-time.After(10 * time.Second):
						logger.LogError(c, "ping data send timeout")
						return
					case <-ctx.Done():
						return
					case <-stopChan:
						return
					}
				case <-ctx.Done():
					return
				case <-stopChan:
					return
				case <-c.Request.Context().Done():
					// 监听客户端断开连接
					return
				}
			}
		})
	}

	// Scanner goroutine with improved error handling
	wg.Add(1)
	common.RelayCtxGo(ctx, func() {
		defer func() {
			wg.Done()
			if r := recover(); r != nil {
				logger.LogError(c, fmt.Sprintf("scanner goroutine panic: %v", r))
			}
			common.SafeSendBool(stopChan, true)
			if common.DebugEnabled {
				println("scanner goroutine exited")
			}
		}()

		for scanner.Scan() {
			// 检查是否需要停止
			select {
			case <-stopChan:
				return
			case <-ctx.Done():
				return
			case <-c.Request.Context().Done():
				return
			default:
			}

			data := scanner.Text()
			if common.DebugEnabled {
				println(data)
			}

			if strings.HasPrefix(data, "[DONE]") {
				if common.DebugEnabled {
					println("received [DONE], stopping scanner")
				}
				return
			}
			if len(data) < 6 {
				continue
			}
			if !strings.HasPrefix(data, "data:") {
				continue
			}
			data = data[5:]
			data = strings.TrimLeft(data, " ")
			data = strings.TrimSuffix(data, "\r")
			if strings.HasPrefix(data, "[DONE]") {
				if common.DebugEnabled {
					println("received [DONE], stopping scanner")
				}
				return
			}
			info.SetFirstResponseTime()

			// 使用超时机制防止写操作阻塞
			done := make(chan bool, 1)
			go func() {
				writeMutex.Lock()
				defer writeMutex.Unlock()
				done <- dataHandler(data)
			}()

			select {
			case success := <-done:
				if !success {
					return
				}
			case <-time.After(10 * time.Second):
				logger.LogError(c, "data handler timeout")
				return
			case <-ctx.Done():
				return
			case <-stopChan:
				return
			}
			if isTerminalStreamData(data) {
				if common.DebugEnabled {
					println("received terminal stream data, stopping scanner")
				}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if err != io.EOF {
				logger.LogError(c, "scanner error: "+err.Error())
			}
		}
	})

	// 主循环等待完成或客户端断开，不再用空闲计时器主动关闭上游流。
	select {
	case <-stopChan:
		// 正常结束
		logger.LogInfo(c, "streaming finished")
	case <-c.Request.Context().Done():
		// 客户端断开连接
		logger.LogInfo(c, "client disconnected")
	}
}
