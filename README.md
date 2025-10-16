ENG

Ping Logger is a simple command-line tool written in Go for monitoring network latency by pinging one or multiple servers. It supports ICMP pings (requires admin privileges on Windows), logs results to CSV files, displays statistics, and generates ASCII graphs of RTT (Round-Trip Time) values.

FEATURES
Input Modes:
- Single host via manual input.
- Multiple hosts via comma-separated input (e.g., 8.8.8.8, google.com).
- Hosts from a targetservers.txt file (one per line, optional interval in minutes).
- Customizable Ping Interval: User-defined in seconds or minutes (e.g., 5 sec, 2 min, or just 10 for 10 seconds). Default: 1 second. Minimum: 1 second.
- Parallel Pinging: Pings multiple hosts concurrently in goroutines.
- Logging: Saves ping results (timestamp + RTT or "loss") to CSV files in a logs/ folder (auto-created).
- Statistics: Average RTT, packet loss count and percentage.
- ASCII Graph: Visualizes RTT over time using asciigraph. Handles stable/flat lines and losses gracefully.
- Graceful Stop: Press Enter or Ctrl+C to stop all pings.
- Validation: Checks IP/domain format; strips http/https prefixes.
- Fallbacks: If file is missing/empty or input invalid, switches to manual mode.

REQUIREMENTS
- Go 1.16+ installed.
Run as Administrator on Windows (for privileged ICMP).

Dependencies:
- github.com/go-ping/ping
- github.com/guptarohit/asciigraph (auto-fetched via go mod tidy).

INSTALLATION
- Clone or download the project.
- Navigate to the project directory: cd C:\Go\ping_logger
- Fetch dependencies: go mod tidy

USAGE
- Run the program (as admin-optional)in power shell:
- go run main.go

Choose mode (1-3):
1: Enter one host, then interval (e.g., 5 sec).
2: Enter hosts separated by ", " (comma + space), then interval per host.
3: Uses targetservers.txt (format: host [minutes] per line, e.g., 8.8.8.8 2 for 2 minutes interval).


- Pinging starts with live output (e.g., Ping 1: 10 ms or loss).
- Press Enter to stop.
Outputs:
- Stats, graph in console; CSV in logs/ (e.g., logs/ping_log_8_8_8_8.csv).

EXAMPLE targetservers.txt
- 8.8.8.8 0.5  # Ping every 30 seconds
- google.com 1 # Every 1 minute
- Comments ignored after #

PROJECT STRUCTURE
- main.go: Entry point and UI logic.
- pinger/pinger.go: Ping logic using go-ping.
- utils/utils.go: Helpers for input, validation, stats.
- graphplotter/plotter.go: ASCII graph generation.
- logs/: Auto-generated for CSV logs.
- go.mod / go.sum: Module dependencies.

LIMITATIONS
- ICMP requires elevated privileges.
- No GUI; console-only.
- Graphs are ASCII (text-based); for images, extend with external libs.
- File intervals in minutes only (for simplicity).

CONTRIBUTING
Feel free to fork and submit PRs for improvements like HTML reports or more units (hours).
License
MIT LICENSE 
(feel free to modify).
