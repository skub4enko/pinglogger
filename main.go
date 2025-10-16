package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"ping_logger/pinger"
	"ping_logger/graphplotter"
	"ping_logger/utils"
)

func main() {
	mode := chooseMode()
	var hosts []utils.HostConfig
	switch mode {
	case 1:
		host := utils.ReadHost("Введите IP или домен: ")
		interval := utils.ReadInterval("Интервал (например, 5 sec или 2 min, Enter для 1 sec): ")
		hosts = []utils.HostConfig{{Host: host, Interval: interval}}
	case 2:
		hosts = readHostsFromInput()
	case 3:
		var err error
		hostsFile := "targetservers.txt"
		hosts, err = utils.ReadHostsFromFile(hostsFile)
		if err != nil || len(hosts) == 0 {
			fmt.Printf("Файл %s не найден или пуст — переключаюсь на ручной ввод.\n", hostsFile)
			host := utils.ReadHost("Введите IP или домен: ")
			interval := utils.ReadInterval("Интервал (например, 5 sec или 2 min, Enter для 1 sec): ")
			hosts = []utils.HostConfig{{Host: host, Interval: interval}}
		}
	}

	fmt.Printf("Пингую %d серверов: %v\n", len(hosts), hosts)

	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		close(stopChan)
	}()

	for _, h := range hosts {
		wg.Add(1)
		go func(hc utils.HostConfig) {
			defer wg.Done()
			pingHost(hc.Host, hc.Interval, stopChan)
		}(h)
	}

	utils.WaitForEnter()
	close(stopChan)
	wg.Wait()
	fmt.Println("Все пинги остановлены.")
}

func chooseMode() int {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Выберите режим:")
		fmt.Println("1) Ввести один адрес")
		fmt.Println("2) Ввести несколько адресов (через запятую и пробел)")
		fmt.Println("3) Использовать адреса из файла targetservers.txt")
		os.Stdout.WriteString("Ваш выбор (1-3): ")
		os.Stdout.Sync()
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		switch input {
		case "1":
			return 1
		case "2":
			return 2
		case "3":
			return 3
		default:
			fmt.Println("Неверный выбор. Попробуйте снова.")
		}
	}
}

func readHostsFromInput() []utils.HostConfig {
	reader := bufio.NewReader(os.Stdin)
	os.Stdout.WriteString("Введите адреса через запятую и пробел (например, 8.8.8.8, google.com): ")
	os.Stdout.Sync()
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var hosts []utils.HostConfig
	rawHosts := strings.Split(input, ", ")
	for _, h := range rawHosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		h = strings.TrimPrefix(h, "http://")
		h = strings.TrimPrefix(h, "https://")
		parts := strings.Split(h, "/")
		h = parts[0]
		if h != "" && utils.IsValidHost(h) {
			interval := utils.ReadInterval(fmt.Sprintf("Интервал для %s (например, 5 sec или 2 min, Enter для 1 sec): ", h))
			hosts = append(hosts, utils.HostConfig{Host: h, Interval: interval})
		} else {
			fmt.Printf("Пропущен невалидный хост: %s\n", h)
		}
	}
	if len(hosts) == 0 {
		fmt.Println("Нет валидных хостов — переключаюсь на ручной ввод.")
		host := utils.ReadHost("Введите IP или домен: ")
		interval := utils.ReadInterval("Интервал (например, 5 sec или 2 min, Enter для 1 sec): ")
		return []utils.HostConfig{{Host: host, Interval: interval}}
	}
	return hosts
}

func pingHost(host string, interval time.Duration, stopChan chan struct{}) {
	rttData := make([]pinger.PingResult, 0)
	var mu sync.Mutex

	config := pinger.PingerConfig{
		Host:     host,
		Interval: interval,
		Timeout:  utils.DefaultTimeout,
	}

	localWg := sync.WaitGroup{}
	localWg.Add(1)
	go func() {
		defer localWg.Done()
		pinger.RunLoop(config, stopChan, &rttData, &mu)
	}()
	localWg.Wait()

	mu.Lock()
	if len(rttData) == 0 {
		fmt.Printf("[%s] Нет данных.\n", host)
		mu.Unlock()
		return
	}

	rttValues := make([]float64, len(rttData))
	for i, res := range rttData {
		rttValues[i] = res.RTT
	}
	avg := utils.CalculateAvg(rttValues)
	losses := utils.CountLosses(rttValues)
	lossRate := float64(losses) / float64(len(rttData)) * 100
	fmt.Printf("[%s] Статистика: Средний RTT: %.2f мс | Потерь: %d из %d (%.2f%%)\n", host, avg, losses, len(rttData), lossRate)

	filename := fmt.Sprintf("logs/ping_log_%s.csv", strings.ReplaceAll(host, ".", "_"))
	logToCSV(rttData, filename)
	fmt.Printf("[%s] Лог в %s\n", host, filename)

	graph := graphplotter.GenerateGraph(rttValues)
	fmt.Printf("[%s] График:\n%s\n", host, graph)
	mu.Unlock()
}

func logToCSV(data []pinger.PingResult, filename string) {
	dir := "logs"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Ошибка создания папки %s: %v. Сохраняю в корень.\n", dir, err)
		filename = strings.TrimPrefix(filename, dir+"/")
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Ошибка создания файла лога: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Timestamp", "RttMs"})

	for _, res := range data {
		timeStr := res.Timestamp.Format(time.RFC3339)
		rttStr := strconv.FormatFloat(res.RTT, 'f', 0, 64)
		if res.RTT == 0 {
			rttStr = "loss"
		}
		writer.Write([]string{timeStr, rttStr})
	}
}