# Hyperdrive: Event-Driven Architecture with MQTT using DUISE

Now, we will delve into the Hyperdrive-API. It is written as part of the fascinating world of Event-Driven Architecture (EDA) again using MQTT as our messaging protocol. We'll recapitulate on the unique concept called **DUISE**—which stands for **Documentation, Unit, Intent, Status, Event**—and see how it is applied on the Hyperdrive-API to provide structured topic hierarchies and message flows in an MQTT-based system.

Our focus will be on understanding how DUISE facilitates communication between services (actors) and clients in a distributed system, specifically within the context of the **Hyperdrive API**. We'll also examine an example that integrates a **RemoteControl** service using dynamic subscriptions.

---

## Table of Contents

1. **Introduction to Event-Driven Architecture and MQTT**
2. **The DUISE Concept**
   - Documentation
   - Unit
   - Intent
   - Status
   - Event
3. **Intent Messaging in DUISE**
   - Sending Commands to All Units
   - Targeting Specific Units
   - Debugging with Caller IDs
4. **Status and Event Messages**
   - Retained vs. Non-Retained Messages
5. **The Hyperdrive API**
   - Supporting Message-Driven and Event-Driven Architectures
   - Adapting Message Structures per Intent Command
6. **Topic Structure Breakdown**
   - Hosts
   - Vehicles
7. **Message Formats**
   - Host Messages
   - Vehicle Messages
8. **Example: Integrating RemoteControl with Dynamic Subscriptions**
9. **Conclusion**

---

## 1. Introduction to Event-Driven Architecture and MQTT

**Event-Driven Architecture (EDA)** is a design paradigm where the flow of the program is determined by events such as user actions, sensor outputs, or message arrivals. EDA decouples event producers from event consumers, enhancing scalability and flexibility.

**MQTT (Message Queuing Telemetry Transport)** is a lightweight, publish-subscribe network protocol that transports messages between devices. It's ideal for scenarios where bandwidth is limited, making it a popular choice for IoT applications.

---

## 2. The DUISE Concept

**DUISE** is a custom framework designed to structure MQTT topics and messages systematically. It stands for:

- **D**ocumentation
- **U**nit
- **I**ntent
- **S**tatus
- **E**vent

### Documentation

The **Documentation** layer serves as a reference point for developers and system integrators. It outlines the structure, expected message formats, and usage guidelines for the MQTT topics.

### Unit

The **Unit** represents individual services or actors within the system. Each unit can be a host, a vehicle, or any other entity that performs actions or processes data.

### Intent

**Intent** is used for sending commands to services (actors). It encapsulates the desired action or command that a client wants the service to perform.

### Status

**Status** messages represent the internal state of a service. These messages are typically **retained** so that clients can retrieve the latest status upon subscription.

### Event

**Event** messages are emitted when certain conditions within a service are met, triggering an event. These messages are **not retained**, meaning subscribers will only receive them if they are actively listening at the time the event occurs.

---

## 3. Intent Messaging in DUISE

### Sending Commands to All Units

Intents can be broadcasted to all units by publishing to a general topic. For example:

```
.../U/I
```

All units subscribed to this topic will receive the command.

### Targeting Specific Units

To send a command to a specific unit (e.g., a particular vehicle or host), you publish to a unit-specific topic:

```
.../U/<unitID>/I
```

Only the unit with the matching `<unitID>` will process the command.

### Debugging with Caller IDs

Including a `{callerID}` in the subtopic helps trace who sent a command, which is invaluable for debugging:

```
.../U/I/{callerID}
```

This structure allows you to monitor and log interactions between clients and services.

---

## 4. Status and Event Messages

### Retained vs. Non-Retained Messages

- **Status Messages**: Represent the current state of a service and are **retained**. Subscribers receive the latest status upon subscribing.

- **Event Messages**: Indicate that an event has occurred. They are **not retained**, so only active subscribers at the time of the event will receive them.

---

## 5. The Hyperdrive API

The **Hyperdrive API** supports both **message-driven** and **event-driven** architectures. It allows clients to control services via commands (intents) and to react to events emitted by services.

### Supporting Message-Driven and Event-Driven Architectures

Clients can interact with services in two ways:

1. **Message-Driven**: Sending commands directly to services using intent messages.
2. **Event-Driven**: Subscribing to event topics to react to service-generated events.

### Adapting Message Structures per Intent Command

Each intent command may have a specific message structure. It's crucial to format messages correctly to ensure services interpret commands as intended.

---

## 6. Topic Structure Breakdown

Let's explore the topic hierarchy for **Hosts** and **Vehicles**.

### Hosts

```
Anki/
  Hosts/
    D/
    U/
      I/
        {callerID} <HostIntent> [Receive] // For all host IDs
      {hostID}/
        I/
          {callerID} <HostIntent> [Receive]
        S/
          online <OnlineStatus> [Send]
          intended/
            discover <DiscoverIntentStatus> [Send]
            discoverSubscription <DiscoverSubscriptionIntentStatus> [Send]
          discover <DiscoverStatus> [Send]
          DIT/
            discoverSubscription <DiscoverSubscriptionStatus> [Send]
        E/
          scanning <ScanningEvent> [Send]
          vehicle/
            discovered/
              {vehicleID} <VehicleDiscoveredEvent>... [Send]
```

