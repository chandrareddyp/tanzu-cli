// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugininventory

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	// Import the sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/vmware-tanzu/tanzu-cli/pkg/cli"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Inventory Suite")
}

const createTableStmt = `
CREATE TABLE IF NOT EXISTS "PluginBinaries" (
	"PluginName"         TEXT NOT NULL,
	"Target"             TEXT NOT NULL,
	"RecommendedVersion" TEXT NOT NULL,
	"Version"            TEXT NOT NULL,
	"Hidden"             INTEGER NOT NULL,
	"Description"        TEXT NOT NULL,
	"Publisher"          TEXT NOT NULL,
	"Vendor"             TEXT NOT NULL,
	"OS"                 TEXT NOT NULL,
	"Architecture"       TEXT NOT NULL,
	"Digest"             TEXT NOT NULL,
	"URI"                TEXT NOT NULL,
	PRIMARY KEY("PluginName", "Target", "Version", "OS", "Architecture")
);`

const createPluginsStmt = `
INSERT INTO PluginBinaries VALUES(
	'management-cluster',
	'k8s',
	'v0.28.0',
	'v0.28.0',
	'false',
	'Kubernetes management cluster operations',
	'tkg',
	'vmware',
	'linux',
	'amd64',
	'0000000000',
	'vmware/tkg/linux/amd64/k8s/management-cluster:v0.28.0');
INSERT INTO PluginBinaries VALUES(
	'management-cluster',
	'k8s',
	'v0.28.0',
	'v0.28.0',
	'false',
	'Kubernetes management cluster operations',
	'tkg',
	'vmware',
	'darwin',
	'amd64',
	'1111111111',
	'vmware/tkg/darwin/amd64/k8s/management-cluster:v0.28.0');
INSERT INTO PluginBinaries VALUES(
	'management-cluster',
	'k8s',
	'v0.28.0',
	'v0.26.0',
	'false',
	'Kubernetes management cluster operations',
	'tkg',
	'vmware',
	'windows',
	'amd64',
	'2222222222',
	'vmware/tkg/windows/amd64/k8s/management-cluster:v0.26.0');
INSERT INTO PluginBinaries VALUES(
	'isolated-cluster',
	'global',
	'v1.2.3',
	'v1.2.3',
	'false',
	'Isolated cluster plugin',
	'otherpublisher',
	'othervendor',
	'linux',
	'amd64',
	'3333333333',
	'othervendor/otherpublisher/linux/amd64/global/isolated-cluster:v1.2.3');
`
const createPluginTMCNoRecommendedVersionStmt = `
INSERT INTO PluginBinaries VALUES(
	'management-cluster',
	'tmc',
	'',
	'v0.0.1',
	'false',
	'Mission-control management cluster operations',
	'tmc',
	'vmware',
	'linux',
	'amd64',
	'0000000000',
	'vmware/tmc/linux/amd64/tmc/management-cluster:v0.0.1');
INSERT INTO PluginBinaries VALUES(
	'management-cluster',
	'tmc',
	'',
	'v0.0.2',
	'false',
	'Mission-control management cluster operations',
	'tmc',
	'vmware',
	'linux',
	'amd64',
	'1111111111',
	'vmware/tmc/linux/amd64/tmc/management-cluster:v0.0.2');
`

