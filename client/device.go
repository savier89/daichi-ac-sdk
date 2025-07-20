package client

// MQTTUser — структура для MQTT-данных
type MQTTUser struct {
	Username string `json:"username"` // ✅ Экспортируемое поле
	Password string `json:"password"` // ✅ Экспортируемое поле
}

// DaichiUser — структура данных пользователя
type DaichiUser struct {
	ID                       int           `json:"id"`
	Token                    string        `json:"token"`
	Email                    string        `json:"email"`
	MQTTUser                 *MQTTUser     `json:"mqttUser"` // ✅ Указатель на структуру
	IsEmailConfirmed         bool          `json:"isEmailConfirmed"`
	Phone                    *string       `json:"phone,omitempty"`
	IsPhoneConfirmed         bool          `json:"isPhoneConfirmed"`
	FIO                      string        `json:"fio"`
	Company                  string        `json:"company"`
	UserType                 string        `json:"userType"`
	AccessRequests           []interface{} `json:"accessRequests"`
	ExpiredIn                *string       `json:"expiredIn,omitempty"`
	DeleteAccountRequestedAt *string       `json:"deleteAccountRequestedAt,omitempty"`
	Image                    *string       `json:"image,omitempty"`
}

// DeviceState — улучшенная структура для поля state
type DeviceState struct {
	IsOn    bool                `json:"isOn"`
	Info    DeviceStateInfo     `json:"info"`
	Details []DeviceStateDetail `json:"details"`
}

// DeviceStateInfo — информация о состоянии устройства
type DeviceStateInfo struct {
	Text      string   `json:"text"`
	Icons     []string `json:"icons"`
	IconsSvg  []string `json:"iconsSvg"`
	IconNames []string `json:"iconNames"`
}

// DeviceStateDetail — детали состояния
type DeviceStateDetail struct {
	Details []struct {
		Icon     *string `json:"icon,omitempty"`
		IconSvg  *string `json:"iconSvg,omitempty"`
		IconName string  `json:"iconName,omitempty"`
		Text     *string `json:"text,omitempty"`
	} `json:"details"`
}

// DeviceFeatures — улучшенная структура для features
type DeviceFeatures struct {
	CanChangeWiFiFromServer bool `json:"canChangeWiFiFromServer"` // Поддержка настройки WiFi через сервер
	ServerTimerSupported    bool `json:"serverTimerSupported"`    // Поддержка таймера
	CanControlByBle         bool `json:"canControlByBle"`         // Поддержка BLE
}

// DeviceTheme — структура темы устройства
type DeviceTheme struct {
	Primary    string   `json:"primary"`    // Основной цвет
	Gradient   []string `json:"gradient"`   // Градиент
	Background string   `json:"background"` // Цвет фона
}

// DaichiBuildingDeviceStruct — структура устройства в здании
type DaichiBuildingDeviceStruct struct {
	ID           int            `json:"id"`
	Serial       string         `json:"serial"`
	Status       string         `json:"status"`
	Title        string         `json:"title"`
	CurTemp      float64        `json:"curTemp"`
	State        DeviceState    `json:"state"`
	Features     DeviceFeatures `json:"features"`
	Theme        DeviceTheme    `json:"theme"`
	CurrentState []struct {
		Text string `json:"text"`
	} `json:"currentState"`
	CurrentStateDetailed []struct {
		Text      string   `json:"text"`
		Icon      string   `json:"icon"`
		IconSvg   string   `json:"iconSvg"`
		IconNames []string `json:"iconNames"`
	} `json:"currentStateDetailed"`

	GroupID           *interface{} `json:"groupId,omitempty"`
	BuildingID        int          `json:"buildingId"`
	LastOnline        string       `json:"lastOnline"`
	CreatedAt         string       `json:"createdAt"`
	Pinned            bool         `json:"pinned"`
	Access            string       `json:"access"`
	Progress          *interface{} `json:"progress,omitempty"`
	CurrentPreset     *interface{} `json:"currentPreset,omitempty"`
	Timer             *interface{} `json:"timer,omitempty"`
	CloudType         string       `json:"cloudType"`
	DistributionType  string       `json:"distributionType"`
	Company           string       `json:"company"`
	IsBle             bool         `json:"isBle"`
	DeviceControlType string       `json:"deviceControlType"`
	FirmwareType      string       `json:"firmwareType"`
	VrfTitle          *interface{} `json:"vrfTitle,omitempty"`
	DeviceType        string       `json:"deviceType"`
	Subscription      *interface{} `json:"subscription,omitempty"`
	SubscriptionID    *int         `json:"subscriptionId,omitempty"`
	WarrantyNumber    *string      `json:"warrantyNumber,omitempty"`
	ConditionerSerial *string      `json:"conditionerSerial,omitempty"`
	UpdatedAt         *string      `json:"updatedAt,omitempty"`
	Online            bool         `json:"online,omitempty"`
}

// IsOnline — проверяет, подключен ли кондиционер
func (d *DaichiBuildingDeviceStruct) IsOnline() bool {
	return d.Status == "connected"
}
