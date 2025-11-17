"""
RemoteControl_test.py — Offline test version (no MQTT)

Controls (Windows):
  W / Up arrow    : increase speed
  S / Down arrow  : decrease speed
  A / Left arrow  : steer left
  D / Right arrow : steer right
  Space           : stop (speed 0)
  Q               : quit
"""

import time

# Windows-only key capture
try:
    import msvcrt
except ImportError:
    msvcrt = None


class RemoteControlTest:
    def __init__(self):
        self.velocity = 0
        self.offset = 0
        self.acceleration = 200
        self.lane_velocity = 300
        self.lane_acceleration = 300
        self._running = True

    def publish_speed(self):
        print(f"[TEST-MQTT] SPEED | velocity={self.velocity}, acceleration={self.acceleration}")

    def publish_lane(self):
        print(f"[TEST-MQTT] LANE  | offset={self.offset}, velocity={self.lane_velocity}")

    def stop(self):
        self.velocity = 0
        print("[ACTION] Stop (velocity=0)")
        self.publish_speed()


def read_keys_loop(ctrl, step_speed=50, step_offset=15):
    if msvcrt is None:
        print("msvcrt not available (Windows only).")
        return

    print("=== TEST MODE ===")
    print("Controls: W/S or ↑/↓ = speed, A/D or ←/→ = steer, Space = stop, Q = quit\n")

    last_publish = 0
    publish_interval = 1.0  # seconds

    while ctrl._running:
        if msvcrt.kbhit():
            ch = msvcrt.getch()
            pressed = None

            if ch in (b"\x00", b"\xe0"):  # Arrow keys
                key = msvcrt.getch()
                if key == b'H':  # Up
                    ctrl.velocity += step_speed
                    pressed = "↑"
                elif key == b'P':  # Down
                    ctrl.velocity = max(0, ctrl.velocity - step_speed)
                    pressed = "↓"
                elif key == b'K':  # Left
                    ctrl.offset -= step_offset
                    pressed = "←"
                elif key == b'M':  # Right
                    ctrl.offset += step_offset
                    pressed = "→"
            else:
                key = ch.decode().lower()
                if key == 'w':
                    ctrl.velocity += step_speed
                    pressed = 'W'
                elif key == 's':
                    ctrl.velocity = max(0, ctrl.velocity - step_speed)
                    pressed = 'S'
                elif key == 'a':
                    ctrl.offset -= step_offset
                    pressed = 'A'
                elif key == 'd':
                    ctrl.offset += step_offset
                    pressed = 'D'
                elif key == ' ':
                    ctrl.stop()
                    pressed = 'Space'
                elif key == 'q':
                    print("[INFO] Quit command received.")
                    ctrl._running = False
                    break

            if pressed:
                print(f"[KEY] {pressed} | Speed={ctrl.velocity} | Offset={ctrl.offset}")

        # Clamp limits
        ctrl.velocity = min(max(ctrl.velocity, 0), 1000)
        ctrl.offset = min(max(ctrl.offset, -100), 100)

        # Simulate publish
        now = time.time()
        if now - last_publish > publish_interval:
            ctrl.publish_speed()
            ctrl.publish_lane()
            last_publish = now

        time.sleep(0.01)


def main():
    ctrl = RemoteControlTest()
    try:
        read_keys_loop(ctrl)
    except KeyboardInterrupt:
        print("[INFO] Interrupted by user.")
    finally:
        ctrl.stop()
        print("[INFO] Exited test mode.")


if __name__ == '__main__':
    main()
