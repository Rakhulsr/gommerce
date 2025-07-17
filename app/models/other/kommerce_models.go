package other

type Meta struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type KomerceCostDetail struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
	Etd         string `json:"etd"`
}

type KomerceShippingCostResponse struct {
	Meta Meta                `json:"meta"`
	Data []KomerceCostDetail `json:"data"`
}

type KomerceDomesticDestination struct {
	ID              int    `json:"id"`
	Label           string `json:"label"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	Type            string `json:"type"`
	ZipCode         string `json:"zip_code"`
	IsCapital       int    `json:"is_capital"`
}

type KomerceDomesticDestinationResponse struct {
	Meta Meta                         `json:"meta"`
	Data []KomerceDomesticDestination `json:"data"`
}
