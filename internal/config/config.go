// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/newrelic/nri-vsphere/internal/model"
	"github.com/newrelic/nri-vsphere/internal/performance"
	"github.com/newrelic/nri-vsphere/internal/tag"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	logrus "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
)

// ArgumentList Available Arguments
type ArgumentList struct {
	sdkArgs.DefaultArgumentList
	Local              bool   `default:"true" help:"Collect local entity info"`
	Entity             string `default:"" help:"Manually set a remote entity name"`
	URL                string `default:"" help:"Required: ESXi or vCenter SDK URL eg. https://172.16.53.129/sdk"`
	User               string `default:"" help:"Required: Username"`
	Pass               string `default:"" help:"Required: Password"`
	DatacenterLocation string `default:"" help:"Datacenter Location of your vCenter or ESXi Host eg. sydney-ultimo"`

	EnableVsphereEvents bool   `default:"false" help:"Set to collect vSphere events"`
	EventsPageSize      string `default:"100" help:"Number of events fetched from the vCenter in each call"`

	EnableVspherePerfMetrics bool   `default:"false" help:"Set to collect vSphere performance metrics"`
	PerfLevel                int    `default:"1" help:"Performance counter level of performance metrics that will be collected"`
	LogAvailableCounters     bool   `default:"false" help:"Print available performance metrics"`
	PerfMetricFile           string `default:"" help:"Location of performance metrics configuration file"`
	ConsiderInstances        bool   `default:"false" help:"The accumuated metrics of entities with multiple instances will be considered separetely"`

	//As a general rule, specify between 10 and 50 entities in a single call to the QueryPerf method.
	//This is a general recommendation because your system configuration may impose different
	//constraints.
	//https://vdc-download.vmware.com/vmwb-repository/dcr-public/cdbbd51c-4824-4a1b-ad43-45df55a76a76/8cb3ed93-cac2-46aa-b329-db5a096af5bc/vsphere-web-services-sdk-67-programming-guide.pdf
	BatchSizePerfEntities string `default:"50" help:"Number of entities requested at the same time when querying performance metrics"`
	BatchSizePerfMetrics  string `default:"50" help:"Number of metrics requested at the same time when querying performance metrics"`

	EnableVsphereTags      bool `default:"false" help:"Set to collect tags. Tags are available when connecting to vcenter"`
	EnableVsphereSnapshots bool `default:"false" help:"Set to collect and process VMs Snapshots data"`
	ValidateSSL            bool `default:"false" help:"Set to validates SSL when connecting to vCenter or Esxi Host"`
	ShowVersion            bool `default:"false" help:"Print build information and exit"`

	IncludeTags string `default:"" help:"Space-separated list of tag categories and values for resource inclusion. \nIf defined, only resources tagged with any of the tags will be included in the results. \nYou must also include 'enable_vsphere_tags' in order for this option to work. \nExample: --include_tags env=prod dc=eu"`
}

type Config struct {
	Args                 ArgumentList
	Integration          *integration.Integration // Integration Infrastructure SDK Integration
	Entity               *integration.Entity      // Entity Infrastructure SDK Entity
	Hostname             string                   // Hostname current host
	Logrus               *logrus.Logger           // Logrus create instance of the logger
	IntegrationName      string                   // IntegrationName name of integration
	IntegrationNameShort string                   // IntegrationNameShort Short Name
	IntegrationVersion   string                   // IntegrationVersion Version
	VMWareClient         *govmomi.Client          // VMWareClient Client
	ViewManager          *view.Manager            // ViewManager Client
	TagCollector         *tag.Collector           // TagsManager Client
	Datacenters          []*model.Datacenter      // Datacenters VMWare
	IsVcenterAPIType     bool                     // IsVcenterAPIType true if connecting to vcenter
	PerfCollector        *performance.PerfCollector
	startTime            time.Time // start time the integration started.
}

func New(buildVersion string) *Config {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\nFlags -events and -metrics are discarded\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	return &Config{
		Logrus:               logrus.New(),
		IntegrationName:      "com.newrelic.vsphere",
		IntegrationNameShort: "vsphere",
		IntegrationVersion:   buildVersion,
		IsVcenterAPIType:     false,
		startTime:            time.Now(),
	}
}

const (
	WindowsPerfMetricFile      = "C:\\Program Files\\New Relic\\newrelic-infra\\integrations.d\\vsphere-performance.metrics"
	LinuxDefaultPerfMetricFile = "/etc/newrelic-infra/integrations.d/vsphere-performance.metrics"
)

func (c *Config) TagCollectionEnabled() bool {
	return c.IsVcenterAPIType && c.Args.EnableVsphereTags
}

func (c *Config) EventCollectionEnabled() bool {
	return c.IsVcenterAPIType && c.Args.EnableVsphereEvents
}

func (c *Config) TagFilteringEnabled() bool {
	return c.TagCollectionEnabled() && len(c.Args.IncludeTags) > 0
}

func (c *Config) PerfMetricsCollectionEnabled() bool {
	return c.Args.EnableVspherePerfMetrics
}

func (c *Config) ConsiderInstancesEnabled() bool {
	return c.Args.ConsiderInstances
}

func (c *Config) Uptime() time.Duration {
	return time.Since(c.startTime)
}
