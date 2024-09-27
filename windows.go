package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Data structure to hold network information
type data struct {
	mac               string
	hostname          string
	hostip            string
	urlcust           string
	urlgoogle         string
	packetLossPercent int
	timeTakenMin      float32
	timeTakenAvg      float32
	timeTakenMax      float32
}

// Get the active MAC address
func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			return iface.HardwareAddr.String(), nil
		}
	}
	return "", fmt.Errorf("no active MAC address found")
}

// Get local hostname
func getHostname() (string, error) {
	return os.Hostname()
}

// Get local IP address
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			return ipNet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no valid IP address found")
}

// Ping and return stats
func ping(url string) (int, float32, float32, float32) {
	output, _ := exec.Command("ping", url, "-n", "5").Output()
	reLoss := regexp.MustCompile((\d+)% loss)
	packetLoss := reLoss.FindStringSubmatch(string(output))
	reRTT := regexp.MustCompile(Minimum = (\d+)ms, Maximum = (\d+)ms, Average = (\d+)ms)
	rtt := reRTT.FindStringSubmatch(string(output))

	var loss int
	if len(packetLoss) > 1 {
		fmt.Sscanf(packetLoss[1], "%d", &loss)
	}
	
	var minRTT, maxRTT, avgRTT float32
	if len(rtt) > 3 {
		fmt.Sscanf(rtt[1], "%f", &minRTT)
		fmt.Sscanf(rtt[2], "%f", &maxRTT)
		fmt.Sscanf(rtt[3], "%f", &avgRTT)
	}

	return loss, minRTT, avgRTT, maxRTT
}

// Perform traceroute
func traceRoute(url string) (string, error) {
	output, err := exec.Command("tracert", url).CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// DNS Lookup
func dnsLookup(url string) ([]string, error) {
	ips, err := net.LookupIP(url)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, ip := range ips {
		results = append(results, ip.String())
	}
	return results, nil
}

// Write results to a file
func writeToFile(data data) error {
	file, err := os.Create("network_info.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "MAC: %s\n", data.mac)
	_, err = fmt.Fprintf(file, "Hostname: %s\n", data.hostname)
	_, err = fmt.Fprintf(file, "IP: %s\n", data.hostip)
	_, err = fmt.Fprintf(file, "URL to Ping: %s\n", data.urlcust)
	_, err = fmt.Fprintf(file, "Google URL: %s\n", data.urlgoogle)
	_, err = fmt.Fprintf(file, "Ping Results:\n")
	_, err = fmt.Fprintf(file, "  Loss: %d%%\n", data.packetLossPercent)
	_, err = fmt.Fprintf(file, "  Min RTT: %.2f ms\n", data.timeTakenMin)
	_, err = fmt.Fprintf(file, "  Avg RTT: %.2f ms\n", data.timeTakenAvg)
	_, err = fmt.Fprintf(file, "  Max RTT: %.2f ms\n", data.timeTakenMax)

	return err
}

func main() {
	var netData data

	netData.mac, _ = getMACAddress()
	netData.hostname, _ = getHostname()
	netData.hostip, _ = getLocalIP()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the URL to ping: ")
	url, _ := reader.ReadString('\n')
	netData.urlcust = strings.TrimSpace(url)

	// You can set a Google URL if needed
	netData.urlgoogle = "www.google.com"

	netData.packetLossPercent, netData.timeTakenMin, netData.timeTakenAvg, netData.timeTakenMax = ping(netData.urlcust)
	tracerouteOutput, _ := traceRoute(netData.urlcust)
	dnsResults, _ := dnsLookup(netData.urlcust)

	// Prepare results for output and file
	var results strings.Builder
	results.WriteString(fmt.Sprintf("MAC: %s\n", netData.mac))
	results.WriteString(fmt.Sprintf("Hostname: %s\n", netData.hostname))
	results.WriteString(fmt.Sprintf("IP: %s\n", netData.hostip))
	results.WriteString(fmt.Sprintf("URL to Ping: %s\n", netData.urlcust))
	results.WriteString(fmt.Sprintf("Google URL: %s\n", netData.urlgoogle))
	results.WriteString("Ping Results:\n")
	results.WriteString(fmt.Sprintf("  Loss: %d%%\n", netData.packetLossPercent))
	results.WriteString(fmt.Sprintf("  Min RTT: %.2f ms\n", netData.timeTakenMin))
	results.WriteString(fmt.Sprintf("  Avg RTT: %.2f ms\n", netData.timeTakenAvg))
	results.WriteString(fmt.Sprintf("  Max RTT: %.2f ms\n", netData.timeTakenMax))
	results.WriteString("DNS Lookup Results:\n")
	for _, ip := range dnsResults {
		results.WriteString(fmt.Sprintf("  IP Address: %s\n", ip))
	}
	results.WriteString("Traceroute Results:\n")
	results.WriteString(tracerouteOutput)

	// Write to file
	err := writeToFile(netData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Results saved to network_info.txt.")
}
