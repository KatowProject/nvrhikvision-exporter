package models

import "encoding/xml"

// --- SYSTEM STATUS (CPU/MEM) ---
type DeviceStatus struct {
	XMLName           xml.Name   `xml:"DeviceStatus"`
	CurrentDeviceTime string     `xml:"currentDeviceTime"`
	DeviceUpTime      int64      `xml:"deviceUpTime"`
	CPUList           CPUList    `xml:"CPUList"`
	MemoryList        MemoryList `xml:"MemoryList"`
}

type CPUList struct {
	CPUs []CPU `xml:"CPU"`
}

type CPU struct {
	CPUDescription string  `xml:"cpuDescription"`
	CPUUtilization float64 `xml:"cpuUtilization"`
}

type MemoryList struct {
	Memories []Memory `xml:"Memory"`
}

type Memory struct {
	MemoryDescription string  `xml:"memoryDescription"`
	MemoryUsage       float64 `xml:"memoryUsage"`
	MemoryAvailable   float64 `xml:"memoryAvailable"`
}

// --- CAMERA CHANNELS ---
type InputProxyChannelList struct {
	XMLName  xml.Name            `xml:"InputProxyChannelList"`
	Version  string              `xml:"version,attr"`
	Size     int                 `xml:"size,attr"`
	Channels []InputProxyChannel `xml:"InputProxyChannel"`
}

type InputProxyChannel struct {
	ID                           string                    `xml:"id"`
	Name                         string                    `xml:"name"`
	SourceInputPortDescriptor    SourceInputPortDescriptor `xml:"sourceInputPortDescriptor"`
	CertificateValidationEnabled bool                      `xml:"certificateValidationEnabled"`
	DefaultAdminPortEnabled      bool                      `xml:"defaultAdminPortEnabled"`
	EnableAnr                    bool                      `xml:"enableAnr"`
	EnableTiming                 bool                      `xml:"enableTiming"`
	DevIndex                     string                    `xml:"devIndex"`
}

type SourceInputPortDescriptor struct {
	ProxyProtocol        string `xml:"proxyProtocol"`
	AddressingFormatType string `xml:"addressingFormatType"`
	IPAddress            string `xml:"ipAddress"`
	ManagePortNo         int    `xml:"managePortNo"`
	SrcInputPort         int    `xml:"srcInputPort"`
	UserName             string `xml:"userName"`
	StreamType           string `xml:"streamType"`
	Model                string `xml:"model"`
	SerialNumber         string `xml:"serialNumber"`
	FirmwareVersion      string `xml:"firmwareVersion"`
	DeviceID             string `xml:"deviceID"`
}

// --- HDD SMART DATA ---
type SmartTestStatus struct {
	XMLName             xml.Name     `xml:"SmartTestStatus"`
	ID                  string       `xml:"id"`
	Temperature         int          `xml:"temprature"`
	PowerOnDay          int          `xml:"powerOnDay"`
	SelfEvaluaingStatus string       `xml:"selfEvaluaingStatus"`
	AllEvaluaingStatus  string       `xml:"allEvaluaingStatus"`
	SelfTestPercent     int          `xml:"selfTestPercent"`
	SelfTestStatus      string       `xml:"selfTestStatus"`
	TestType            string       `xml:"testType"`
	TestResultList      []TestResult `xml:"TestResultList>TestResult"`
}

// TestResult represents a single SMART attribute from a disk.
// Vendor-specific notes:
//   - Seagate: IDs 1 and 7 use packed raw values; use Value (normalized) for scoring.
//   - Seagate/WD: IDs 190 and 194 use packed raw values; byte 0 is the current temperature in Celsius.
type TestResult struct {
	AttributeID   int    `xml:"attributeID"`
	AttributeName string `xml:"attributeName"`
	Status        string `xml:"status"`
	Flags         int    `xml:"flags"`
	Thresholds    int    `xml:"thresholds"`
	Value         int    `xml:"value"` // Normalized value (0-100), more reliable for scoring.
	Worst         int    `xml:"worst"`
	RawValue      int64  `xml:"rawValue"`
}
