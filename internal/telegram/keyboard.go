package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var timezoneRegions = map[string][]string{
	"Europe": {
		"Europe/Moscow", "Europe/London", "Europe/Berlin",
		"Europe/Paris", "Europe/Rome", "Europe/Madrid",
		"Europe/Kiev", "Europe/Warsaw", "Europe/Istanbul",
		"Europe/Minsk", "Europe/Bucharest", "Europe/Helsinki",
	},
	"Asia": {
		"Asia/Tokyo", "Asia/Shanghai", "Asia/Kolkata",
		"Asia/Dubai", "Asia/Bangkok", "Asia/Singapore",
		"Asia/Seoul", "Asia/Taipei", "Asia/Jakarta",
		"Asia/Almaty", "Asia/Tashkent", "Asia/Tbilisi",
	},
	"Americas": {
		"America/New_York", "America/Chicago", "America/Denver",
		"America/Los_Angeles", "America/Toronto", "America/Sao_Paulo",
		"America/Mexico_City", "America/Argentina/Buenos_Aires",
	},
	"Pacific": {
		"Pacific/Auckland", "Pacific/Fiji", "Pacific/Honolulu",
		"Australia/Sydney", "Australia/Melbourne", "Australia/Perth",
	},
}

var regionOrder = []string{"Europe", "Asia", "Americas", "Pacific"}

var weekdayNames = []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}

func TimezoneRegionsKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, region := range regionOrder {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(region, "tz_region:"+region),
		))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func TimezoneCitiesKeyboard(region string) tgbotapi.InlineKeyboardMarkup {
	cities, ok := timezoneRegions[region]
	if !ok {
		return tgbotapi.NewInlineKeyboardMarkup()
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(cities); i += 2 {
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(cities[i], "tz:"+cities[i]),
		}
		if i+1 < len(cities) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(cities[i+1], "tz:"+cities[i+1]))
		}
		rows = append(rows, row)
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "tz_region:back"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func QuotesCountKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1", "count:1"),
			tgbotapi.NewInlineKeyboardButtonData("2", "count:2"),
			tgbotapi.NewInlineKeyboardButtonData("3", "count:3"),
		),
	)
}

func WeekdaysKeyboard(selected []int) tgbotapi.InlineKeyboardMarkup {
	isSelected := make(map[int]bool)
	for _, d := range selected {
		isSelected[d] = true
	}

	var row1, row2 []tgbotapi.InlineKeyboardButton
	for i := 1; i <= 6; i++ {
		label := weekdayNames[i]
		if isSelected[i] {
			label = "✅ " + label
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("day:%d", i))
		if i <= 3 {
			row1 = append(row1, btn)
		} else {
			row2 = append(row2, btn)
		}
	}

	sundayLabel := weekdayNames[0]
	if isSelected[0] {
		sundayLabel = "✅ " + sundayLabel
	}
	row3 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(sundayLabel, "day:0"),
	)

	doneRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("✅ Готово", "days_done"),
	)

	return tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, doneRow)
}

func HourKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for startHour := 0; startHour < 24; startHour += 6 {
		var row []tgbotapi.InlineKeyboardButton
		for h := startHour; h < startHour+6; h++ {
			label := fmt.Sprintf("%02d", h)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("hour:%02d", h)))
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func MinuteKeyboard() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for startMin := 0; startMin < 60; startMin += 30 {
		var row []tgbotapi.InlineKeyboardButton
		for m := startMin; m < startMin+30; m += 5 {
			label := fmt.Sprintf("%02d", m)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("minute:%02d", m)))
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func SkipKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏭ Пропустить", "skip"),
		),
	)
}
