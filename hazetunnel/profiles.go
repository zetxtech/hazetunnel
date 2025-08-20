package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mileusna/useragent"
)

// Predefined dictionary with browser versions and their corresponding utls values.
// Updated as of utls v1.8.0.
// Based on: https://github.com/refraction-networking/utls/blob/master/u_common.go
var utlsDict = map[string]map[int]string{
	"Firefox": {
		-1:  "120", // default to latest supported
		55:  "55",
		56:  "56",
		63:  "63",
		65:  "65",
		99:  "99",
		102: "102",
		105: "105",
		120: "120",
	},
	"Chrome": {
		-1:  "133", // default to latest supported in utls v1.8.0
		58:  "58",
		62:  "62",
		70:  "70",
		72:  "72",
		83:  "83",
		87:  "87",
		96:  "96",
		100: "100",
		102: "102",
		106: "106",
		112: "112_PSK",
		114: "114_PSK",
		115: "115_PQ",
		120: "120",
		131: "131",
		133: "133",
	},
	"iOS": {
		-1: "14",  // default to latest supported
		11: "111", // legacy "111" means 11.1
		12: "12.1",
		13: "13",
		14: "14",
	},
	"Android": {
		-1: "11",
	},
	"Edge": {
		-1: "85",
		85: "85",
		// Note: Edge 106 exists in utls but marked as incompatible
	},
	"Safari": {
		-1: "16.0",
	},
	"360Browser": {
		-1: "7.5",
		// Note: 360Browser 11.0 exists in utls but marked as incompatible
	},
	"QQBrowser": {
		-1: "11.1",
	},
}

func uagentToUtls(uagent string) (string, string, error) {
	ua := useragent.Parse(uagent)
	utlsVersion, err := utlsVersion(ua.Name, ua.Version)
	if err != nil {
		return "", "", err
	}
	return ua.Name, utlsVersion, nil
}

func utlsVersion(browserName, browserVersion string) (string, error) {
	if versions, ok := utlsDict[browserName]; ok {
		// Extract the major version number from the browser version string
		majorVersionStr := strings.Split(browserVersion, ".")[0]
		majorVersion, err := strconv.Atoi(majorVersionStr)
		if err != nil {
			return "", fmt.Errorf("error parsing major version number from browser version: %v", err)
		}

		// Find the highest version that is less than or equal to the browser version
		var selectedVersion int = -1
		for version := range versions {
			if version <= majorVersion && version > selectedVersion {
				selectedVersion = version
			}
		}

		if utls, ok := versions[selectedVersion]; ok {
			return utls, nil
		} else {
			return "", fmt.Errorf("no UTLS value found for browser '%s' with version '%s'", browserName, browserVersion)
		}
	}
	return "", fmt.Errorf("browser '%s' not found in UTLS dictionary", browserName)
}

// GetSupportedBrowsers returns a list of all supported browsers
func GetSupportedBrowsers() []string {
	browsers := make([]string, 0, len(utlsDict))
	for browser := range utlsDict {
		browsers = append(browsers, browser)
	}
	return browsers
}

// GetSupportedVersions returns all supported versions for a given browser
func GetSupportedVersions(browserName string) []int {
	if versions, ok := utlsDict[browserName]; ok {
		versionList := make([]int, 0, len(versions))
		for version := range versions {
			if version != -1 { // exclude default version
				versionList = append(versionList, version)
			}
		}
		return versionList
	}
	return nil
}
