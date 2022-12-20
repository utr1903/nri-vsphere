// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package process

import (
	"fmt"

	"github.com/newrelic/nri-vsphere/internal/config"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/vmware/govmomi/vim25/types"
)

func createDatastoreSamples(config *config.Config) {
	for _, dc := range config.Datacenters {
		for _, ds := range dc.Datastores {

			// filtering here will to avoid sending data to backend
			if config.TagFilteringEnabled() && !config.TagCollector.MatchObjectTags(ds.Self) {
				continue
			}

			datacenterName := dc.Datacenter.Name

			entityName := sanitizeEntityName(config, ds.Summary.Name, datacenterName)

			dataStoreID := ds.Summary.Url

			e, ms, err := createNewEntityWithMetricSet(config, entityTypeDatastore, entityName, dataStoreID)
			if err != nil {
				config.Logrus.WithError(err).WithField("datastoreName", entityName).WithField("dataStoreID", dataStoreID).Error("failed to create metricSet")
				continue
			}

			if config.Args.DatacenterLocation != "" {
				checkError(config.Logrus, ms.SetMetric("datacenterLocation", config.Args.DatacenterLocation, metric.ATTRIBUTE))
			}

			if config.IsVcenterAPIType {
				checkError(config.Logrus, ms.SetMetric("datacenterName", datacenterName, metric.ATTRIBUTE))
			}

			checkError(config.Logrus, ms.SetMetric("name", ds.Summary.Name, metric.ATTRIBUTE))
			checkError(config.Logrus, ms.SetMetric("fileSystemType", ds.Summary.Type, metric.ATTRIBUTE))
			checkError(config.Logrus, ms.SetMetric("overallStatus", string(ds.OverallStatus), metric.ATTRIBUTE))
			checkError(config.Logrus, ms.SetMetric("accessible", fmt.Sprintf("%t", ds.Summary.Accessible), metric.ATTRIBUTE))
			checkError(config.Logrus, ms.SetMetric("vmCount", len(ds.Vm), metric.GAUGE))
			checkError(config.Logrus, ms.SetMetric("hostCount", len(ds.Host), metric.GAUGE))
			checkError(config.Logrus, ms.SetMetric("url", ds.Summary.Url, metric.ATTRIBUTE))
			checkError(config.Logrus, ms.SetMetric("capacity", float64(ds.Summary.Capacity)/(1<<30), metric.GAUGE))
			checkError(config.Logrus, ms.SetMetric("freeSpace", float64(ds.Summary.FreeSpace)/(1<<30), metric.GAUGE))
			checkError(config.Logrus, ms.SetMetric("uncommitted", float64(ds.Summary.Uncommitted)/(1<<30), metric.GAUGE))

			switch info := ds.Info.(type) {
			case *types.NasDatastoreInfo:
				if info.Nas != nil {
					checkError(config.Logrus, ms.SetMetric("nas.remoteHost", info.Nas.RemoteHost, metric.ATTRIBUTE))
					checkError(config.Logrus, ms.SetMetric("nas.remotePath", info.Nas.RemotePath, metric.ATTRIBUTE))
				}
			}

			// Tags
			if config.TagCollectionEnabled() {
				tagsByCategory := config.TagCollector.GetTagsByCategories(ds.Self)
				for k, v := range tagsByCategory {
					checkError(config.Logrus, ms.SetMetric(tagsPrefix+k, v, metric.ATTRIBUTE))
					// add tags to inventory due to the inventory workaround
					addTagsToInventory(config, e, k, v)
				}
			}

			// Performance metrics
			if config.PerfMetricsCollectionEnabled() {
				perfMetrics := dc.GetPerfMetrics(ds.Self)

				for _, perfMetric := range perfMetrics {
					checkError(config.Logrus, ms.SetMetric(perfMetricPrefix+perfMetric.Counter, perfMetric.Value, metric.GAUGE))

					// Build metrics block for the instance metrics
					if config.ConsiderInstancesEnabled() {
						for key, val := range perfMetric.InstanceValues {
							_, ims, err := createNewEntityWithMetricSet(config, entityTypeHost+"Instance", entityName, dataStoreID)
							if err != nil {
								config.Logrus.WithError(err).
									WithField("datastoreName", entityName).
									WithField("dataStoreID", dataStoreID).
									Error("failed to create metricSet")
							} else {
								// Add attributes
								checkError(config.Logrus, ims.SetMetric("dataStoreID", fmt.Sprintf("%v", dataStoreID), metric.ATTRIBUTE))
								checkError(config.Logrus, ims.SetMetric("datacenterLocation", fmt.Sprintf("%v", config.Args.DatacenterLocation), metric.ATTRIBUTE))
								checkError(config.Logrus, ims.SetMetric("datacenterName", fmt.Sprintf("%v", datacenterName), metric.ATTRIBUTE))

								// Add metric of the instance
								checkError(config.Logrus, ims.SetMetric("instanceName", fmt.Sprintf("%v", key), metric.ATTRIBUTE))
								checkError(config.Logrus, ims.SetMetric(perfMetricPrefix+perfMetric.Counter, val, metric.GAUGE))
							}
						}
					}
				}
			}
		}
	}
}
