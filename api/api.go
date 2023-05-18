package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"
)

var logger *log.Logger

func StartAPI(logger *log.Logger) error {
	// Start the HTTP server
	http.HandleFunc("/IpamDriver.RequestAddress", requestAddressHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Printf("Failed to start the API server: %v", err)
		return err
	}

	return nil
}


func handleResponse(packet dhcp4.Packet) {
	// Extract the IP address from the DHCP response
	ipAddr := packet.YIAddr().String()

	// Extract the subnet mask from the DHCP response
	subnetMask := net.IP(packet.Option(dhcp4.OptionSubnetMask)).String()

	// Extract the default gateway from the DHCP response
	router := net.IP(packet.Option(dhcp4.OptionRouter)).String()

	// Log the extracted information
	logger.Printf("IP Address: %s\n", ipAddr)
	logger.Printf("Subnet Mask: %s\n", subnetMask)
	logger.Printf("Default Gateway: %s\n", router)

	// Forward the extracted information to the Docker IPAM driver
	err := forwardToIPAMDriver(ipAddr, subnetMask, router)
	if err != nil {
		logger.Printf("Failed to forward information to IPAM driver: %v", err)
	}
}

func forwardToIPAMDriver(ipAddr, subnetMask, router string) error {
	// This function would interact with your IPAM driver to forward the
	// extracted information. The implementation would depend on your specific
	// IPAM driver and use case.

	// For example, you might send an HTTP request to the IPAM driver:
	reqBody := map[string]string{
		"ipAddr":     ipAddr,
		"subnetMask": subnetMask,
		"router":     router,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/IpamDriver.RequestAddress", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("IPAM driver returned status code %d", resp.StatusCode)
	}

	return nil
}

func requestAddressHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req struct {
		PoolID  string `json:"PoolID"`
		Address string `json:"Address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse the MAC address
	hwAddr, err := net.ParseMAC(req.Address)
	if err != nil {
		logger.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a connection to listen for DHCP requests
	conn, err := dhcp4client.NewInetSock(dhcp4client.SetLocalAddr(net.UDPAddr{IP: net.IPv4zero, Port: 68}))
	if err != nil {
		logger.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a DHCP client
	client, err := dhcp4client.New(dhcp4client.HardwareAddr(hwAddr), dhcp4client.Connection(conn))
	if err != nil {
		logger.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a DHCP request
	packet, err := client.Request()
	if err != nil {
		logger.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract the IP address from the DHCP response
	allocatedIP := packet.YIAddr().String()

	// Send the response
	var res struct {
		Address string `json:"Address"`
	}
	res.Address = allocatedIP
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Printf("Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
