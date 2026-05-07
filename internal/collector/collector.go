package collector

import (
	"fmt"
	"log"
	"sync"

	"github.com/KatowProject/nvrhikvision-exporter/internal/client"
	"github.com/KatowProject/nvrhikvision-exporter/internal/models"
	"github.com/prometheus/client_golang/prometheus"
)

const maxHDDSlots = 8

var (
	hddHealthDesc = prometheus.NewDesc(
		"hikvision_hdd_health_percent",
		"Health percentage of HDD based on SMART data (Sentinel Algorithm)",
		[]string{"nvr_ip", "disk_id"}, nil,
	)
	hddTemperatureDesc = prometheus.NewDesc(
		"hikvision_hdd_temperature_celsius",
		"HDD Temperature in Celsius (decoded from SMART ID 194/190 packed format)",
		[]string{"nvr_ip", "disk_id"}, nil,
	)
	hddPowerOnDaysDesc = prometheus.NewDesc(
		"hikvision_hdd_power_on_days",
		"HDD Power On Days",
		[]string{"nvr_ip", "disk_id"}, nil,
	)
	hddStatusDesc = prometheus.NewDesc(
		"hikvision_hdd_status",
		"HDD Status (1=ok/functional, 0=not ok)",
		[]string{"nvr_ip", "disk_id", "self_status", "all_status"}, nil,
	)
	hddSmartAttributeDesc = prometheus.NewDesc(
		"hikvision_hdd_smart_attribute",
		"HDD SMART Attribute Raw Values",
		[]string{"nvr_ip", "disk_id", "attribute_id", "attribute_name", "status"}, nil,
	)
	hddSmartNormalizedDesc = prometheus.NewDesc(
		"hikvision_hdd_smart_normalized",
		"HDD SMART Normalized Values (0-100)",
		[]string{"nvr_ip", "disk_id", "attribute_id", "attribute_name"}, nil,
	)
	cameraStatusDesc = prometheus.NewDesc(
		"hikvision_camera_status",
		"Camera online status (1=Online, 0=Offline)",
		[]string{"nvr_ip", "channel_id", "camera_name", "camera_ip"}, nil,
	)
	cameraInfoDesc = prometheus.NewDesc(
		"hikvision_camera_info",
		"Camera information (always 1)",
		[]string{"nvr_ip", "channel_id", "camera_name", "camera_ip", "model", "serial_number", "firmware_version"}, nil,
	)
	cpuUsageDesc = prometheus.NewDesc(
		"hikvision_cpu_usage_percent",
		"NVR CPU Utilization percentage",
		[]string{"nvr_ip"}, nil,
	)
	memUsageDesc = prometheus.NewDesc(
		"hikvision_memory_usage_percent",
		"NVR Memory Usage percentage",
		[]string{"nvr_ip"}, nil,
	)
	uptimeDesc = prometheus.NewDesc(
		"hikvision_uptime_seconds",
		"NVR Uptime in seconds",
		[]string{"nvr_ip"}, nil,
	)
)

// smartAttrNames maps attribute IDs to human-readable names.
// It is used as a fallback when the XML does not include attributeName.
// All attributes are based on actual XML responses from Hikvision devices.
var smartAttrNames = map[int]string{
	1:   "read_error_rate",
	3:   "spin_up_time",
	4:   "start_stop_count",
	5:   "reallocated_sectors",
	7:   "seek_error_rate",
	9:   "power_on_hours",
	10:  "spin_retry_count",
	12:  "power_cycle_count",
	18:  "head_flying_hours",
	187: "reported_uncorrectable_errors",
	188: "command_timeout",
	190: "airflow_temperature",
	192: "power_off_retract_count",
	193: "load_cycle_count",
	194: "temperature",
	195: "hardware_ecc_recovered",
	197: "current_pending_sectors",
	198: "offline_uncorrectable",
	199: "udma_crc_error",
	240: "head_flying_hours_alt",
	241: "total_lbas_written",
	242: "total_lbas_read",
}

