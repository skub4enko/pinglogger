package pinger

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

type PingerConfig struct {
	Host     string
	Interval time.Duration
	Timeout  time.Duration
}

// Экспортируем тип: большая буква P
type PingResult struct {
	Timestamp time.Time
	RTT       float64
}

func RunLoop(config PingerConfig, stopChan chan struct{}, rttData *[]PingResult, mu *sync.Mutex) {
	pinger, err := ping.NewPinger(config.Host)
	if err != nil {
		fmt.Printf("Ошибка создания пингера: %v\nПодсказка: Запустите от админа.\n", err)
		return
	}
	pinger.Count = 1
	pinger.Interval = config.Interval
	pinger.Timeout = config.Timeout
	pinger.SetPrivileged(true)

	index := 1
	for {
		select {
		case <-stopChan:
			pinger.Stop()
			return
		default:
			startTime := time.Now()
			err := pinger.Run()
			pingTime := time.Since(startTime)

			timestamp := time.Now()
			mu.Lock()
			if err != nil {
				fmt.Printf("Ошибка пинга %d: %v\n", index, err)
				*rttData = append(*rttData, PingResult{Timestamp: timestamp, RTT: 0}) // Большая P
			} else {
				stats := pinger.Statistics()
				if stats.PacketsRecv == 1 {
					rtt := float64(stats.AvgRtt.Milliseconds())
					*rttData = append(*rttData, PingResult{Timestamp: timestamp, RTT: rtt}) // Большая P
					fmt.Printf("Пинг %d: %.0f мс\n", index, rtt)
				} else {
					*rttData = append(*rttData, PingResult{Timestamp: timestamp, RTT: 0}) // Большая P
					fmt.Printf("Пинг %d: потеря пакета\n", index)
				}
			}
			mu.Unlock()
			index++

			sleepTime := config.Interval - pingTime
			if sleepTime > 0 {
				time.Sleep(sleepTime)
			}
		}
	}
}