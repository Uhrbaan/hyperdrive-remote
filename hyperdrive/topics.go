package hyperdrive

import (
	"fmt"

	"github.com/google/uuid"
)

var (
	callerID = uuid.NewString()
)

const (
	RootTopic         = "Anki"
	HostsTopicBase    = RootTopic + "/Hosts/U"    // Base for all specific Host (U) topics
	VehiclesTopicBase = RootTopic + "/Vehicles/U" // Base for all specific Vehicle (U) topics
)

type Host string

func (h Host) Id() string {
	return string(h)
}

func DefaultHost() Host {
	return Host("hyperdrive")
}

// Anki/Hosts/U/{hostID}/I/{callerID}
func (h Host) IntentTopic() string {
	return fmt.Sprintf("%s/%s/I/%s", HostsTopicBase, h.Id(), callerID)
}

// Anki/Hosts/U/{hostID}/S
func (h Host) StatusTopic() string {
	return fmt.Sprintf("%s/%s/S", HostsTopicBase, h.Id())
}

// Anki/Hosts/U/{hostID}/S/online
func (h Host) StatusOnlineTopic() string {
	return fmt.Sprintf("%s/online", h.StatusTopic())
}

// statusIntendedBase constructs the base path for intended statuses.
func (h Host) statusIntendedBase() string {
	return fmt.Sprintf("%s/intended", h.StatusTopic())
}

// Anki/Hosts/U/{hostID}/S/intended/discover
func (h Host) StatusIntendedDiscoverTopic() string {
	return fmt.Sprintf("%s/discover", h.statusIntendedBase())
}

// Anki/Hosts/U/{hostID}/S/intended/discoverSubscription
func (h Host) StatusIntendedDiscoverSubscriptionTopic() string {
	return fmt.Sprintf("%s/discoverSubscription", h.statusIntendedBase())
}

// Anki/Hosts/U/{hostID}/S/discover
func (h Host) StatusDiscoverTopic() string {
	return fmt.Sprintf("%s/discover", h.StatusTopic())
}

// statusDITBase constructs the base path for DIT statuses.
func (h Host) statusDITBase() string {
	return fmt.Sprintf("%s/DIT", h.StatusTopic())
}

// Anki/Hosts/U/{hostID}/S/DIT/discoverSubscription
func (h Host) StatusDITDiscoverSubscriptionTopic() string {
	return fmt.Sprintf("%s/discoverSubscription", h.statusDITBase())
}

// eventsBase constructs the base path for events.
func (h Host) eventsBase() string {
	return fmt.Sprintf("%s/%s/E", HostsTopicBase, h.Id())
}

// Anki/Hosts/U/{hostID}/E/scanning
func (h Host) EventScanningTopic() string {
	return fmt.Sprintf("%s/scanning", h.eventsBase())
}

// Anki/Hosts/U/{hostID}/E/vehicle/discovered/{vehicleID}
func (h Host) EventVehicleDiscoveredTopic(vehicleID string) string {
	return fmt.Sprintf("%s/vehicle/discovered/%s", h.eventsBase(), vehicleID)
}

type Vehicle string

func (v Vehicle) Id() string {
	return string(v)
}

// Anki/Vehicles/U/{vehicleID}/I/{callerID}
func (v Vehicle) IntentTopic() string {
	return fmt.Sprintf("%s/%s/I/%s", VehiclesTopicBase, v.Id(), callerID)
}

// Anki/Vehicles/U/{vehicleID}/S
func (v Vehicle) StatusTopic() string {
	return fmt.Sprintf("%s/%s/S", VehiclesTopicBase, v.Id())
}

// Anki/Vehicles/U/{vehicleID}/S/online
func (v Vehicle) StatusOnlineTopic() string {
	return fmt.Sprintf("%s/online", v.StatusTopic())
}

// statusIntendedBase constructs the base path for intended statuses.
func (v Vehicle) statusIntendedBase() string {
	return fmt.Sprintf("%s/intended", v.StatusTopic())
}

// Anki/Vehicles/U/{vehicleID}/S/intended/connect
func (v Vehicle) StatusIntendedConnectTopic() string {
	return fmt.Sprintf("%s/connect", v.statusIntendedBase())
}

// Anki/Vehicles/U/{vehicleID}/S/intended/lights
func (v Vehicle) StatusIntendedLightsTopic() string {
	return fmt.Sprintf("%s/lights", v.statusIntendedBase())
}

// Anki/Vehicles/U/{vehicleID}/S/intended/lane
func (v Vehicle) StatusIntendedLaneTopic() string {
	return fmt.Sprintf("%s/lane", v.statusIntendedBase())
}

// Anki/Vehicles/U/{vehicleID}/S/intended/speed
func (v Vehicle) StatusIntendedSpeedTopic() string {
	return fmt.Sprintf("%s/speed", v.statusIntendedBase())
}

// Anki/Vehicles/U/{vehicleID}/S/intended/cancelLane
func (v Vehicle) StatusIntendedCancelLaneTopic() string {
	return fmt.Sprintf("%s/cancelLane", v.statusIntendedBase())
}

func (v Vehicle) StatusIntendedConnectSubscriptionTopic() string {
	return fmt.Sprintf("%s/connectSubscription", v.statusIntendedBase())
}

func (v Vehicle) StatusIntendedLightsSubscriptionTopic() string {
	return fmt.Sprintf("%s/lightsSubscription", v.statusIntendedBase())
}

func (v Vehicle) StatusIntendedLaneSubscriptionTopic() string {
	return fmt.Sprintf("%s/laneSubscription", v.statusIntendedBase())
}

