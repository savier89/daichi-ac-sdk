package client

// DeviceFunctionControl — структура управления функцией устройства
type DeviceFunctionControl struct {
	FunctionID int      `json:"functionId"`
	Value      *float64 `json:"value,omitempty"` // Для числовых значений
	IsOn       *bool    `json:"isOn,omitempty"`  // Для boolean
	Parameters *string  `json:"parameters,omitempty"`
}

// DeviceControlRequest — запрос на управление устройством
type DeviceControlRequest struct {
	CmdID               int                   `json:"cmdId"`
	Value               DeviceFunctionControl `json:"value"`
	ConflictResolveData *string               `json:"conflictResolveData,omitempty"`
}