### Vehicles

```
Anki/
  Vehicles/
    D/
    U/
      I/
        {callerID} <VehicleIntent> [Receive] // For all vehicle IDs
      {vehicleID}/
        I/
          {callerID} <VehicleIntent> [Receive]
        S/
          online <OnlineStatus> [Send]
          intended/
            connect <ConnectIntentStatus> [Send]
            lights <LightsIntentStatus> [Send]
            lane <LaneIntentStatus> [Send]
            speed <SpeedIntentStatus> [Send]
            cancelLane <CancelLaneIntentStatus> [Send]
            connectSubscription <ConnectSubscriptionIntentStatus> [Send]
            lightsSubscription <LightsSubscriptionIntentStatus> [Send]
            laneSubscription <LaneSubscriptionIntentStatus> [Send]
            speedSubscription <SpeedSubscriptionIntentStatus> [Send]
            cancelLaneSubscription <CancelLaneSubscriptionIntentStatus> [Send]
          status <StatusStatus> [Send]
          DIT/
            connectSubscription <ConnectSubscriptionStatus> [Send]
            lightsSubscription <LightsSubscriptionStatus> [Send]
            laneSubscription <LaneSubscriptionStatus> [Send]
            speedSubscription <SpeedSubscriptionStatus> [Send]
            cancelLaneSubscription <CancelLaneSubscriptionStatus> [Send]
          onTrack <OnTrackStatus> [Send]
          battery/
            charging <ChargingStatus> [Send]
            level <LevelStatus> [Send]
          version <VersionStatus> [Send]
        E/
          connection <ConnectionEvent>... [Send]
          rssi <RssiEvent>... [Send]
          offset <OffsetEvent>... [Send]
          lineDrift <LineDriftEvent>... [Send]
          hillCounter <HillCounterEvent>... [Send]
          wheelDistance <WheelDistanceEvent>... [Send]
          track <TrackEvent>... [Send]
          speed <SpeedEvent>... [Send]
          laneChanged <LaneChangedEvent>... [Send]
```

---

## 7. Message Formats

Understanding the message structures is crucial for effective communication between clients and services.

### Host Messages

#### **HostIntent**

Commands sent to the host service to control vehicle discovery.

```yaml
- type: discover
  payload:
    value: {true|false} # Default: false

- type: discoverSubscription
  payload:
    topic: {topic-filter} # Default: null
    subscribe: {true|false} # Default: false
```

#### **OnlineStatus** and **DiscoverIntentStatus**

Indicate whether the service is online or if the discovery intent is active.

```yaml
value: {true|false}
```

#### **DiscoverStatus** and **ScanningEvent**

Provide timestamps and status of the discovery process.

```yaml
timestamp: {nanos}
value: {true|false}
```

#### **VehicleDiscoveredEvent**

Details about a newly discovered vehicle.

```yaml
timestamp: {nanos}
value:
  model: {model-name}
  rssi: {-127...0}
```

#### **DiscoverSubscriptionIntentStatus**

Status of the discovery subscription intent.

```yaml
timestamp: {nanos}
value:
  topic: {topic-filter}
  subscribe: {true|false}
```

#### **DiscoverSubscriptionStatus**

List of active discovery subscriptions.

```yaml
timestamp: {nanos}
value: {topic-filter}...
```

### Vehicle Messages

#### **VehicleIntent**

Commands to control vehicle actions.

```yaml
- type: connect
  payload:
    value: {true|false} # Default: false

- type: speed
  payload:
    velocity: {-100...1000} # Default: 0
    acceleration: {0...2000} # Default: 0

- type: lane
  payload:
    velocity: {0...1000} # Default: 0
    acceleration: {0...1000} # Default: 0
    offset: {-100.0...100.0} # Default: 0.0
    offsetFromCenter: {-100.0...100.0} # Default: 0.0

- type: cancelLane
  payload:
    value: {true|false} # Default: false

- type: lights
  payload:
    frontGreen: {<LightEffect>}
    frontRed: {<LightEffect>}
    tail: {<LightEffect>}
    engineRed: {<LightEffect>}
    engineGreen: {<LightEffect>}
    engineBlue: {<LightEffect>}

- type: [connect|speed|lane|cancelLane|lights]Subscription
  payload:
    topic: {topic-filter} # Default: null
    subscribe: {true|false} # Default: false
```

#### **OnlineStatus** and **ConnectIntentStatus**

Show the connection status of the vehicle.

```yaml
value: {true|false}
```

#### **ChargingStatus**, **OnTrackStatus**, and **ConnectionEvent**

```yaml
timestamp: {nanos}
value: {true|false}
```

#### **LevelStatus**

Battery level information.

```yaml
timestamp: {nanos}
value: {0...100}
```

#### **VersionStatus**

