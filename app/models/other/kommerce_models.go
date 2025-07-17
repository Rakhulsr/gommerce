package other

type KomerceMeta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

type KomerceSearchDestinationResponse struct {
	Meta KomerceMeta                  `json:"meta"`
	Data []KomerceDomesticDestination `json:"data"`
}

type KomerceSingleDestinationResponse struct {
	Meta KomerceMeta                `json:"meta"`
	Data KomerceDomesticDestination `json:"data"`
}

type KomerceDomesticDestination struct {
	ID              int    `json:"id"`
	Label           string `json:"label"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	ZipCode         string `json:"zip_code"`
}

type DropdownItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type KomerceShippingCostResponse struct {
	Meta KomerceMeta            `json:"meta"`
	Data []KomerceCourierResult `json:"data"`
}

type KomerceCostData struct {
	OriginDetails      KomerceOriginDestination `json:"origin_details"`
	DestinationDetails KomerceOriginDestination `json:"destination_details"`
	Results            []KomerceCourierResult   `json:"results"`
}

type KomerceOriginDestination struct {
	SubdistrictID string `json:"subdistrict_id"`
	ProvinceID    string `json:"province_id"`
	Province      string `json:"province"`
	City          string `json:"city"`
	Type          string `json:"type"`
	Subdistrict   string `json:"subdistrict"`
}

type KomerceCourierResult struct {
	Code  string              `json:"code"`
	Name  string              `json:"name"`
	Costs []KomerceCostResult `json:"costs"`
}

type KomerceCostResult struct {
	Service     string `json:"service"`
	Description string `json:"description"`
	Cost        []struct {
		Value int    `json:"value"`
		Etd   string `json:"etd"`
		Note  string `json:"note"`
	} `json:"cost"`
}
