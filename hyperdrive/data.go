package hyperdrive

// --- Shared Base Structs for Composition ---

// Base for statuses and events that only have a true/false value and a timestamp.
type ValueStatusEvent struct {
	Timestamp int64 `json:"timestamp,omitempty"`
	Value     bool  `json:"value"`
}

// Base for IntentStatus messages that confirm a subscription setup.
type SubscriptionIntentStatusBase struct {
	Timestamp int64 `json:"timestamp,omitempty"`
	Value     struct {
		Topic     string `json:"topic"`
		Subscribe bool   `json:"subscribe"`
	} `json:"value"`
}

// Base for DIT Status messages that confirm a subscription is active.
type SubscriptionStatusBase struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	Value     string `json:"value"` // Represents the {topic-filter}... string
}

// --- Host Messages ---

// HostIntent represents the command structure sent to a Host.
type HostIntent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"` // Payload structure varies by Type
}

// DiscoverIntentPayload is for HostIntent type "discover".
type DiscoverIntentPayload struct {
	Value bool `json:"value"` // default: false
}

// DiscoverSubscriptionIntentPayload is for HostIntent type "discoverSubscription".
type DiscoverSubscriptionIntentPayload struct {
	Topic     string `json:"topic"`     // default: null
	Subscribe bool   `json:"subscribe"` // default: false
}

// OnlineStatus is used for "online" topics.
type OnlineStatus struct {
	Value bool `json:"value"`
}

// DiscoverIntentStatus is used for "intended/discover" topics.
type DiscoverIntentStatus struct {
	Value bool `json:"value"`
}

// DiscoverStatus is used for "discover" topics.
// Also used for ScanningEvent.
type DiscoverStatus struct {
	Timestamp int64 `json:"timestamp"`
	Value     bool  `json:"value"`
}

// ScanningEvent is used for "E/scanning" topics.
// It shares the same structure as DiscoverStatus.
type ScanningEvent DiscoverStatus

// DiscoverSubscriptionIntentStatus is used for "intended/discoverSubscription" topics.
type DiscoverSubscriptionIntentStatus SubscriptionIntentStatusBase

// DiscoverSubscriptionStatus is used for "DIT/discoverSubscription" topics.
type DiscoverSubscriptionStatus SubscriptionStatusBase

// VehicleDiscoveredEvent is used for "E/vehicle/discovered/{vehicleID}" topics.
type VehicleDiscoveredEvent struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		Model string `json:"model"`
		Rssi  int    `json:"rssi"` // Range: -127 to 0
	} `json:"value"`
}

// --- Vehicle Messages ---

// VehicleIntent represents the command structure sent to a Vehicle.
type VehicleIntent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"` // Payload structure varies by Type
}

// ConnectIntentPayload is for VehicleIntent type "connect".
type ConnectIntentPayload struct {
	Value bool `json:"value"` // default: false
}

// SpeedIntentPayload is for VehicleIntent type "speed".
type SpeedIntentPayload struct {
	Velocity     int `json:"velocity"`     // Range: -100 to 1000, default: 0
	Acceleration int `json:"acceleration"` // Range: 0 to 2000, default: 0
}

// LaneIntentPayload is for VehicleIntent type "lane".
type LaneIntentPayload struct {
	Velocity         int     `json:"velocity"`         // Range: 0 to 1000, default: 0
	Acceleration     int     `json:"acceleration"`     // Range: 0 to 1000, default: 0
	Offset           float64 `json:"offset"`           // Range: -100.0 to 100.0, default: 0.0
	OffsetFromCenter float64 `json:"offsetFromCenter"` // Range: -100.0 to 100.0, default: 0.0
}

// CancelLaneIntentPayload is for VehicleIntent type "cancelLane".
type CancelLaneIntentPayload struct {
	Value bool `json:"value"` // default: false
}

// LightEffect represents the detailed structure for a single light setting.
type LightEffect struct {
	Effect    string `json:"effect"`    // Values: off|steady|fade|pulse|flash|strobe
	Start     int    `json:"start"`     // Range: 0 to 15, default: 0
	End       int    `json:"end"`       // Range: 0 to 15, default: 0
	Frequency int    `json:"frequency"` // Range: 0 to 255, default: 0
}

// LightsIntentPayload is for VehicleIntent type "lights".
type LightsIntentPayload struct {
	FrontGreen  LightEffect `json:"frontGreen"`
	FrontRed    LightEffect `json:"frontRed"`
	Tail        LightEffect `json:"tail"`
	EngineRed   LightEffect `json:"engineRed"`
	EngineGreen LightEffect `json:"engineGreen"`
	EngineBlue  LightEffect `json:"engineBlue"`
}

// SubscriptionIntentPayload (General) is used for all vehicle subscription intents.
type SubscriptionIntentPayload struct {
	Topic     string `json:"topic"`     // default: null
	Subscribe bool   `json:"subscribe"` // default: false
}

// --- Vehicle Statuses and Events ---

// ConnectIntentStatus is used for "intended/connect" topics.
type ConnectIntentStatus struct {
	Value bool `json:"value"`
}

// ChargingStatus is used for "battery/charging" topics.
// Also used for OnTrackStatus and ConnectionEvent.
type ChargingStatus ValueStatusEvent