Firmware or software version of the vehicle.

```yaml
timestamp: {nanos}
value: {Version-String}
```

#### **RssiEvent**

Signal strength information.

```yaml
timestamp: {nanos}
value: {-127...0}
```

#### **LightEffect**

Defines the lighting effect for vehicle lights.

```yaml
effect: {off|steady|fade|pulse|flash|strobe}
start: {0...15}
end: {0...15}
frequency: {0...255}
```

#### **LightsIntentStatus**

Current lighting settings of the vehicle.

```yaml
timestamp: {nanos}
value:
  frontRed: {<LightEffect> | null}
  frontGreen: {<LightEffect> | null}
  tail: {<LightEffect> | null}
  engineRed: {<LightEffect> | null}
  engineGreen: {<LightEffect> | null}
  engineBlue: {<LightEffect> | null}
```

#### **SpeedIntentStatus**

Current speed settings.

```yaml
timestamp: {nanos}
value:
  velocity: {-100...1000}
  acceleration: {0...2000}
```

#### **LaneIntentStatus**

Current lane change settings.

```yaml
timestamp: {nanos}
value:
  velocity: {0...1000}
  acceleration: {0...2000}
  offsetFromCenter: {-100...100}
  offset: {-100...100}
```

#### **CancelLaneIntentStatus**

Indicates whether a lane change has been canceled.

```yaml
timestamp: {nanos}
value: {true|false}
```

#### **StatusStatus**

Overall status of the vehicle's connection.

```yaml
timestamp: {nanos}
value: {connected | disconnected | lost}
```

#### **[Intent]SubscriptionIntentStatus**

Status of subscription intents.

```yaml
timestamp: {nanos}
value:
  topic: {topic-filter}
  subscribe: {true|false}
```

#### **[Intent]SubscriptionStatus**

List of active subscriptions.

```yaml
timestamp: {nanos}
value: {topic-filter}...
```

---

## 8. Example: Integrating RemoteControl with Dynamic Subscriptions

Let's see how we can integrate a `RemoteControl` service with our system using dynamic subscriptions. This works, if the creator of the `RemoteControl` accepts the messages required by `Hyperdrive`. This way:

#### **RemoteControl**
```
RemoteControl/
  U/
    <RemoteControlID>/
     E/
        hosts/
          discover <HostIntent-Payload> [Send]
        vehicles/
          connect <VehicleIntent-Connect-Payload> [Send]
          lights <VehicleIntent-Lights-Payload> [Send]
          laneChange <VehicleIntent-LaneChange-Payload> [Send]
          speed <VehicleIntent-Speed-Payload> [Send]
          panic <VehicleIntent-Speed-Payload> & <VehicleIntent-Lights-Payload> [Send]

```

### Scenario

We have a `RemoteControl` service that publishes commands on its own topics. We want our `Hyperdrive` services to listen to these commands without hardcoding the topics.

### Solution

Use dynamic subscriptions to subscribe to the `RemoteControl` topics.

#### Step 1: Subscribing Hosts to Discovery Events

Publish an intent to the `Anki/Hosts/U/I` topic to subscribe to the `RemoteControl` discovery events.

```yaml
- type: discoverSubscription
  payload:
    topic: RemoteControl/+/E/hosts/discover
    subscribe: true
```

#### Step 2: Subscribing Vehicles to Control Events

Publish intents to `Anki/Vehicles/U/I` to subscribe to various `RemoteControl` events.

```yaml
- type: connectSubscription
  payload:
    topic: RemoteControl/+/E/vehicles/connect/#
    subscribe: true

- type: lightsSubscription
  payload:
    topic: RemoteControl/+/E/vehicles/lights/#
    subscribe: true

- type: laneSubscription
  payload:
    topic: RemoteControl/+/E/vehicles/laneChange/#
    subscribe: true

- type: speedSubscription
  payload:
    topic: RemoteControl/+/E/vehicles/speed/#
    subscribe: true

- type: speedSubscription
  payload:
    topic: RemoteControl/+/E/vehicles/panic/#
    subscribe: true

- type: lightsSubscription
  payload:
    topic: RemoteControl/+/E/vehicles/panic/#
    subscribe: true
```

### Explanation

- **Dynamic Subscriptions**: By subscribing to topics at runtime, services can react to external events without prior knowledge of the topics.
- **Topic Filters**: The `+` and `#` wildcards allow for flexible topic matching.
- **Integration**: This approach seamlessly integrates the `RemoteControl` service into our system.

---

## 9. Conclusion

We've explored how the DUISE concept structures MQTT topics and messages to support both message-driven and event-driven architectures. By understanding the topic hierarchy and message formats, we can design scalable and flexible systems.

Key takeaways:

- **DUISE Framework**: Provides a systematic way to organize topics and messages.
- **Intents**: Allow clients to send commands to services, either globally or to specific units.
- **Status and Events**: Offer insight into the internal state and occurrences within services.
- **Dynamic Subscriptions**: Enable services to adapt to new topics and integrate with other systems seamlessly.

---





