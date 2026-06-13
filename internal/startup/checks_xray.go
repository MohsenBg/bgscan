package startup

import (
	"bgscan/internal/core/process"
	"bgscan/internal/core/xray"
)

// checkXrayHealth verifies that the Xray binary is available, executable,
// and capable of loading outbound configuration templates.
//
// The check performs the following steps:
//  1. Locate the Xray binary.
//  2. Ensure the binary has executable permissions.
//  3. Retrieve and print the Xray version.
//  4. Discover available outbound templates.
//  5. Validate each outbound configuration.
func checkXrayHealth() {
	info("[INFO] Finding Xray binary...")

	path, err := xray.FindXrayBinary()
	if err != nil {
		binaryMissing("Xray", "xray")
		errMsg("Binary lookup error", err)
		return
	}

	successf("[SUCCESS] Xray found at: %s", path)

	info("[INFO] Ensuring Xray binary is executable...")
	if err := process.EnsureExecutable(path); err != nil {
		errMsg("Failed to set executable bit for Xray binary", err)
		return
	}
	success("[SUCCESS] Xray binary is executable")

	info("[INFO] Checking Xray version...")

	version, err := xray.XrayVersion()
	if err != nil {
		errMsg("Failed to retrieve Xray version", err)
		return
	}

	successf("[SUCCESS] Xray version: %s", version)

	info("[INFO] Searching for configuration templates...")

	outbounds, err := xray.ListOutboundTemplates()
	if err != nil {
		errMsg("Failed to retrieve outbounds", err)
		return
	}

	infof("[INFO] Found %d outbound templates.", len(outbounds))

	for _, outbound := range outbounds {
		infof("[INFO] Validating outbound: %s", outbound.Name)

		if err := xray.ValidateOutbound(outbound.Name); err != nil {
			errMsg("Outbound validation failed: "+outbound.Name, err)
			continue
		}

		successf("[VALID] %s OK", outbound.Name)
	}

	success("[XRAY] Health check completed successfully ✅")
}
