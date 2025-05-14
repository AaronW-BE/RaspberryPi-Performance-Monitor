package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Embed static files from the "static" directory into the binary
//
//go:embed static/*
var staticFiles embed.FS

// Metrics struct defines the structure for system performance metrics
type Metrics struct {
	CPUPercent      []float64 `json:"cpu_percent"`      // CPU usage percentage for each core
	CPUOverall      float64   `json:"cpu_overall"`      // Overall CPU usage percentage
	MemoryTotal     uint64    `json:"memory_total"`     // Total memory (MB)
	MemoryUsed      uint64    `json:"memory_used"`      // Used memory (MB)
	MemoryAvailable uint64    `json:"memory_available"` // Available memory (MB)
	MemoryPercent   float64   `json:"memory_percent"`   // Memory usage percentage
	DiskTotal       uint64    `json:"disk_total"`       // Total disk space (GB)
	DiskUsed        uint64    `json:"disk_used"`        // Used disk space (GB)
	DiskFree        uint64    `json:"disk_free"`        // Free disk space (GB)
	DiskPercent     float64   `json:"disk_percent"`     // Disk usage percentage
	DiskReadBytes   uint64    `json:"disk_read_bytes"`  // Disk read rate (KB/s)
	DiskWriteBytes  uint64    `json:"disk_write_bytes"` // Disk write rate (KB/s)
	NetSentBytes    uint64    `json:"net_sent_bytes"`   // Network sent rate (KB/s)
	NetRecvBytes    uint64    `json:"net_recv_bytes"`   // Network received rate (KB/s)
	Temperature     float64   `json:"temperature"`      // CPU temperature (°C)
	Uptime          uint64    `json:"uptime"`           // System uptime (seconds)
	Timestamp       int64     `json:"timestamp"`        // Current timestamp
}

// getTemperature retrieves the CPU temperature using the vcgencmd command
func getTemperature() float64 {
	// Execute the command to measure temperature
	out, err := exec.Command("vcgencmd", "measure_temp").Output()
	if err != nil {
		return 0.0 // Return 0.0 if the command fails
	}
	// Parse the temperature value from the output (e.g., "temp=45.0'C")
	tempStr := strings.Split(string(out), "=")[1]
	temp, _ := strconv.ParseFloat(strings.Split(tempStr, "'")[0], 64)
	return temp
}

// getMetrics collects system performance metrics
func getMetrics() Metrics {
	// CPU Usage
	// Get CPU usage percentage for each core over a 1-second interval
	cpuPercents, _ := cpu.Percent(time.Second, true)
	var cpuOverall float64
	// Calculate the overall CPU usage by averaging the per-core percentages
	for _, p := range cpuPercents {
		cpuOverall += p
	}
	if len(cpuPercents) > 0 {
		cpuOverall /= float64(len(cpuPercents))
	}

	// Memory
	// Retrieve virtual memory statistics
	vm, _ := mem.VirtualMemory()
	memoryTotal := vm.Total / 1024 / 1024         // Convert to MB
	memoryUsed := vm.Used / 1024 / 1024           // Convert to MB
	memoryAvailable := vm.Available / 1024 / 1024 // Convert to MB

	// Disk
	// Get disk usage statistics for the root directory
	diskInfo, _ := disk.Usage("/")
	diskTotal := diskInfo.Total / 1024 / 1024 / 1024 // Convert to GB
	diskUsed := diskInfo.Used / 1024 / 1024 / 1024   // Convert to GB
	diskFree := diskInfo.Free / 1024 / 1024 / 1024   // Convert to GB

	// Disk I/O
	// Retrieve disk I/O statistics
	diskIO, _ := disk.IOCounters()
	var readBytes, writeBytes uint64
	// Sum up read and write bytes across all disks
	for _, io := range diskIO {
		readBytes += io.ReadBytes / 1024   // Convert to KB
		writeBytes += io.WriteBytes / 1024 // Convert to KB
	}

	// Network I/O
	// Retrieve network I/O statistics (not per-interface)
	netIO, _ := net.IOCounters(false)
	var sentBytes, recvBytes uint64
	// Sum up sent and received bytes across all interfaces
	for _, n := range netIO {
		sentBytes += n.BytesSent / 1024 // Convert to KB
		recvBytes += n.BytesRecv / 1024 // Convert to KB
	}

	// System Uptime
	// Get the system uptime in seconds
	uptime, _ := host.Uptime()

	// Return the collected metrics
	return Metrics{
		CPUPercent:      cpuPercents,
		CPUOverall:      cpuOverall,
		MemoryTotal:     memoryTotal,
		MemoryUsed:      memoryUsed,
		MemoryAvailable: memoryAvailable,
		MemoryPercent:   vm.UsedPercent,
		DiskTotal:       diskTotal,
		DiskUsed:        diskUsed,
		DiskFree:        diskFree,
		DiskPercent:     diskInfo.UsedPercent,
		DiskReadBytes:   readBytes,
		DiskWriteBytes:  writeBytes,
		NetSentBytes:    sentBytes,
		NetRecvBytes:    recvBytes,
		Temperature:     getTemperature(),
		Uptime:          uptime,
		Timestamp:       time.Now().Unix(),
	}
}

func main() {
	// Define command-line flags
	// Port to listen on, defaulting to 3000
	portPtr := flag.Int("port", 3000, "Port to listen on (default 3000)")
	flag.Parse()

	// Check for PORT environment variable
	// Override the port if a valid PORT environment variable is set
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			*portPtr = p
		}
	}

	// Validate port number
	// Ensure the port is within the valid range (1-65535)
	if *portPtr <= 0 || *portPtr > 65535 {
		fmt.Fprintf(os.Stderr, "Invalid port number: %d. Using default port 3000.\n", *portPtr)
		*portPtr = 3000
	}

	// Initialize Fiber app
	app := fiber.New()

	// Serve embedded static files
	// Create a sub-filesystem for the "static" directory within the embedded files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to access embedded static files: %v\n", err)
		os.Exit(1)
	}
	// Configure Fiber to serve static files
	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.FS(staticFS),
		Index:  "index.html",
		Browse: false,
	}))

	// Metrics endpoint
	// Responds with system performance metrics in JSON format
	app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(getMetrics())
	})

	// Start the server
	addr := fmt.Sprintf(":%d", *portPtr)
	fmt.Printf("Server starting on http://0.0.0.0:%d\n", *portPtr)
	// Listen on the specified port
	if err := app.Listen(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
