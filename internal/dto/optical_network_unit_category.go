package dto

type OpticalNetworkUnitCategory struct {
	ONUs  []*OpticalNetworkUnitInfo `json:"onus"`
	Total int                       `json:"total"`
}

type CategorizedOpticalNetworkUnits struct {
	Online  OpticalNetworkUnitCategory `json:"online"`
	Offline OpticalNetworkUnitCategory `json:"offline"`
	Total   int                        `json:"total"`
}
