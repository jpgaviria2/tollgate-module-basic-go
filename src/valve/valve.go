package valve

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Module-level logger with pre-configured module field
var logger = logrus.WithField("module", "valve")

// openGates keeps track of MAC addresses that have been authorized
var (
	openGates  = make(map[string]*time.Timer)
	gatesMutex = &sync.Mutex{}
	// Bandwidth limits for different tiers (in kbps)
	bandwidthLimits = map[string]int{
		"free":    2048, // 2Mbps for free tier
		"premium": 0,    // 0 = unlimited for premium
		"staff":   0,    // 0 = unlimited for staff
	}
)

// setBandwidthLimit applies traffic control rules to limit bandwidth for a MAC address
func setBandwidthLimit(macAddress string, tier string) error {
	limit, exists := bandwidthLimits[tier]
	if !exists {
		return fmt.Errorf("unknown tier: %s", tier)
	}

	// If limit is 0, remove any existing limits (unlimited)
	if limit == 0 {
		return removeBandwidthLimit(macAddress)
	}

	// Apply bandwidth limit using tc (traffic control)
	// This requires the interface to be configured with HTB qdisc
	cmd := exec.Command("tc", "class", "add", "dev", "br-lan", "parent", "1:1", "classid", "1:"+getClassID(macAddress),
		"htb", "rate", strconv.Itoa(limit)+"kbit", "ceil", strconv.Itoa(limit)+"kbit")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
			"tier":        tier,
			"limit":       limit,
			"error":       err,
			"output":      string(output),
		}).Warn("Failed to set bandwidth limit, may already exist or tc not configured")
		// Don't return error - some systems may not have tc configured
	}

	// Add filter to match the MAC address to the class
	filterCmd := exec.Command("tc", "filter", "add", "dev", "br-lan", "protocol", "ip", "parent", "1:0",
		"prio", "1", "u32", "match", "u16", "0x0800", "0xFFFF", "at", "-2",
		"match", "u32", "0x"+strings.Replace(macAddress, ":", "", -1), "0xFFFFFFFF", "at", "-12",
		"flowid", "1:"+getClassID(macAddress))
	filterOutput, filterErr := filterCmd.CombinedOutput()
	if filterErr != nil {
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
			"tier":        tier,
			"error":       filterErr,
			"output":      string(filterOutput),
		}).Warn("Failed to add tc filter")
	}

	logger.WithFields(logrus.Fields{
		"mac_address": macAddress,
		"tier":        tier,
		"limit_kbps":  limit,
	}).Info("Applied bandwidth limit")

	return nil
}

// removeBandwidthLimit removes traffic control rules for a MAC address
func removeBandwidthLimit(macAddress string) error {
	classID := getClassID(macAddress)

	// Remove filter first
	filterCmd := exec.Command("tc", "filter", "del", "dev", "br-lan", "protocol", "ip", "parent", "1:0",
		"prio", "1", "u32", "match", "u16", "0x0800", "0xFFFF", "at", "-2",
		"match", "u32", "0x"+strings.Replace(macAddress, ":", "", -1), "0xFFFFFFFF", "at", "-12",
		"flowid", "1:"+classID)
	filterCmd.Run() // Ignore errors, filter may not exist

	// Remove class
	classCmd := exec.Command("tc", "class", "del", "dev", "br-lan", "classid", "1:"+classID)
	classCmd.Run() // Ignore errors, class may not exist

	logger.WithFields(logrus.Fields{
		"mac_address": macAddress,
	}).Info("Removed bandwidth limit")

	return nil
}

// getClassID generates a unique class ID for a MAC address
func getClassID(macAddress string) string {
	// Convert last 2 bytes of MAC to a number for class ID
	parts := strings.Split(macAddress, ":")
	if len(parts) >= 2 {
		// Use last byte as minor class ID (2-255)
		minor := parts[len(parts)-1]
		if val, err := strconv.ParseInt(minor, 16, 64); err == nil {
			if val < 2 {
				val = 2 // Reserve 1 for root class
			}
			return fmt.Sprintf("%d", val)
		}
	}
	return "2" // Fallback
}