type NVRTarget struct {
	Client *client.Client
	IP     string
	Name   string
}

type HikvisionCollector struct {
	Targets []*NVRTarget
}

func NewCollector(targets []*NVRTarget) *HikvisionCollector {
	return &HikvisionCollector{
		Targets: targets,
	}
}

func (c *HikvisionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- hddHealthDesc
	ch <- hddTemperatureDesc
	ch <- hddPowerOnDaysDesc
	ch <- hddStatusDesc
	ch <- hddSmartAttributeDesc
	ch <- hddSmartNormalizedDesc
	ch <- cameraStatusDesc
	ch <- cameraInfoDesc
	ch <- cpuUsageDesc
	ch <- memUsageDesc
	ch <- uptimeDesc
}

func (c *HikvisionCollector) Collect(ch chan<- prometheus.Metric) {
	for _, target := range c.Targets {
		c.collectFromNVR(target, ch)
	}
}

func (c *HikvisionCollector) collectFromNVR(target *NVRTarget, ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	// 1. COLLECT CAMERA STATUS
	wg.Add(1)
	go func() {
		defer wg.Done()
		var camList models.InputProxyChannelList
		if err := target.Client.FetchXML("/ISAPI/ContentMgmt/InputProxy/channels", &camList); err != nil {
			log.Printf("[%s] Error fetching camera list: %v", target.IP, err)
			return
		}

		log.Printf("[%s] Found %d camera channels", target.IP, len(camList.Channels))

		for _, cam := range camList.Channels {
			statusVal := 0.0
			if cam.SourceInputPortDescriptor.IPAddress != "" {
				statusVal = 1.0
			}

			ch <- prometheus.MustNewConstMetric(
				cameraStatusDesc,
				prometheus.GaugeValue,
				statusVal,
				target.IP,
				cam.ID,
				cam.Name,
				cam.SourceInputPortDescriptor.IPAddress,
			)
			ch <- prometheus.MustNewConstMetric(
				cameraInfoDesc,
				prometheus.GaugeValue,
				1,
				target.IP,
				cam.ID,
				cam.Name,
				cam.SourceInputPortDescriptor.IPAddress,
				cam.SourceInputPortDescriptor.Model,
				cam.SourceInputPortDescriptor.SerialNumber,
				cam.SourceInputPortDescriptor.FirmwareVersion,
			)
		}
	}()

	// 2. COLLECT SYSTEM STATUS
	wg.Add(1)
	go func() {
		defer wg.Done()
		var sys models.DeviceStatus
		if err := target.Client.FetchXML("/ISAPI/System/status", &sys); err != nil {
			log.Printf("[%s] Error fetching system status: %v", target.IP, err)
			return
		}

		cpuUtil := 0.0
		if len(sys.CPUList.CPUs) > 0 {
			cpuUtil = sys.CPUList.CPUs[0].CPUUtilization
		}

		memUsage := 0.0
		memAvailable := 0.0
		if len(sys.MemoryList.Memories) > 0 {
			memUsage = sys.MemoryList.Memories[0].MemoryUsage
			memAvailable = sys.MemoryList.Memories[0].MemoryAvailable
		}

		totalMemory := memUsage + memAvailable
		memPercent := CalculateMemoryPercent(memUsage, totalMemory)

		log.Printf("[%s] CPU: %.2f%%, Memory: %.2f/%.2f MB (%.2f%%), Uptime: %ds",
			target.IP, cpuUtil, memUsage, totalMemory, memPercent, sys.DeviceUpTime)

		ch <- prometheus.MustNewConstMetric(cpuUsageDesc, prometheus.GaugeValue, cpuUtil, target.IP)
		ch <- prometheus.MustNewConstMetric(uptimeDesc, prometheus.CounterValue, float64(sys.DeviceUpTime), target.IP)
		ch <- prometheus.MustNewConstMetric(memUsageDesc, prometheus.GaugeValue, memPercent, target.IP)
	}()

	// 3. COLLECT HDD SMART DATA
	for i := 1; i <= maxHDDSlots; i++ {
		wg.Add(1)
		go func(diskID int) {
			defer wg.Done()

			var smart models.SmartTestStatus
			path := fmt.Sprintf("/ISAPI/ContentMgmt/Storage/hdd/%d/SMARTTest/status", diskID)
			if err := target.Client.FetchXML(path, &smart); err != nil {
				return
			}

			diskIDStr := fmt.Sprintf("%d", diskID)

			// Build rawAttrs for decodeTemperature and temperature export.
			rawAttrs := make(map[int]int64)
			for _, res := range smart.TestResultList {
				rawAttrs[res.AttributeID] = res.RawValue
			}

			// HDD Health Score
			health := CalculateHDDHealth(smart)
			ch <- prometheus.MustNewConstMetric(
				hddHealthDesc,
				prometheus.GaugeValue,
				health,
				target.IP, diskIDStr,
			)

			// HDD temperature uses the decoded value, not the raw struct field,
			// because IDs 194 and 190 use a vendor-specific packed format.
			decodedTemp := decodeTemperature(smart, rawAttrs)
			ch <- prometheus.MustNewConstMetric(
				hddTemperatureDesc,
				prometheus.GaugeValue,
				float64(decodedTemp),
				target.IP, diskIDStr,
			)

			// HDD Power On Days
			ch <- prometheus.MustNewConstMetric(
				hddPowerOnDaysDesc,
				prometheus.GaugeValue,
				float64(smart.PowerOnDay),
				target.IP, diskIDStr,
			)

			// HDD Status
			statusVal := 0.0
			if smart.SelfEvaluaingStatus == "ok" && smart.AllEvaluaingStatus == "functional" {
				statusVal = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				hddStatusDesc,
				prometheus.GaugeValue,
				statusVal,
				target.IP, diskIDStr, smart.SelfEvaluaingStatus, smart.AllEvaluaingStatus,
			)

			// Export all SMART attributes reported by the device.
			// The map is only used as a fallback for attribute names.
			// Guard against duplicates so the same attribute ID is not exported twice
			// (this can happen when XML parsing fails and AttributeID becomes 0).
			exportedAttrs := make(map[int]struct{})
			for _, result := range smart.TestResultList {
				// Skip attribute ID 0; this indicates a parsing failure or malformed XML.
				if result.AttributeID == 0 {
					continue
				}
				// Skip duplicate attribute IDs within the same disk.
				if _, seen := exportedAttrs[result.AttributeID]; seen {
					continue
				}
				exportedAttrs[result.AttributeID] = struct{}{}

				// Resolve the attribute name using XML, then the map, then "attr_<id>".
				attrName := result.AttributeName
				if attrName == "" {
					if mapped, ok := smartAttrNames[result.AttributeID]; ok {
						attrName = mapped
					} else {
						attrName = fmt.Sprintf("attr_%d", result.AttributeID)
					}
				}

				ch <- prometheus.MustNewConstMetric(
					hddSmartAttributeDesc,
					prometheus.GaugeValue,
					float64(result.RawValue),
					target.IP, diskIDStr, fmt.Sprintf("%d", result.AttributeID), attrName, result.Status,
				)

				ch <- prometheus.MustNewConstMetric(
					hddSmartNormalizedDesc,
					prometheus.GaugeValue,
					float64(result.Value),
					target.IP, diskIDStr, fmt.Sprintf("%d", result.AttributeID), attrName,
				)
			}

			log.Printf("[%s] HDD Slot %d: Health=%.0f%%, Temp=%d°C, Status=%s",
				target.IP, diskID, health, decodedTemp, smart.SelfEvaluaingStatus)
		}(i)
	}

	wg.Wait()
}
