"""
RemoteControl.py — Simplified MQTT remote control

Controls (Windows):
  W / Up arrow    : increase speed
  S / Down arrow  : decrease speed
  A / Left arrow  : steer left
  D / Right arrow : steer right
  Space           : stop (speed 0)
  Q               : quit
"""

import json
import sys
import time
import argparse
import paho.mqtt.client as mqtt

# Windows-only key capture
try:
    import msvcrt
except ImportError:
    msvcrt = None

# MQTT settings
DEFAULT_BROKER = "10.42.0.1"
DEFAULT_PORT = 1883
SPEED_TOPIC = "RemoteControl/U/E/vehicles/speed"
LANE_TOPIC = "RemoteControl/U/E/vehicles/lane"
WILL_TOPIC = "RemoteControl/U/S/online"

class RemoteControl:
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.client = mqtt.Client()
        self.velocity = 0
        self.offset = 0
        self.acceleration = 200
        self.lane_velocity = 300
        self.lane_acceleration = 300
        self._running = False

        self.client.on_connect = lambda c, u, f, rc: print(f"Connected (rc={rc})")
        self.client.on_disconnect = lambda c, u, rc: print(f"Disconnected (rc={rc})")

        self.client.will_set(WILL_TOPIC, json.dumps({"value": "false"}), qos=1, retain=True)

    def connect(self):
        self.client.connect(self.host, self.port)
        self.client.loop_start()
        self.client.publish(WILL_TOPIC, json.dumps({"value": "true"}), qos=1, retain=True)

    def disconnect(self):
        self.client.publish(WILL_TOPIC, json.dumps({"value": "false"}), qos=1, retain=True)
        self.client.loop_stop()
        self.client.disconnect()

    def publish_speed(self):
        payload = f"velocity: {self.velocity}\nacceleration: {self.acceleration}"
        self.client.publish(SPEED_TOPIC, payload)
        print(f"[PUB] Speed -> {self.velocity}")

    def publish_lane(self):
        payload = f"velocity: {self.lane_velocity}\nacceleration: {self.lane_acceleration}\noffset: {self.offset}"
        self.client.publish(LANE_TOPIC, payload)
        print(f"[PUB] Lane -> offset={self.offset}")

    def stop(self):
        self.velocity = 0
        self.publish_speed()


def read_keys_loop(ctrl, step_speed=50, step_offset=15):
    if msvcrt is None:
        print("msvcrt not available (Windows only).")
        return

    print("Controls: W/S or ↑/↓ = speed, A/D or ←/→ = steer, Space = stop, Q = quit")

    last_publish = 0
    publish_interval = 0.05

    while ctrl._running:
        if msvcrt.kbhit():
            ch = msvcrt.getch()
            if ch in (b"\x00", b"\xe0"):  # Arrow keys
                key = msvcrt.getch()
                if key == b'H':  # Up
                    ctrl.velocity += step_speed
                elif key == b'P':  # Down
                    ctrl.velocity = max(0, ctrl.velocity - step_speed)
                elif key == b'K':  # Left
                    ctrl.offset -= step_offset
                elif key == b'M':  # Right
                    ctrl.offset += step_offset
            else:
                key = ch.decode().lower()
                if key == 'w':
                    ctrl.velocity += step_speed
                elif key == 's':
                    ctrl.velocity = max(0, ctrl.velocity - step_speed)
                elif key == 'a':
                    ctrl.offset -= step_offset
                elif key == 'd':
                    ctrl.offset += step_offset
                elif key == ' ':
                    ctrl.stop()
                elif key == 'q':
                    ctrl._running = False
                    break

        ctrl.velocity = min(max(ctrl.velocity, 0), 1000)
        ctrl.offset = min(max(ctrl.offset, -100), 100)

        now = time.time()
        if now - last_publish > publish_interval:
            ctrl.publish_speed()
            ctrl.publish_lane()
            last_publish = now

        time.sleep(0.01)


def main():
    parser = argparse.ArgumentParser(description='Simple MQTT RemoteControl')
    parser.add_argument('--broker', default=DEFAULT_BROKER)
    parser.add_argument('--port', type=int, default=DEFAULT_PORT)
    args = parser.parse_args()

    ctrl = RemoteControl(args.broker, args.port)

    try:
        ctrl.connect()
    except Exception as e:
        print(f"Failed to connect: {e}")
        sys.exit(1)

    ctrl._running = True
    try:
        read_keys_loop(ctrl)
    except KeyboardInterrupt:
        pass
    finally:
        ctrl.stop()
        time.sleep(0.2)
        ctrl.disconnect()
        print("Exited.")


if __name__ == '__main__':
    main()