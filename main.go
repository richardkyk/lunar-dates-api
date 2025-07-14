package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/dromara/carbon/v2"
	"github.com/syumai/workers"
)

var heavenlyStems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var earthlyBranches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

var lunarMonthReplacer = strings.NewReplacer(
	"腊", "十二",
	"冬", "十一",
	"闰", "閏",
)

func GetGanZhiYear(year int) string {
	base := 1984
	offset := (year - base) % 60
	if offset < 0 {
		offset += 60
	}
	heavenly := heavenlyStems[offset%10]
	earthly := earthlyBranches[offset%12]
	return heavenly + earthly
}

type LunarDate struct {
	Date  string `json:"date"`
	Lunar string `json:"lunar"`
}

func GetLunarDates(year int) ([]LunarDate, error) {
	dates := make([]LunarDate, 0)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Println("Error:", err)
		return nil, err
	}

	for month := 1; month <= 12; month++ {
		// Creates a Carbon instance from the specified lunar date. The date is represented in UTC
		first := carbon.CreateFromLunar(year, month, 1, false)
		// We need to convert the UTC date to the local time zone of Shanghai and strip the timestamp
		firstDateParsed, err := time.Parse(time.RFC3339, first.ToIso8601String())
		if err != nil {
			log.Println("Error:", err)
			return nil, err
		}
		firstDateString := firstDateParsed.In(loc).Format("2006-01-02")

		fifteenth := carbon.CreateFromLunar(year, month, 15, false)
		fifteenthDateParsed, err := time.Parse(time.RFC3339, fifteenth.ToIso8601String())
		if err != nil {
			log.Println("Error:", err)
			return nil, err
		}
		fifteenthDateString := fifteenthDateParsed.In(loc).Format("2006-01-02")

		// We create a new Carbon instance using the local date so that we get the correct lunar date
		firstLunar := carbon.Parse(firstDateString).Lunar()
		fifteenthLunar := carbon.Parse(fifteenthDateString).Lunar()

		yearString := fmt.Sprintf("%s年", GetGanZhiYear(firstLunar.Year()))
		monthString := lunarMonthReplacer.Replace(firstLunar.ToMonthString())

		dates = append(dates, LunarDate{
			Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, firstLunar.ToDayString()),
			Date:  firstDateString,
		}, LunarDate{
			Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, fifteenthLunar.ToDayString()),
			Date:  fifteenthDateString,
		})

		if firstLunar.LeapMonth() == month {
			leapFirst := carbon.CreateFromLunar(year, month, 1, true)
			leapFirstDateParsed, err := time.Parse(time.RFC3339, leapFirst.ToIso8601String())
			if err != nil {
				log.Println("Error:", err)
				return nil, err
			}
			leapFirstDateString := leapFirstDateParsed.In(loc).Format("2006-01-02")

			leapFifteenth := carbon.CreateFromLunar(year, month, 15, true)
			leapFifteenthDateParsed, err := time.Parse(time.RFC3339, leapFifteenth.ToIso8601String())
			if err != nil {
				log.Println("Error:", err)
				return nil, err
			}
			leapFifteenthDateString := leapFifteenthDateParsed.In(loc).Format("2006-01-02")

			leapFirstLunar := carbon.Parse(leapFirstDateString).Lunar()
			leapFifteenthLunar := carbon.Parse(leapFifteenthDateString).Lunar()

			yearString := fmt.Sprintf("%s年", GetGanZhiYear(leapFirstLunar.Year()))
			monthString := lunarMonthReplacer.Replace(leapFirstLunar.ToMonthString())

			dates = append(dates, LunarDate{
				Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, leapFirstLunar.ToDayString()),
				Date:  leapFirstDateString,
			}, LunarDate{
				Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, leapFifteenthLunar.ToDayString()),
				Date:  leapFifteenthDateString,
			})
		}
	}
	return dates, nil

}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		yearParam := req.URL.Query().Get("year")
		year := time.Now().Year()
		if yearParam != "" {
			if parsed, err := strconv.Atoi(yearParam); err == nil {
				year = parsed
			}
		}

		dates, err := GetLunarDates(year)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(dates); err != nil {
			log.Println("Error:", err)
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
