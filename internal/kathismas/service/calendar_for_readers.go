package service

import (
	"time"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func getTotalKathismas() [20]int {
	return [20]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
}

func GetCalendarYear(startCalendarDate time.Time, year int) map[int][]int {
	tableYear := make(map[int][]int)
	var currentDayList []int
	currentMonth := 1

	for startCalendarDate.Year() != year+1 {
		if startCalendarDate.Year() == year+1 {
			break
		}
		if startCalendarDate.Day() == 1 {
			if len(currentDayList) > 0 {
				tableYear[currentMonth] = currentDayList
			}
			currentDayList = []int{}
			currentMonth = int(startCalendarDate.Month())
		}
		currentDayList = append(currentDayList, startCalendarDate.Day())
		startCalendarDate = startCalendarDate.AddDate(0, 0, 1)
	}

	if len(currentDayList) > 0 {
		tableYear[currentMonth] = currentDayList
	}

	return tableYear
}

func GetEasterDate(year int) time.Time {
	a := year % 4
	b := year % 7
	c := year % 19
	d := (19*c + 15) % 30
	e := (2*a + 4*b - d + 34) % 7
	month := (d + e + 114) / 31
	day := ((d + e + 114) % 31) + 1

	// Перевод даты из Юлианского в Григорианский календарь
	easter := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	offset := year/100 - year/400 - 2
	easter = easter.AddDate(0, 0, offset)
	return easter
}

func GetBoundaryDays(easterDay time.Time) (startNoReading, endNoReading time.Time) {
	startNoReading = easterDay.AddDate(0, 0, -3)
	endNoReading = easterDay.AddDate(0, 0, 6)
	return startNoReading, endNoReading
}

func GetNumberDaysInYear(year int) int {
	startYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	endYear := startYear.AddDate(1, 0, 0)
	return int(endYear.Sub(startYear).Hours() / 24)
}

func CreateCalendar(startDate time.Time, startKathisma, year int) *orderedmap.OrderedMap[int, map[int]int] {
	if year == 0 {
		year = startDate.Year()
	}
	easterDay := GetEasterDate(year)
	startNoReading, endNoReading := GetBoundaryDays(easterDay)
	numberDaysInYear := GetNumberDaysInYear(year)
	totalKathismas := getTotalKathismas()
	readersMap := orderedmap.New[int, map[int]int]()
	for _, numberKathisma := range totalKathismas {
		allKathismas := GetListDate(startNoReading, endNoReading, numberKathisma, numberDaysInYear, totalKathismas)
		if startKathisma > 19 {
			startKathisma = 0
		}
		readersMap.Set(numberKathisma, allKathismas)
		startKathisma += 1
	}
	return readersMap
}

func CreateCalendarForGroup(startOffset, year int) *orderedmap.OrderedMap[int, map[int]int] {
	if year == 0 {
		year = time.Now().Year()
	}
	easterDay := GetEasterDate(year)
	startNoReading, endNoReading := GetBoundaryDays(easterDay)
	numberDaysInYear := GetNumberDaysInYear(year)
	totalKathismas := getTotalKathismas()
	readersMap := orderedmap.New[int, map[int]int]()

	currentKathisma := startOffset
	for _, readerNumber := range totalKathismas {
		allKathismas := GetListDate(startNoReading, endNoReading, currentKathisma, numberDaysInYear, totalKathismas)
		readersMap.Set(readerNumber, allKathismas)
		currentKathisma++
		if currentKathisma > 20 {
			currentKathisma = 1
		}
	}
	return readersMap
}