var _ = Describe("Unit tests for plugin inventory", func() {
	var (
		err       error
		inventory PluginInventory
		dbFile    *os.File
		tmpDir    string
		plugins   []*PluginInventoryEntry
	)

	Describe("Getting plugins from inventory", func() {
		Context("With an empty DB file", func() {
			BeforeEach(func() {
				tmpDir, err = os.MkdirTemp(os.TempDir(), "")
				Expect(err).To(BeNil(), "unable to create temporary directory")

				// Create empty file for the DB
				dbFile, err = os.Create(filepath.Join(tmpDir, sqliteDBFileName))
				Expect(err).To(BeNil())

				inventory = NewSQLiteInventory(tmpDir, tmpDir)
			})
			AfterEach(func() {
				os.RemoveAll(tmpDir)
			})
			It("should return an error", func() {
				_, err = inventory.GetAllPlugins()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to setup DB"))
			})
		})
		Context("With an empty DB table", func() {
			BeforeEach(func() {
				tmpDir, err = os.MkdirTemp(os.TempDir(), "")
				Expect(err).To(BeNil(), "unable to create temporary directory")

				// Create DB file
				dbFile, err = os.Create(filepath.Join(tmpDir, sqliteDBFileName))
				Expect(err).To(BeNil())
				// Open DB with the sqlite driver
				db, err := sql.Open("sqlite3", dbFile.Name())
				Expect(err).To(BeNil(), "failed to open the DB for testing")
				defer db.Close()

				// Create the table but don't add any rows
				_, err = db.Exec(createTableStmt)
				Expect(err).To(BeNil(), "failed to create DB table for testing")

				inventory = NewSQLiteInventory(tmpDir, tmpDir)
			})
			AfterEach(func() {
				os.RemoveAll(tmpDir)
			})
			It("should return an empty list of plugins with no error", func() {
				plugins, err = inventory.GetAllPlugins()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(plugins)).To(Equal(0))
			})
		})
		Describe("With a DB table with two plugins", func() {
			BeforeEach(func() {
				tmpDir, err = os.MkdirTemp(os.TempDir(), "")
				Expect(err).To(BeNil(), "unable to create temporary directory")

				// Create DB file
				dbFile, err = os.Create(filepath.Join(tmpDir, sqliteDBFileName))
				Expect(err).To(BeNil())
				// Open DB with the sqlite driver
				db, err := sql.Open("sqlite3", dbFile.Name())
				Expect(err).To(BeNil(), "failed to open the DB for testing")
				defer db.Close()

				// Create the table
				_, err = db.Exec(createTableStmt)
				Expect(err).To(BeNil(), "failed to create DB table for testing")

				// Add a plugin entry to the DB
				_, err = db.Exec(createPluginsStmt)
				Expect(err).To(BeNil(), "failed to create plugin for testing")

				inventory = NewSQLiteInventory(tmpDir, tmpDir)
			})
			AfterEach(func() {
				os.RemoveAll(tmpDir)
			})
			Context("When getting all plugins", func() {
				It("should return a list of two plugins with no error", func() {
					plugins, err = inventory.GetAllPlugins()
					Expect(err).ToNot(HaveOccurred())
					Expect(len(plugins)).To(Equal(2))

					for _, p := range plugins {
						if p.Name == "management-cluster" {
							Expect(p.RecommendedVersion).To(Equal("v0.28.0"))
							Expect(string(p.Target)).To(Equal("kubernetes"))
							Expect(p.Description).To(Equal("Kubernetes management cluster operations"))
							Expect(p.Vendor).To(Equal("vmware"))
							Expect(p.Publisher).To(Equal("tkg"))

							Expect(len(p.Artifacts)).To(Equal(2))
							artifactList := p.Artifacts["v0.28.0"]
							Expect(len(artifactList)).To(Equal(2))
							for _, a := range artifactList {
								if a.OS == "linux" {
									Expect(a.Arch).To(Equal("amd64"))
									Expect(a.Digest).To(Equal("0000000000"))
									Expect(a.Image).To(Equal(tmpDir + "/vmware/tkg/linux/amd64/k8s/management-cluster:v0.28.0"))
								} else {
									Expect(a.OS).To(Equal("darwin"))
									Expect(a.Arch).To(Equal("amd64"))
									Expect(a.Digest).To(Equal("1111111111"))
									Expect(a.Image).To(Equal(tmpDir + "/vmware/tkg/darwin/amd64/k8s/management-cluster:v0.28.0"))
								}
							}

							artifactList = p.Artifacts["v0.26.0"]
							Expect(len(artifactList)).To(Equal(1))

							Expect(artifactList[0].OS).To(Equal("windows"))
							Expect(artifactList[0].Arch).To(Equal("amd64"))
							Expect(artifactList[0].Digest).To(Equal("2222222222"))
							Expect(artifactList[0].Image).To(Equal(tmpDir + "/vmware/tkg/windows/amd64/k8s/management-cluster:v0.26.0"))
						} else {
							Expect(p.Name).To(Equal("isolated-cluster"))

							Expect(p.RecommendedVersion).To(Equal("v1.2.3"))
							Expect(string(p.Target)).To(Equal(""))
							Expect(p.Description).To(Equal("Isolated cluster plugin"))
							Expect(p.Vendor).To(Equal("othervendor"))
							Expect(p.Publisher).To(Equal("otherpublisher"))

							Expect(len(p.Artifacts)).To(Equal(1))
							a := p.Artifacts["v1.2.3"]
							Expect(len(a)).To(Equal(1))

							Expect(a[0].OS).To(Equal("linux"))
							Expect(a[0].Arch).To(Equal("amd64"))
							Expect(a[0].Digest).To(Equal("3333333333"))
							Expect(a[0].Image).To(Equal(tmpDir + "/othervendor/otherpublisher/linux/amd64/global/isolated-cluster:v1.2.3"))
						}
					}
				})
			})
			Context("When getting a specific plugin version for k8s for an os/arch", func() {
				It("should return a list of one plugin with no error", func() {
					plugins, err = inventory.GetPlugins(&PluginInventoryFilter{
						Name:    "management-cluster",
						Target:  "kubernetes",
						Version: "v0.26.0",
						OS:      "windows",
						Arch:    "amd64",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(len(plugins)).To(Equal(1))

					p := plugins[0]
					Expect(p.Name).To(Equal("management-cluster"))

					Expect(p.RecommendedVersion).To(Equal("v0.28.0"))
					Expect(string(p.Target)).To(Equal("kubernetes"))
					Expect(p.Description).To(Equal("Kubernetes management cluster operations"))
					Expect(p.Vendor).To(Equal("vmware"))
					Expect(p.Publisher).To(Equal("tkg"))

					Expect(len(p.Artifacts)).To(Equal(1))
					artifactList := p.Artifacts["v0.26.0"]
					Expect(len(artifactList)).To(Equal(1))
					a := artifactList[0]
					Expect(a.OS).To(Equal("windows"))
					Expect(a.Arch).To(Equal("amd64"))
					Expect(a.Digest).To(Equal("2222222222"))
					Expect(a.Image).To(Equal(tmpDir + "/vmware/tkg/windows/amd64/k8s/management-cluster:v0.26.0"))
				})
			})
			Context("When getting the recommended version of a plugin for an os/arch", func() {
				It("should return a list of one plugin with no error", func() {
					plugins, err = inventory.GetPlugins(&PluginInventoryFilter{
						Name:    "management-cluster",
						Target:  "kubernetes",
						Version: cli.VersionLatest,
						OS:      "darwin",
						Arch:    "amd64",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(len(plugins)).To(Equal(1))

					p := plugins[0]
					Expect(p.Name).To(Equal("management-cluster"))

					Expect(p.RecommendedVersion).To(Equal("v0.28.0"))
					Expect(string(p.Target)).To(Equal("kubernetes"))
					Expect(p.Description).To(Equal("Kubernetes management cluster operations"))
					Expect(p.Vendor).To(Equal("vmware"))
					Expect(p.Publisher).To(Equal("tkg"))

					Expect(len(p.Artifacts)).To(Equal(1))
					artifactList := p.Artifacts["v0.28.0"]
					Expect(len(artifactList)).To(Equal(1))
					a := artifactList[0]
					Expect(a.OS).To(Equal("darwin"))
					Expect(a.Arch).To(Equal("amd64"))
					Expect(a.Digest).To(Equal("1111111111"))
					Expect(a.Image).To(Equal(tmpDir + "/vmware/tkg/darwin/amd64/k8s/management-cluster:v0.28.0"))
				})
			})
			Context("When getting plugins by vendor", func() {
				It("should return a list of one plugin with no error", func() {
					plugins, err = inventory.GetPlugins(&PluginInventoryFilter{
						Vendor: "vmware",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(len(plugins)).To(Equal(1))

					p := plugins[0]
					Expect(p.Name).To(Equal("management-cluster"))
					Expect(len(p.Artifacts)).To(Equal(2))
					Expect(p.Artifacts["v0.26.0"]).ToNot(BeNil())
					Expect(p.Artifacts["v0.28.0"]).ToNot(BeNil())

					Expect(p.RecommendedVersion).To(Equal("v0.28.0"))
					Expect(string(p.Target)).To(Equal("kubernetes"))
					Expect(p.Description).To(Equal("Kubernetes management cluster operations"))
					Expect(p.Vendor).To(Equal("vmware"))
					Expect(p.Publisher).To(Equal("tkg"))
				})
			})
			Context("When getting plugins by publisher", func() {
				It("should return a list of one plugin with no error", func() {
					plugins, err = inventory.GetPlugins(&PluginInventoryFilter{
						Publisher: "otherpublisher",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(len(plugins)).To(Equal(1))

					p := plugins[0]
					Expect(p.Name).To(Equal("isolated-cluster"))
					Expect(len(p.Artifacts)).To(Equal(1))
					Expect(p.Artifacts["v1.2.3"]).ToNot(BeNil())

					Expect(p.RecommendedVersion).To(Equal("v1.2.3"))
					Expect(string(p.Target)).To(Equal(""))
					Expect(p.Description).To(Equal("Isolated cluster plugin"))
					Expect(p.Vendor).To(Equal("othervendor"))
					Expect(p.Publisher).To(Equal("otherpublisher"))
				})
			})
		})
		Describe("With a DB table with one plugin and no recommended version", func() {
			BeforeEach(func() {
				tmpDir, err = os.MkdirTemp(os.TempDir(), "")
				Expect(err).To(BeNil(), "unable to create temporary directory")

				// Create DB file
				dbFile, err = os.Create(filepath.Join(tmpDir, sqliteDBFileName))
				Expect(err).To(BeNil())
				// Open DB with the sqlite driver
				db, err := sql.Open("sqlite3", dbFile.Name())
				Expect(err).To(BeNil(), "failed to open the DB for testing")
				defer db.Close()

				// Create the table
				_, err = db.Exec(createTableStmt)
				Expect(err).To(BeNil(), "failed to create DB table for testing")

				// Add a plugin entry to the DB
				_, err = db.Exec(createPluginTMCNoRecommendedVersionStmt)
				Expect(err).To(BeNil(), "failed to create plugin for testing")

				inventory = NewSQLiteInventory(tmpDir, tmpDir)
			})
			AfterEach(func() {
				os.RemoveAll(tmpDir)
			})
			Context("When getting all plugins", func() {
				It("should return a list of one plugin with no error and recommended version set", func() {
					plugins, err = inventory.GetAllPlugins()
					Expect(err).ToNot(HaveOccurred())
					Expect(len(plugins)).To(Equal(1))

					p := plugins[0]
					Expect(p.Name).To(Equal("management-cluster"))
					Expect(len(p.Artifacts)).To(Equal(2))
					Expect(p.Artifacts["v0.0.1"]).ToNot(BeNil())
					Expect(p.Artifacts["v0.0.2"]).ToNot(BeNil())

					Expect(p.RecommendedVersion).To(Equal("v0.0.2"))
					Expect(string(p.Target)).To(Equal("mission-control"))
					Expect(p.Vendor).To(Equal("vmware"))
					Expect(p.Publisher).To(Equal("tmc"))

					artifactList := p.Artifacts["v0.0.2"]
					Expect(len(artifactList)).To(Equal(1))
					a := artifactList[0]
					Expect(a.OS).To(Equal("linux"))
					Expect(a.Arch).To(Equal("amd64"))
					Expect(a.Digest).To(Equal("1111111111"))
					Expect(a.Image).To(Equal(tmpDir + "/vmware/tmc/linux/amd64/tmc/management-cluster:v0.0.2"))
				})
			})
			Context("When getting a non-existent plugin version for tmc for an os/arch", func() {
				It("should return an empty list of plugins no error", func() {
					plugins, err = inventory.GetPlugins(&PluginInventoryFilter{
						Name:    "management-cluster",
						Target:  "mission-control",
						Version: "v1.2.3",
						OS:      "windows",
						Arch:    "amd64",
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(len(plugins)).To(Equal(0))
				})
			})
		})
	})
})