func (v Vehicle) StatusIntendedSpeedSubscriptionTopic() string {
	return fmt.Sprintf("%s/speedSubscription", v.statusIntendedBase())
}

func (v Vehicle) StatusIntendedCancelLaneSubscriptionTopic() string {
	return fmt.Sprintf("%s/cancelLaneSubscription", v.statusIntendedBase())
}

// Anki/Vehicles/U/{vehicleID}/S/status
func (v Vehicle) StatusStatusTopic() string {
	return fmt.Sprintf("%s/status", v.StatusTopic())
}

// Anki/Vehicles/U/{vehicleID}/S/onTrack
func (v Vehicle) StatusOnTrackTopic() string {
	return fmt.Sprintf("%s/onTrack", v.StatusTopic())
}

// Topic: Anki/Vehicles/U/{vehicleID}/S/version
func (v Vehicle) StatusVersionTopic() string {
	return fmt.Sprintf("%s/version", v.StatusTopic())
}

// --- Vehicle Battery Statuses ---

// StatusBatteryBase constructs the base path for battery statuses.
func (v Vehicle) StatusBatteryBase() string {
	return fmt.Sprintf("%s/battery", v.StatusTopic())
}

// StatusBatteryChargingTopic constructs the topic for ChargingStatus.
// Topic: Anki/Vehicles/U/{vehicleID}/S/battery/charging
func (v Vehicle) StatusBatteryChargingTopic() string {
	return fmt.Sprintf("%s/charging", v.StatusBatteryBase())
}

// StatusBatteryLevelTopic constructs the topic for LevelStatus.
// Topic: Anki/Vehicles/U/{vehicleID}/S/battery/level
func (v Vehicle) StatusBatteryLevelTopic() string {
	return fmt.Sprintf("%s/level", v.StatusBatteryBase())
}

// --- Vehicle DIT Statuses ---

// StatusDITBase constructs the base path for DIT statuses.
func (v Vehicle) StatusDITBase() string {
	return fmt.Sprintf("%s/DIT", v.StatusTopic())
}

// StatusDITConnectSubscriptionTopic constructs the topic for ConnectSubscriptionStatus.
func (v Vehicle) StatusDITConnectSubscriptionTopic() string {
	return fmt.Sprintf("%s/connectSubscription", v.StatusDITBase())
}

func (v Vehicle) StatusDITLightsSubscriptionTopic() string {
	return fmt.Sprintf("%s/lightsSubscription", v.StatusDITBase())
}

func (v Vehicle) StatusDITLaneSubscriptionTopic() string {
	return fmt.Sprintf("%s/laneSubscription", v.StatusDITBase())
}

func (v Vehicle) StatusDITSpeedSubscriptionTopic() string {
	return fmt.Sprintf("%s/speedSubscription", v.StatusDITBase())
}

func (v Vehicle) StatusDITCancelLaneSubscriptionTopic() string {
	return fmt.Sprintf("%s/cancelLaneSubscription", v.StatusDITBase())
}

// --- Vehicle Events ---

// EventsBase constructs the base path for events.
func (v Vehicle) EventsBase() string {
	return fmt.Sprintf("%s/%s/E", VehiclesTopicBase, v.Id())
}

// EventConnectionTopic constructs the topic for ConnectionEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/connection
func (v Vehicle) EventConnectionTopic() string {
	return fmt.Sprintf("%s/connection", v.EventsBase())
}

// EventRssiTopic constructs the topic for RssiEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/rssi
func (v Vehicle) EventRssiTopic() string {
	return fmt.Sprintf("%s/rssi", v.EventsBase())
}

// EventOffsetTopic constructs the topic for OffsetEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/offset
func (v Vehicle) EventOffsetTopic() string {
	return fmt.Sprintf("%s/offset", v.EventsBase())
}

// EventLineDriftTopic constructs the topic for LineDriftEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/lineDrift
func (v Vehicle) EventLineDriftTopic() string {
	return fmt.Sprintf("%s/lineDrift", v.EventsBase())
}

// EventHillCounterTopic constructs the topic for HillCounterEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/hillCounter
func (v Vehicle) EventHillCounterTopic() string {
	return fmt.Sprintf("%s/hillCounter", v.EventsBase())
}

// EventWheelDistanceTopic constructs the topic for WheelDistanceEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/wheelDistance
func (v Vehicle) EventWheelDistanceTopic() string {
	return fmt.Sprintf("%s/wheelDistance", v.EventsBase())
}

// EventTrackTopic constructs the topic for TrackEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/track
func (v Vehicle) EventTrackTopic() string {
	return fmt.Sprintf("%s/track", v.EventsBase())
}

// EventSpeedTopic constructs the topic for SpeedEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/speed
func (v Vehicle) EventSpeedTopic() string {
	return fmt.Sprintf("%s/speed", v.EventsBase())
}

// EventLaneChangedTopic constructs the topic for LaneChangedEvent.
// Topic: Anki/Vehicles/U/{vehicleID}/E/laneChanged
func (v Vehicle) EventLaneChangedTopic() string {
	return fmt.Sprintf("%s/laneChanged", v.EventsBase())
}

// --- Global Intent Topics (For all hosts/vehicles) ---

// IntentTopicAllHosts constructs the topic for a caller sending an intent to ALL hosts.
// Topic: Anki/Hosts/U/I/{callerID}
func IntentTopicAllHosts() string {
	return fmt.Sprintf("%s/I/%s", HostsTopicBase, callerID)
}

// IntentTopicAllVehicles constructs the topic for a caller sending an intent to ALL vehicles.
// Topic: Anki/Vehicles/U/I/{callerID}
func IntentTopicAllVehicles() string {
	return fmt.Sprintf("%s/I/%s", VehiclesTopicBase, callerID)
}
