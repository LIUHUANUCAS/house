package main

type DailyHouse struct {
	// MonthData MonthData `json:"month_data"`
	Day       string    `json:"day"`
	DailyData DailyData `json:"daily_data"`
}
type MonthData struct {
	TotalCount float64 `json:"total_count"`
	TotalArea  float64 `json:"total_area"`
	HouseCount float64 `json:"house_count"`
	HouseArea  float64 `json:"house_area"`
}
type DailyData struct {
	TotalCount float64 `json:"total_count"`
	TotalArea  float64 `json:"total_area"`
	HouseCount float64 `json:"house_count"`
	HouseArea  float64 `json:"house_area"`
}

func getDefaultDailyHouse() DailyHouse {
	return DailyHouse{
		Day: "2025-04-08",
		DailyData: DailyData{
			TotalCount: 744,
			TotalArea:  64840,
			HouseCount: 619,
			HouseArea:  58754.18,
		},
	}
}
