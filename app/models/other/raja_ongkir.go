package other

type ProvinceResponse struct {
	RajaOngkir struct {
		Query  interface{} `json:"query"`
		Status struct {
			Code        int    `json:"code"`
			Description string `json:"description"`
		} `json:"status"`
		Results []Province `json:"results"`
	} `json:"rajaongkir"`
}
type ProvinceData struct {
	Results []Province `json:"results"`
}

type Province struct {
	ID   string `json:"province_id"`
	Name string `json:"province"`
}

type SingleProvinceResponse struct {
	RajaOngkir struct {
		Query  interface{} `json:"query"`
		Status struct {
			Code        int    `json:"code"`
			Description string `json:"description"`
		} `json:"status"`
		Results Province `json:"results"` // KOREKSI: Ini adalah objek Province tunggal
	} `json:"rajaongkir"`
}
type City struct {
	ID         string `json:"city_id"`
	ProvinceID string `json:"province_id"`
	Province   string `json:"province"`
	Type       string `json:"type"`
	Name       string `json:"city_name"`
	PostalCode string `json:"postal_code"`
}

type CityResponse struct {
	RajaOngkir struct {
		Query  interface{} `json:"query"`
		Status struct {
			Code        int    `json:"code"`
			Description string `json:"description"`
		} `json:"status"`
		Results []City `json:"results"`
	} `json:"rajaongkir"`
}

type SingleCityResponse struct {
	RajaOngkir struct {
		Query  interface{} `json:"query"`
		Status struct {
			Code        int    `json:"code"`
			Description string `json:"description"`
		} `json:"status"`
		Results City `json:"results"` // KOREKSI: Ini adalah objek City tunggal
	} `json:"rajaongkir"`
}

type CostResponse struct {
	RajaOngkir struct {
		Query  interface{} `json:"query"`
		Status struct {
			Code        int    `json:"code"`
			Description string `json:"description"`
		} `json:"status"`
		Results []CourierResult `json:"results"`
	} `json:"rajaongkir"`
}

type CourierResult struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Costs []Cost `json:"costs"`
}

type OngkirResponse struct {
	OngkirData OngkirData `json:"rajaongkir"`
}

type OngkirData struct {
}

type Cost struct {
	Service     string              `json:"service"`
	Description string              `json:"description"`
	Cost        []ServiceCostDetail `json:"cost"`
}

type ServiceCostDetail struct {
	Value int    `json:"value"`
	Etd   string `json:"etd"`
	Note  string `json:"note"`
}

type Courier struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
