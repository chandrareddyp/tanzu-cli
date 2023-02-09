// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package plugininventory implements an interface to deal with a plugin inventory.
// It encapsulates the logic that deals with how plugin inventories are stored
// so that other entities can use the plugin inventory without knowing its
// implementation details.
package plugininventory

import (
	"github.com/vmware-tanzu/tanzu-cli/pkg/distribution"
	configtypes "github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
)

// PluginInventoryEntry represents the inventory information
// about a single plugin as found by the inventory backend.
type PluginInventoryEntry struct {
	// Name of the plugin
	Name string
	// Target to which the plugin applies
	Target configtypes.Target
	// Description of the plugin
	Description string
	// Publisher is the name of the publisher of this plugin
	// (e.g., a product group within a company)
	Publisher string
	// Vendor is the name of the vendor of this plugin (e.g., a company's name)
	Vendor string
	// Recommended version that Tanzu CLI should install by default.
	// The value should be a valid semantic version as defined in
	// https://semver.org/. E.g., 2.0.1
	RecommendedVersion string
	// AvailableVersions is the list of versions available for this plugin.
	// The values are sorted in the semver prescribed order as defined in
	// https://github.com/Masterminds/semver#sorting-semantic-versions.
	AvailableVersions []string
	// Artifacts contains an artifact list for every available version.
	Artifacts distribution.Artifacts
}

// PluginInventory is the interface to interact with a plugin inventory.
// It can be used to get the plugin information for all plugins in the
// inventory.
type PluginInventory interface {
	GetAllPlugins() ([]*PluginInventoryEntry, error)
}