// initTrafficControl initializes the traffic control qdisc on the bridge interface
// This must be called before applying bandwidth limits
func initTrafficControl() error {
	// Check if HTB qdisc is already set up
	cmd := exec.Command("tc", "qdisc", "show", "dev", "br-lan")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check tc qdisc: %w", err)
	}

	// If HTB is already configured, don't reconfigure
	if strings.Contains(string(output), "htb") {
		logger.Debug("Traffic control already initialized on br-lan")
		return nil
	}

	// Remove any existing qdisc
	delCmd := exec.Command("tc", "qdisc", "del", "dev", "br-lan", "root")
	delCmd.Run() // Ignore errors, may not exist

	// Add HTB qdisc
	addCmd := exec.Command("tc", "qdisc", "add", "dev", "br-lan", "root", "handle", "1:", "htb", "default", "1")
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add HTB qdisc: %w (output: %s)", err, string(output))
	}

	// Add root class with unlimited bandwidth
	rootCmd := exec.Command("tc", "class", "add", "dev", "br-lan", "parent", "1:", "classid", "1:1", "htb", "rate", "1000mbit", "ceil", "1000mbit")
	if output, err := rootCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add root class: %w (output: %s)", err, string(output))
	}

	logger.Info("Initialized traffic control on br-lan interface")
	return nil
}

// InitTrafficControl initializes traffic control on the bridge interface
func InitTrafficControl() error {
	return initTrafficControl()
}

// authorizeMAC authorizes a MAC address using ndsctl and applies bandwidth limits
func authorizeMAC(macAddress string, tier string) error {
	cmd := exec.Command("ndsctl", "auth", macAddress)
	output, err := cmd.Output()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
			"tier":        tier,
			"error":       err,
		}).Error("Error authorizing MAC address")
		return err
	}

	logger.WithFields(logrus.Fields{
		"mac_address": macAddress,
		"tier":        tier,
		"output":      string(output),
	}).Info("Authorization successful for MAC")

	// Apply bandwidth limiting based on tier
	if err := setBandwidthLimit(macAddress, tier); err != nil {
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
			"tier":        tier,
			"error":       err,
		}).Warn("Failed to apply bandwidth limit, but authorization succeeded")
	}

	return nil
}

// deauthorizeMAC deauthorizes a MAC address using ndsctl and removes bandwidth limits
func deauthorizeMAC(macAddress string) error {
	cmd := exec.Command("ndsctl", "deauth", macAddress)
	output, err := cmd.Output()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
			"error":       err,
		}).Error("Error deauthorizing MAC address")
		return err
	}

	logger.WithFields(logrus.Fields{
		"mac_address": macAddress,
		"output":      string(output),
	}).Debug("Deauthorization successful for MAC")

	// Remove bandwidth limiting
	if err := removeBandwidthLimit(macAddress); err != nil {
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
			"error":       err,
		}).Warn("Failed to remove bandwidth limit, but deauthorization succeeded")
	}

	return nil
}

// OpenGateUntil opens the gate (if not opened yet) and sets a timer until the timestamp.
// If there is already a timer running, it will extend the timer.
func OpenGateUntil(macAddress string, untilTimestamp int64, tier string) error {
	now := time.Now().Unix()

	// Calculate duration until the target timestamp
	durationSeconds := untilTimestamp - now

	// If the timestamp is in the past, return an error
	if durationSeconds <= 0 {
		return fmt.Errorf("timestamp %d is in the past (current time: %d)", untilTimestamp, now)
	}

	logger.WithFields(logrus.Fields{
		"mac_address":      macAddress,
		"until_timestamp":  untilTimestamp,
		"duration_seconds": durationSeconds,
	}).Info("Opening gate until timestamp")

	gatesMutex.Lock()
	defer gatesMutex.Unlock()

	// Check if the MAC is already in openGates
	existingTimer, exists := openGates[macAddress]

	if !exists {
		// MAC not in openGates, authorize it
		err := authorizeMAC(macAddress, tier)
		if err != nil {
			return fmt.Errorf("error authorizing MAC: %w", err)
		}
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
		}).Debug("New authorization for MAC")
	} else {
		// MAC already in openGates, stop the existing timer
		if existingTimer != nil {
			existingTimer.Stop()
		}
		logger.WithFields(logrus.Fields{
			"mac_address": macAddress,
		}).Debug("Extending access for already authorized MAC")
	}

	// Create a new timer that will call deauthorizeMAC when it expires
	duration := time.Duration(durationSeconds) * time.Second
	timer := time.AfterFunc(duration, func() {
		err := deauthorizeMAC(macAddress)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"mac_address": macAddress,
				"error":       err,
			}).Error("Error deauthorizing MAC after timeout")
		} else {
			logger.WithFields(logrus.Fields{
				"mac_address": macAddress,
			}).Debug("Successfully deauthorized MAC after timeout")
		}

		// Remove the MAC from openGates once timer expires
		gatesMutex.Lock()
		delete(openGates, macAddress)
		gatesMutex.Unlock()
	})

	// Store the timer in openGates
	openGates[macAddress] = timer

	return nil
}
