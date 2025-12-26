package excel

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain"
	"github.com/DjaPy/fot-twenty-readers-go/internal/kathismas/domain/services"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/xuri/excelize/v2"
)

const FontTrebuchet = "Trebuchet MS"

func addKathismaNumbersToXLS(xls *excelize.File, number int, sheetName string) error {
	style, err := xls.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 3},
			{Type: "top", Color: "000000", Style: 3},
			{Type: "bottom", Color: "000000", Style: 3},
			{Type: "right", Color: "000000", Style: 3},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Family: FontTrebuchet, Bold: true, Size: 16},
	})
	if err != nil {
		return fmt.Errorf("failed create new style for xls %v", err)
	}
	var errs []error
	err1 := xls.SetCellValue(sheetName, "A2", strconv.Itoa(number))
	if err1 != nil {
		errs = append(errs, err1)
	}
	err2 := xls.SetCellStyle(sheetName, "A2", "A2", style)
	if err2 != nil {
		errs = append(errs, err2)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed work with cell %v", errors.Join(errs...))
	}
	return nil
}

func addHeaderOfMonthToWs(xls *excelize.File, sheetName string) error {
	cellAddressMonth := map[string]string{
		"B2": "ЯНВ", "C2": "ФЕВ", "D2": "МАРТ", "E2": "АПР",
		"F2": "МАЙ", "G2": "ИЮН", "H2": "ИЮЛ", "I2": "АВГ",
		"J2": "СЕН", "K2": "ОКТ", "L2": "НОЯ", "M2": "ДЕК",
	}
	style, err := xls.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 0},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Family: "Calibri", Color: "FF8080", Size: 16},
	})
	if err != nil {
		return fmt.Errorf("failed creat new style %v", err)
	}
	var errs []error
	for k, v := range cellAddressMonth {
		err1 := xls.SetCellValue(sheetName, k, v)
		if err1 != nil {
			errs = append(errs, err1)
			continue
		}
		err2 := xls.SetCellStyle(sheetName, k, k, style)
		if err2 != nil {
			errs = append(errs, err1)
			continue
		}
	}
	return errors.Join(errs...)
}

func addColumnWithNumberDayToWs(xls *excelize.File, sheetName string) error {
	style, _ := xls.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Family: FontTrebuchet, Size: 12},
	})
	var errs []error
	for number := 1; number <= 31; number++ {
		numberCell := number + 2
		cellNameLeft := fmt.Sprintf("A%d", numberCell)
		cellNameRight := fmt.Sprintf("N%d", numberCell)
		err1 := xls.SetCellValue(sheetName, cellNameLeft, strconv.Itoa(number))
		if err1 != nil {
			errs = append(errs, err1)
		}
		err2 := xls.SetCellValue(sheetName, cellNameRight, strconv.Itoa(number))
		if err2 != nil {
			errs = append(errs, err2)
		}
		err3 := xls.SetCellStyle(sheetName, cellNameLeft, cellNameLeft, style)
		if err3 != nil {
			errs = append(errs, err3)
		}
		err4 := xls.SetCellStyle(sheetName, cellNameRight, cellNameRight, style)
		if err4 != nil {
			errs = append(errs, err4)
		}
	}
	return errors.Join(errs...)
}

func getFrameNumberDay(symbol string, start, end int) map[int]string {
	frameNumberDay := make(map[int]string)
	for num := start; num <= end; num++ {
		frameNumberDay[num] = symbol + strconv.Itoa(num)
	}
	return frameNumberDay
}

func CreateCalendarForReaderToXLS(
	xls *excelize.File,
	calendarTable map[int][]int,
	allKathisma map[int]int,
	year int,
	sheetName string,
) error {
	style, _ := xls.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Family: FontTrebuchet, Size: 14, Color: "000000"},
	})

	pinkStyle, _ := xls.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Font:      &excelize.Font{Family: FontTrebuchet, Size: 14, Color: "000000"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FF8080"}},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 4},
			{Type: "bottom", Color: "000000", Style: 4},
		},
	})

	cellStep := 1
	frameMonth := map[int]string{
		1: "B", 2: "C", 3: "D", 4: "E", 5: "F", 6: "G", 7: "H", 8: "I", 9: "J", 10: "K", 11: "L", 12: "M",
	}
	frameNumberDayA := getFrameNumberDay("A", 3, 33) // A = 1
	frameNumberDayN := getFrameNumberDay("N", 3, 33) // N = 1
	var errs []error
	textErr := "failed create calendar for reader %v"
	for num := range frameNumberDayN {
		err1 := xls.SetCellValue(sheetName, frameNumberDayN[num], strconv.Itoa(num-2))
		if err1 != nil {
			errs = append(errs, err1)
		}
		err2 := xls.SetCellValue(sheetName, frameNumberDayA[num], strconv.Itoa(num-2))
		if err2 != nil {
			errs = append(errs, err2)
		}
	}
	if crErr := errors.Join(errs...); crErr != nil {
		return fmt.Errorf(textErr, crErr)
	}

	for month, days := range calendarTable {
		cellMonth := frameMonth[month]
		cellNameIndex := 2
		var keyDayStr string
		for _, day := range days {
			cellNameIndex += cellStep
			cellName := cellMonth + strconv.Itoa(cellNameIndex)
			targetDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			dayNow := targetDate.YearDay()

			var cellStyle int
			if keyDay, ok := allKathisma[dayNow]; !ok {
				keyDayStr = ""
				cellStyle = pinkStyle
			} else {
				keyDayStr = strconv.Itoa(keyDay)
				cellStyle = style
			}

			err1 := xls.SetCellValue(sheetName, cellName, keyDayStr)
			if err1 != nil {
				errs = append(errs, err1)
			}
			err2 := xls.SetCellStyle(sheetName, cellName, cellName, cellStyle)
			if err2 != nil {
				errs = append(errs, err2)
			}
		}
		if crErr := errors.Join(errs...); crErr != nil {
			return fmt.Errorf(textErr, crErr)
		}
	}
	return nil
}

