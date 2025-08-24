package unm

import (
	"context"
	"pon_watcher/internal/dto"
	"pon_watcher/internal/usecase"
	"pon_watcher/pkg/unm"
	"strings"
)

type ONUManagerAdapter struct {
	server *unm.UNMServer
}

func NewONUManagerAdapter(server *unm.UNMServer) usecase.OpticalNetworkUnitManager {
	return &ONUManagerAdapter{
		server: server,
	}
}

func (a *ONUManagerAdapter) FetchAllONUInformation(ctx context.Context, params usecase.OnuInfoParam) (*dto.OpticalNetworkUnitInfo, error) {
	unmONUInfo, err := a.server.FetchAllOpticalNetworkUnitInformation(ctx, params.PONSlot, params.PONNumber, params.OLT, params.PhysicalAddr)
	if err != nil {
		return nil, err
	}

	onuInfo := &dto.OpticalNetworkUnitInfo{
		DeviceName:        params.DeviceName,
		OnuID:             unmONUInfo.OnuID,
		RxPower:           unmONUInfo.RxPower,
		RxPowerStatus:     unmONUInfo.RxPowerStatus,
		TxPower:           unmONUInfo.TxPower,
		TxPowerStatus:     unmONUInfo.TxPowerStatus,
		CurrTxBias:        unmONUInfo.CurrTxBias,
		CurrTxBiasStatus:  unmONUInfo.CurrTxBiasStatus,
		Temperature:       unmONUInfo.Temperature,
		TemperatureStatus: unmONUInfo.TemperatureStatus,
		Voltage:           unmONUInfo.Voltage,
		VoltageStatus:     unmONUInfo.VoltageStatus,
		PTxPower:          unmONUInfo.PTxPower,
		PRxPower:          unmONUInfo.PRxPower,
	}

	parts := strings.Split(params.DeviceName, "|")
	if len(parts) != 2 {
		return onuInfo, nil
	}

	onuInfo.SplitterName = strings.TrimSpace(parts[0])

	subParts := strings.Split(parts[1], "-")
	if len(subParts) != 2 {
		return onuInfo, nil
	}

	onuInfo.SplitterPort = strings.TrimSpace(subParts[0])
	onuInfo.ClientName = strings.TrimSpace(subParts[1])

	return onuInfo, nil
}

func (a *ONUManagerAdapter) FetchAllONU(ctx context.Context, params usecase.OnuParam) ([]*dto.OpticalNetworkUnit, error) {
	rawOnus, err := a.server.FindAllOpticalNetworkUnits(ctx, params.OLT, params.PONSlot, params.PONNumber, params.Filter)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.OpticalNetworkUnit, 0, len(rawOnus))
	for _, onu := range rawOnus {
		dtos = append(dtos, &dto.OpticalNetworkUnit{
			OltID:    onu.OltID,
			PonID:    onu.PonID,
			OnuNo:    onu.OnuNo,
			Name:     onu.Name,
			Desc:     onu.Desc,
			OnuType:  onu.OnuType,
			IP:       onu.IP,
			AuthType: onu.AuthType,
			Mac:      onu.Mac,
			LoID:     onu.LoID,
			Pwd:      onu.Pwd,
			SwVer:    onu.SwVer,
		})
	}

	return dtos, nil
}
