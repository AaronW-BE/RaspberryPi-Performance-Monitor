# Raspberry Pi Performance Monitor

A lightweight web-based performance monitoring tool for Raspberry Pi devices, built with Go and served using the Fiber framework. This application provides real-time metrics such as CPU usage, memory usage, disk I/O, network activity, temperature, and system uptime, visualized through an interactive HTML interface.

## Features
- Real-time monitoring of CPU, memory, disk, and network usage.
- Detailed per-core CPU usage with responsive charts.
- Temperature monitoring using `vcgencmd`.
- Customizable refresh rate and temperature units via a settings modal.
- Dark/light theme toggle.
- Served as a single binary for easy deployment on Linux ARM64 systems.

## Prerequisites
- Go 1.18 or higher
- Raspberry Pi with a Linux ARM64 environment
- Internet connection for fetching dependencies during build

## Installation

### Dependencies
Install the required Go packages:
```bash
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/fiber/v2/middleware/filesystem
go get github.com/shirou/gopsutil/v3
```

### Build
1. Clone the repository or set up your project directory with the provided `main.go` and `static/` folder (containing `index.html`).
2. Set the target environment for Raspberry Pi (Linux ARM64):
   ```powershell
   $env:GOOS="linux"
   $env:GOARCH="arm64"
   ```
3. Create a directory for the binary output:
   ```bash
   mkdir bin
   ```
4. Build the project:
   ```bash
   go build -o ./bin/monitor main.go
   ```
5. Transfer the `bin/monitor` binary and the `static/` directory to your Raspberry Pi.

### Deployment
- Copy the `monitor` binary and `static/` directory to your Raspberry Pi.
- Make the binary executable:
  ```bash
  chmod +x ./monitor
  ```
- Run the application:
  ```bash
  ./monitor
  ```
- The server will start on `http://0.0.0.0:3000` (or the port specified via the `--port` flag).

## Usage
- Open a web browser and navigate to `http://<raspberry-pi-ip>:3000`.
- The dashboard displays real-time performance metrics.
- Click "Expand/Collapse Core Details" to view per-core CPU usage charts.
- Use the settings modal (gear icon) to adjust the refresh rate, data points, and temperature unit.
- Toggle between light and dark themes using the moon/sun icon.

### Command-Line Options
- Specify a custom port:
  ```bash
  ./monitor --port 8080
  ```
- Use the `PORT` environment variable:
  ```bash
  export PORT=8080
  ./monitor
  ```

## Configuration
- The `static/index.html` file contains the frontend interface. Modify it to customize the layout or add features.
- Metrics are fetched from the `/metrics` endpoint, implemented in `main.go`.

## Contributing
Contributions are welcome! Please fork the repository and submit pull requests with improvements or bug fixes.

## License
This project is licensed under the MIT License. See the `LICENSE` file for details (if applicable).

## Acknowledgements
- Built with [Fiber](https://gofiber.io/) for the web server.
- Utilizes [gopsutil](https://github.com/shirou/gopsutil) for system metrics.
- Frontend powered by Chart.js, Tailwind CSS, and Font Awesome.

## Last Updated
03:43 AM PDT, Wednesday, May 14, 2025