// OnTrackStatus is used for "onTrack" topics.
type OnTrackStatus ValueStatusEvent

// ConnectionEvent is used for "E/connection" topics.
type ConnectionEvent ValueStatusEvent

// LevelStatus is used for "battery/level" topics.
type LevelStatus struct {
	Timestamp int64 `json:"timestamp"`
	Value     int   `json:"value"` // Range: 80 to 100 (Battery Percentage)
}

// VersionStatus is used for "version" and "E/offset" (in documentation) topics.
type VersionStatus struct {
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value"` // Represents {Version-String}
}

// RssiEvent is used for "E/rssi" topics.
type RssiEvent struct {
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value"` // Represents {0|Infinity}
}

// LineDriftEvent is used for "E/lineDrift" topics.
type LineDriftEvent struct {
	Timestamp int64 `json:"timestamp"`
	Value     int   `json:"value"` // Range: 0 to 255
}

// HillCounterEvent is used for "E/hillCounter" topics.
type HillCounterEvent struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		Up   int `json:"up"`   // Range: 0 to 255
		Down int `json:"down"` // Range: 0 to 255
	} `json:"value"`
}

// WheelDistanceEvent is used for "E/wheelDistance" topics.
type WheelDistanceEvent struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		Left  int `json:"left"`  // Range: 0 to 255
		Right int `json:"right"` // Range: 0 to 255
	} `json:"value"`
}

// TrackEvent is used for "E/track" topics.
type TrackEvent struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		TrackID       int    `json:"trackID"`       // Range: 0 to 127
		TrackLocation int    `json:"trackLocation"` // Range: 0 to 127
		Direction     string `json:"direction"`     // Values: left|right
	} `json:"value"`
}

// SpeedEvent is used for "E/speed" topics.
type SpeedEvent struct {
	Timestamp int64 `json:"timestamp"`
	Value     int   `json:"value"` // Range: -100 to 1000
}

// LaneChangedEvent is used for "E/laneChanged" topics.
type LaneChangedEvent struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"` // Range: -100.0 to 100.0 (Offset)
}

// LightsIntentStatus is used for "intended/lights" topics.
type LightsIntentStatus struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		FrontRed    *LightEffect `json:"frontRed"` // Pointer/Optional because value is {<LightEffect> | null}
		FrontGreen  *LightEffect `json:"frontGreen"`
		Tail        *LightEffect `json:"tail"`
		EngineRed   *LightEffect `json:"engineRed"`
		EngineGreen *LightEffect `json:"engineGreen"`
		EngineBlue  *LightEffect `json:"engineBlue"`
	} `json:"value"`
}

// SpeedIntentStatus is used for "intended/speed" topics.
type SpeedIntentStatus struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		Velocity     int `json:"velocity"`
		Acceleration int `json:"acceleration"`
	} `json:"value"`
}

// LaneIntentStatus is used for "intended/lane" topics.
type LaneIntentStatus struct {
	Timestamp int64 `json:"timestamp"`
	Value     struct {
		Velocity         int     `json:"velocity"`
		Acceleration     int     `json:"acceleration"`
		OffsetFromCenter float64 `json:"offsetFromCenter"`
		Offset           float64 `json:"offset"`
	} `json:"value"`
}

// CancelLaneIntentStatus is used for "intended/cancelLane" topics.
type CancelLaneIntentStatus struct {
	Timestamp int64 `json:"timestamp"`
	Value     bool  `json:"value"`
}

// StatusStatus is used for "status" topics.
type StatusStatus struct {
	Timestamp int64  `json:"timestamp"`
	Value     string `json:"value"` // Values: connected | disconnected | lost
}

// --- Vehicle DIT and Subscription Intent Statuses (Shared Structs) ---

// ConnectSubscriptionIntentStatus is used for "intended/connectSubscription" topics.
type ConnectSubscriptionIntentStatus SubscriptionIntentStatusBase

// SpeedSubscriptionIntentStatus is used for "intended/speedSubscription" topics.
type SpeedSubscriptionIntentStatus SubscriptionIntentStatusBase

// LaneSubscriptionIntentStatus is used for "intended/laneSubscription" topics.
type LaneSubscriptionIntentStatus SubscriptionIntentStatusBase

// CancelLaneSubscriptionIntentStatus is used for "intended/cancelLaneSubscription" topics.
type CancelLaneSubscriptionIntentStatus SubscriptionIntentStatusBase

// LightsSubscriptionIntentStatus is used for "intended/lightsSubscription" topics.
type LightsSubscriptionIntentStatus SubscriptionIntentStatusBase

// ConnectSubscriptionStatus is used for "DIT/connectSubscription" topics.
type ConnectSubscriptionStatus SubscriptionStatusBase

// SpeedSubscriptionStatus is used for "DIT/speedSubscription" topics.
type SpeedSubscriptionStatus SubscriptionStatusBase

// LaneSubscriptionStatus is used for "DIT/laneSubscription" topics.
type LaneSubscriptionStatus SubscriptionStatusBase

// CancelLaneSubscriptionStatus is used for "DIT/cancelLaneSubscription" topics.
type CancelLaneSubscriptionStatus SubscriptionStatusBase

// LightsSubscriptionStatus is used for "DIT/lightsSubscription" topics.
type LightsSubscriptionStatus SubscriptionStatusBase
