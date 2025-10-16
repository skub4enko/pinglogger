package graphplotter

import (
	"fmt"
	"math"

	"github.com/guptarohit/asciigraph"
)

func GenerateGraph(rttData []float64) string {
	if len(rttData) == 0 {
		return "Нет данных для графика."
	}

	// Фильтр: игнорируем 0 для масштаба (потери внизу)
	validData := make([]float64, len(rttData))
	copy(validData, rttData)
	minRTT, maxRTT := float64(math.MaxFloat64), 0.0
	allZero := true
	for _, v := range rttData {
		if v > 0 {
			allZero = false
			if v < minRTT {
				minRTT = v
			}
			if v > maxRTT {
				maxRTT = v
			}
		}
	}
	if allZero {
		return "Все пинги потеряны или ошибки. График пуст."
	}

	// Проверка на плоскую линию
	isFlat := minRTT == maxRTT
	if isFlat {
		return fmt.Sprintf("График RTT (мс) по времени. Стабильный RTT (%.0f мс) — плоская линия без вариаций.\n%s", maxRTT, asciigraph.Plot(validData,
			asciigraph.Precision(0),
			asciigraph.Height(5), // Меньше высота для плоских
			asciigraph.Width(80),
			asciigraph.Offset(4),
		))
	}

	// Обычный график с caption сверху
	caption := "График RTT (мс) по времени (попытки). 0 мс = потери пакетов или ошибки."
	plot := asciigraph.Plot(validData,
		asciigraph.Precision(0),
		asciigraph.Height(10), // Уменьшил для компактности
		asciigraph.Width(80),
		asciigraph.Offset(4),
		asciigraph.Caption(""), // Убрал отсюда
	)

	return fmt.Sprintf("%s\nMin: %.0f мс | Max: %.0f мс\n%s", caption, minRTT, maxRTT, plot)
}