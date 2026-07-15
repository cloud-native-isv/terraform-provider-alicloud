package flinkworkspace

import "fmt"

type VSwitch struct {
	ID       string
	RegionID string
	VPCID    string
	ZoneID   string
}

type VSwitchTopology struct {
	PrimaryZoneID string
	StandbyZoneID string
}

func ValidateVSwitchTopology(regionID, vpcID, legacyPrimaryZoneID, legacyStandbyZoneID string, primary, standby []VSwitch) (VSwitchTopology, error) {
	primaryZoneID, err := validateVSwitchGroup("primary", regionID, vpcID, primary)
	if err != nil {
		return VSwitchTopology{}, err
	}
	standbyZoneID := ""
	if len(standby) > 0 {
		standbyZoneID, err = validateVSwitchGroup("standby", regionID, vpcID, standby)
		if err != nil {
			return VSwitchTopology{}, err
		}
		if standbyZoneID == primaryZoneID {
			return VSwitchTopology{}, fmt.Errorf("primary and standby vSwitch groups must belong to different zones, both resolved to %q", primaryZoneID)
		}
	}

	if legacyPrimaryZoneID != "" && legacyPrimaryZoneID != primaryZoneID {
		return VSwitchTopology{}, fmt.Errorf("legacy primary zone %q does not match vSwitch-derived zone %q", legacyPrimaryZoneID, primaryZoneID)
	}
	if legacyStandbyZoneID != "" {
		if standbyZoneID == "" {
			return VSwitchTopology{}, fmt.Errorf("legacy standby zone %q was configured without a standby vSwitch group", legacyStandbyZoneID)
		}
		if legacyStandbyZoneID != standbyZoneID {
			return VSwitchTopology{}, fmt.Errorf("legacy standby zone %q does not match vSwitch-derived zone %q", legacyStandbyZoneID, standbyZoneID)
		}
	}
	return VSwitchTopology{PrimaryZoneID: primaryZoneID, StandbyZoneID: standbyZoneID}, nil
}

func validateVSwitchGroup(name, regionID, vpcID string, vSwitches []VSwitch) (string, error) {
	if len(vSwitches) == 0 {
		return "", fmt.Errorf("%s vSwitch group must contain at least one vSwitch", name)
	}
	zoneID := ""
	for _, vSwitch := range vSwitches {
		if vSwitch.ID == "" || vSwitch.ZoneID == "" || vSwitch.VPCID == "" {
			return "", fmt.Errorf("%s vSwitch group contains incomplete metadata for %q", name, vSwitch.ID)
		}
		if vSwitch.RegionID != "" && vSwitch.RegionID != regionID {
			return "", fmt.Errorf("%s vSwitch %q belongs to region %q, want %q", name, vSwitch.ID, vSwitch.RegionID, regionID)
		}
		if vSwitch.VPCID != vpcID {
			return "", fmt.Errorf("%s vSwitch %q belongs to VPC %q, want %q", name, vSwitch.ID, vSwitch.VPCID, vpcID)
		}
		if zoneID == "" {
			zoneID = vSwitch.ZoneID
		} else if vSwitch.ZoneID != zoneID {
			return "", fmt.Errorf("all vSwitches in the %s group must belong to the same zone; got %q and %q", name, zoneID, vSwitch.ZoneID)
		}
	}
	return zoneID, nil
}
