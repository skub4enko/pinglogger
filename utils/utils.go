package utils

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	DefaultInterval = time.Second
	DefaultTimeout  = time.Duration(1 << 63 - 1)
)

type HostConfig struct {
	Host     string
	Interval time.Duration
}

func ReadHost(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	os.Stdout.WriteString(prompt)
	os.Stdout.Sync()
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "https://")
	parts := strings.Split(input, "/")
	input = parts[0]

	if input == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: пустой хост или IP.")
		os.Exit(1)
	}

	if net.ParseIP(input) == nil {
		if !strings.Contains(input, ".") || len(input) < 2 {
			fmt.Fprintln(os.Stderr, "Ошибка: невалидный IP или домен.")
			os.Exit(1)
		}
	}
	return input
}

func WaitForEnter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Пингование запущено. Нажмите Enter для остановки...")
	reader.ReadString('\n')
}

func CalculateAvg(data []float64) float64 {
	sum := 0.0
	count := 0
	for _, v := range data {
		if v > 0 {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func CountLosses(data []float64) int {
	losses := 0
	for _, v := range data {
		if v == 0 {
			losses++
		}
	}
	return losses
}

func IsValidHost(host string) bool {
	if host == "" {
		return false
	}
	if net.ParseIP(host) != nil {
			return true
		}
	if strings.Contains(host, ".") && len(host) >= 2 {
		return true
	}
	return false
}

// ParseInterval: парсит "5 sec", "2 min", или "10" (сек по дефолту)
func ParseInterval(input string) time.Duration {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return DefaultInterval
	}

	// Разделяем число и единицу
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return DefaultInterval
	}

	valueStr := parts[0]
	unit := "sec" // дефолт
	if len(parts) > 1 {
		unit = parts[1]
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil || value <= 0 {
		fmt.Printf("Неверное число в интервале, использую дефолт (%s)\n", DefaultInterval.String())
		return DefaultInterval
	}

	var multiplier time.Duration
	switch unit {
	case "sec", "s", "second", "seconds":
		multiplier = time.Second
	case "min", "m", "minute", "minutes":
		multiplier = time.Minute
	default:
		fmt.Printf("Неверная единица '%s', использую секунды.\n", unit)
		multiplier = time.Second
	}

	duration := time.Duration(value * float64(multiplier))
	if duration < time.Second {
		fmt.Printf("Интервал слишком мал, минимум 1 сек. Использую дефолт.\n")
		return DefaultInterval
	}
	return duration
}

// ReadInterval: теперь читает строку и парсит
func ReadInterval(prompt string) time.Duration {
	reader := bufio.NewReader(os.Stdin)
	os.Stdout.WriteString(prompt)
	os.Stdout.Sync()
	input, _ := reader.ReadString('\n')
	return ParseInterval(input)
}

// ReadHostsFromFile: интервал в минутах как число (для совместимости)
func ReadHostsFromFile(filename string) ([]HostConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []HostConfig
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		host := parts[0]
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		hostParts := strings.Split(host, "/")
		host = hostParts[0]

		if !IsValidHost(host) {
			fmt.Printf("Пропущен невалидный хост: %s\n", host)
			continue
		}

		interval := DefaultInterval
		if len(parts) > 1 {
			// Здесь число в минутах для файла
			minutes, err := strconv.ParseFloat(parts[1], 64)
			if err == nil && minutes >= 0.0167 { // мин 1 сек = 0.0167 мин
				interval = time.Duration(minutes * float64(time.Minute))
			} else {
				fmt.Printf("Неверный интервал для %s, использую дефолт\n", host)
			}
		}
		hosts = append(hosts, HostConfig{Host: host, Interval: interval})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return hosts, nil
}