package main

// DailyHouse  daily house req data
type DailyHouse struct {
	MonthData MonthData `json:"month_data"`
	Month     string    `json:"month"`
	Day       string    `json:"day"`
	DailyData DailyData `json:"daily_data"`
}

// DailyHouseResp  daily house resp data
type DailyHouseResp struct {
	Day       string    `json:"day"`
	DailyData DailyData `json:"daily_data"`
}

// MonthHouseResp HouseResp  month house resp data
type MonthHouseResp struct {
	MonthData MonthData `json:"month_data"`
	Month     string    `json:"month"`
}

// MonthData month house data
type MonthData struct {
	TotalCount float64 `json:"total_count"`
	TotalArea  float64 `json:"total_area"`
	HouseCount float64 `json:"house_count"`
	HouseArea  float64 `json:"house_area"`
}

// DailyData daily house data
type DailyData struct {
	TotalCount float64 `json:"total_count"`
	TotalArea  float64 `json:"total_area"`
	HouseCount float64 `json:"house_count"`
	HouseArea  float64 `json:"house_area"`
	HousePrice float64 `json:"house_price"`
	TotalPrice float64 `json:"total_price"`
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