func CreateXlSCalendar(startDate time.Time, startKathisma, year int) (*bytes.Buffer, error) {
	if year == 0 {
		year = startDate.Year()
	}
	calendarTable := services.GetCalendarYear(startDate, year)
	calendarKathismas := services.CreateCalendar(startDate, startKathisma, year)

	xls := excelize.NewFile()
	defer func() {
		if err := xls.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	for pair := calendarKathismas.Oldest(); pair != nil; pair = pair.Next() {
		sheetName := fmt.Sprintf("Чтец %d", pair.Key)

		if _, err := xls.NewSheet(sheetName); err != nil {
			return nil, fmt.Errorf("failed create sheet %v", err)
		}
		err1 := addKathismaNumbersToXLS(xls, pair.Key, sheetName)
		if err1 != nil {
			return nil, fmt.Errorf("failed add kafismas number %v", err1)
		}
		err2 := addHeaderOfMonthToWs(xls, sheetName)
		if err2 != nil {
			return nil, fmt.Errorf("failed create header of months %v", err2)
		}
		err3 := addColumnWithNumberDayToWs(xls, sheetName)
		if err3 != nil {
			return nil, fmt.Errorf("failed add column with number day %v", err3)
		}
		err4 := CreateCalendarForReaderToXLS(xls, calendarTable, pair.Value, year, sheetName)
		if err4 != nil {
			return nil, fmt.Errorf("failed create calendar %v", err4)
		}
		if startKathisma > 19 {
			startKathisma = 0
		}
		startKathisma += 1
	}
	p := getPathForFile()
	err := xls.SaveAs(p.outFile)
	if err != nil {
		return nil, fmt.Errorf("failed save %v", err)
	}
	result, err := xls.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed write to buffer %v", err)
	}
	return result, nil
}

type CalendarGeneratorImpl struct{}

func NewCalendarGenerator() *CalendarGeneratorImpl {
	return &CalendarGeneratorImpl{}
}

func (g *CalendarGeneratorImpl) GenerateForGroup(
	year, startOffset int,
) (*bytes.Buffer, domain.CalendarMap, error) {
	if year == 0 {
		year = time.Now().Year()
	}

	startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	calendarTable := services.GetCalendarYear(startDate, year)
	calendarKathismas := services.CreateCalendarForGroup(startOffset, year)

	calendarData := convertToCalendarMap(calendarKathismas)

	xls := excelize.NewFile()
	defer func() {
		if err := xls.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for pair := calendarKathismas.Oldest(); pair != nil; pair = pair.Next() {
		sheetName := fmt.Sprintf("Чтец %d", pair.Key)

		if _, err := xls.NewSheet(sheetName); err != nil {
			return nil, nil, fmt.Errorf("failed create sheet %v", err)
		}
		err1 := addKathismaNumbersToXLS(xls, pair.Key, sheetName)
		if err1 != nil {
			return nil, nil, fmt.Errorf("failed add kafismas number %v", err1)
		}
		err2 := addHeaderOfMonthToWs(xls, sheetName)
		if err2 != nil {
			return nil, nil, fmt.Errorf("failed create header of months %v", err2)
		}
		err3 := addColumnWithNumberDayToWs(xls, sheetName)
		if err3 != nil {
			return nil, nil, fmt.Errorf("failed add column with number day %v", err3)
		}
		err4 := CreateCalendarForReaderToXLS(xls, calendarTable, pair.Value, year, sheetName)
		if err4 != nil {
			return nil, nil, fmt.Errorf("failed create calendar %v", err4)
		}
	}

	p := getPathForFile()
	err := xls.SaveAs(p.outFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed save %v", err)
	}

	result, err := xls.WriteToBuffer()
	if err != nil {
		return nil, nil, fmt.Errorf("failed write to buffer %v", err)
	}

	return result, calendarData, nil
}

func convertToCalendarMap(orderedMap *orderedmap.OrderedMap[int, map[int]int]) domain.CalendarMap {
	calendarMap := make(domain.CalendarMap)
	for pair := orderedMap.Oldest(); pair != nil; pair = pair.Next() {
		calendarMap[pair.Key] = pair.Value
	}
	return calendarMap
}
