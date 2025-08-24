package dto

type OpticalNetworkUnit struct {
	OltID    string `json:"olt_id,omitempty"`
	PonID    string `json:"pon_id,omitempty"`
	OnuNo    string `json:"onu_no,omitempty"`
	Name     string `json:"name,omitempty"`
	Desc     string `json:"desc,omitempty"`
	OnuType  string `json:"onu_type,omitempty"`
	IP       string `json:"ip,omitempty"`
	AuthType string `json:"auth_type,omitempty"`
	Mac      string `json:"mac,omitempty"`
	LoID     string `json:"lo_id,omitempty"`
	Pwd      string `json:"pwd,omitempty"`
	SwVer    string `json:"sw_ver,omitempty"`
	HwVer    string `json:"hw_ver,omitempty"`
}

type OpticalNetworkUnitInfo struct {
	SplitterName      string `json:"splitter_name,omitempty"`
	SplitterPort      string `json:"splitter_port,omitempty"`
	ClientName        string `json:"client_name,omitempty"`
	DeviceName        string `json:"device_name,omitempty"`
	OnuID             string `json:"onu_id,omitempty"`
	RxPower           string `json:"rx_power,omitempty"`
	RxPowerStatus     string `json:"rx_power_status,omitempty"`
	TxPower           string `json:"tx_power,omitempty"`
	TxPowerStatus     string `json:"tx_power_status,omitempty"`
	CurrTxBias        string `json:"curr_tx_bias,omitempty"`
	CurrTxBiasStatus  string `json:"curr_tx_bias_status,omitempty"`
	Temperature       string `json:"temperature,omitempty"`
	TemperatureStatus string `json:"temperature_status,omitempty"`
	Voltage           string `json:"voltage,omitempty"`
	VoltageStatus     string `json:"voltage_status,omitempty"`
	PTxPower          string `json:"p_tx_power,omitempty"`
	PRxPower          string `json:"p_rx_power,omitempty"`
	PhysicalAddress   string `json:"-"`
}

type OpticalNetworkUnitState struct {
	Onus  OpticalNetworkUnitInfo `json:"onus,omitempty"`
	Total uint                   `json:"total,omitempty"`
}

type OpticalNetworkUnitResponse struct {
	Online  OpticalNetworkUnitState `json:"online,omitempty"`
	Offline OpticalNetworkUnitState `json:"offline,omitempty"`
	Total   uint                    `json:"total,omitempty"`
}
