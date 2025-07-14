package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	carbon.SetTimezone(time.Now().Location().String())
	dates := make([]LunarDate, 0)

	for month := 1; month <= 12; month++ {
		first := carbon.CreateFromLunar(year, month, 1, false)
		fifteenth := carbon.CreateFromLunar(year, month, 15, false)

		yearString := fmt.Sprintf("%s年", GetGanZhiYear(first.Lunar().Year()))
		monthString := lunarMonthReplacer.Replace(first.Lunar().ToMonthString())

		dates = append(dates, LunarDate{
			Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, first.Lunar().ToDayString()),
			Date:  first.ToDateString(),
		}, LunarDate{
			Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, fifteenth.Lunar().ToDayString()),
			Date:  fifteenth.ToDateString(),
		})

		if first.Lunar().LeapMonth() == month {
			leapFirst := carbon.CreateFromLunar(year, month, 1, true)
			leapFifteenth := carbon.CreateFromLunar(year, month, 15, true)

			dates = append(dates, LunarDate{
				Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, leapFirst.Lunar().ToDayString()),
				Date:  leapFirst.ToDateString(),
			}, LunarDate{
				Lunar: fmt.Sprintf("%s%s%s", yearString, monthString, leapFifteenth.Lunar().ToDayString()),
				Date:  leapFifteenth.ToDateString(),
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
			fmt.Println("Error:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(dates); err != nil {
			fmt.Println("Error:", err)
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	})
	workers.Serve(nil) // use http.DefaultServeMux
}
