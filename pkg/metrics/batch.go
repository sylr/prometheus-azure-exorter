package metrics

import (
	"context"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/batch/2019-08-01.10.0/batch"
	azurebatch "github.com/Azure/azure-sdk-for-go/services/batch/mgmt/2019-08-01/batch"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/sylr/prometheus-azure-exporter/pkg/azure"
	"github.com/sylr/prometheus-azure-exporter/pkg/config"
	qdsync "sylr.dev/libqd/sync"
)

var (
	mu                        = sync.Mutex{}
	batchPoolQuota            = newBatchPoolQuota()
	batchDedicatedCoreQuota   = newBatchDedicatedCoreQuota()
	batchPoolsDedicatedNodes  = newBatchPoolsDedicatedNodes()
	batchPoolsNodesState      = newBatchPoolsNodesState()
	batchPoolsAllocationState = newBatchPoolsAllocationState()
	batchPoolsMetadata        = newBatchPoolsMetadata()
	batchJobsTasksActive      = newBatchJobsTasksActive()
	batchJobsTasksRunning     = newBatchJobsTasksRunning()
	batchJobsTasksCompleted   = newBatchJobsTasksCompleted()
	batchJobsTasksSucceeded   = newBatchJobsTasksSucceeded()
	batchJobsTasksFailed      = newBatchJobsTasksFailed()
	batchJobsInfo             = newBatchJobsInfo()
	batchJobsStates           = newBatchJobsStates()
	batchJobsMetadata         = newBatchJobsMetadata()
)

// -----------------------------------------------------------------------------

func newBatchPoolQuota() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "pool_quota",
			Help:      "Quota of pool for batch account",
		},
		[]string{"subscription", "resource_group", "account"},
	)
}

func newBatchDedicatedCoreQuota() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "dedicated_core_quota",
			Help:      "Quota of dedicated core for batch account",
		},
		[]string{"subscription", "resource_group", "account"},
	)
}

func newBatchPoolsDedicatedNodes() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "pool_dedicated_nodes",
			Help:      "Number of dedicated nodes for batch pool",
		},
		[]string{"subscription", "resource_group", "account", "pool"},
	)
}

func newBatchPoolsNodesState() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "pool_node_state",
			Help:      "Number of nodes for each states",
		},
		[]string{"subscription", "resource_group", "account", "pool", "state"},
	)
}

func newBatchPoolsAllocationState() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "pool_allocation_state",
			Help:      "Allocation state of the pool",
		},
		[]string{"subscription", "resource_group", "account", "pool", "state"},
	)
}

func newBatchPoolsMetadata() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "pool_metadata",
			Help:      "Informative vector with pool metadata",
		},
		[]string{"subscription", "resource_group", "account", "pool", "metadata", "value"},
	)
}

func newBatchJobsTasksActive() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_tasks_active",
			Help:      "Number of active batch job task",
		},
		[]string{"subscription", "resource_group", "account", "job_id"},
	)
}

func newBatchJobsTasksRunning() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_tasks_running",
			Help:      "Number of running batch job task",
		},
		[]string{"subscription", "resource_group", "account", "job_id"},
	)
}

func newBatchJobsTasksCompleted() *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_tasks_completed_total",
			Help:      "Total number of completed batch job task",
		},
		[]string{"subscription", "resource_group", "account", "job_id"},
	)
}

func newBatchJobsTasksSucceeded() *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_tasks_succeeded_total",
			Help:      "Total number of succeeded batch job task",
		},
		[]string{"subscription", "resource_group", "account", "job_id"},
	)
}

func newBatchJobsTasksFailed() *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_tasks_failed_total",
			Help:      "Total number of failed batch job task",
		},
		[]string{"subscription", "resource_group", "account", "job_id"},
	)
}

func newBatchJobsInfo() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_info",
			Help:      "Informative vector about job",
		},
		[]string{"subscription", "resource_group", "account", "job_id", "job_name", "pool"},
	)
}

func newBatchJobsStates() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_state",
			Help:      "State of job",
		},
		[]string{"subscription", "resource_group", "account", "job_id", "state"},
	)
}

func newBatchJobsMetadata() *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "azure",
			Subsystem: "batch",
			Name:      "job_metadata",
			Help:      "Informative vector with job metadata",
		},
		[]string{"subscription", "resource_group", "account", "job", "metadata", "value"},
	)
}

// -----------------------------------------------------------------------------

func init() {
	prometheus.MustRegister(batchPoolQuota)
	prometheus.MustRegister(batchDedicatedCoreQuota)
	prometheus.MustRegister(batchPoolsDedicatedNodes)
	prometheus.MustRegister(batchPoolsNodesState)
	prometheus.MustRegister(batchPoolsAllocationState)
	prometheus.MustRegister(batchPoolsMetadata)
	prometheus.MustRegister(batchJobsTasksActive)
	prometheus.MustRegister(batchJobsTasksRunning)
	prometheus.MustRegister(batchJobsTasksCompleted)
	prometheus.MustRegister(batchJobsTasksSucceeded)
	prometheus.MustRegister(batchJobsTasksFailed)
	prometheus.MustRegister(batchJobsInfo)
	prometheus.MustRegister(batchJobsStates)
	prometheus.MustRegister(batchJobsMetadata)

	if GetUpdateMetricsFunctionInterval("batch") == nil {
		RegisterUpdateMetricsFunction("batch", UpdateBatchMetrics)
	}
}

// UpdateBatchMetrics updates batch metrics
func UpdateBatchMetrics(ctx context.Context) error {
	var err error

	contextLogger := log.WithFields(log.Fields{
		"_id":   ctx.Value("id").(string),
		"_func": "UpdateBatchMetrics",
	})

	azureClients := azure.NewAzureClients()
	sub, err := azure.GetSubscription(ctx, azureClients, os.Getenv("AZURE_SUBSCRIPTION_ID"))

	if err != nil {
		contextLogger.Errorf("Unable to get subscription: %s", err)
		return err
	}

	batchAccounts, err := azure.ListSubscriptionBatchAccounts(ctx, azureClients, sub)

	if err != nil {
		contextLogger.Errorf("Unable to list account azure batch accounts: %s", err)
		return err
	}

	// Create new metric vectors
	nextBatchPoolQuota := newBatchPoolQuota()
	nextBatchDedicatedCoreQuota := newBatchDedicatedCoreQuota()
	nextBatchPoolsDedicatedNodes := newBatchPoolsDedicatedNodes()
	nextBatchPoolsNodesState := newBatchPoolsNodesState()
	nextBatchPoolsAllocationState := newBatchPoolsAllocationState()
	nextBatchPoolsMetadata := newBatchPoolsMetadata()
	nextBatchJobsTasksActive := newBatchJobsTasksActive()
	nextBatchJobsTasksRunning := newBatchJobsTasksRunning()
	nextBatchJobsTasksCompleted := newBatchJobsTasksCompleted()
	nextBatchJobsTasksSucceeded := newBatchJobsTasksSucceeded()
	nextBatchJobsTasksFailed := newBatchJobsTasksFailed()
	nextBatchJobsInfo := newBatchJobsInfo()
	nextBatchJobsStates := newBatchJobsStates()
	nextBatchJobsMetadata := newBatchJobsMetadata()

	wg := qdsync.NewCancelableWaitGroup(ctx, 50)

	for i := range *batchAccounts {
		accountProperties, _ := azure.ParseResourceID(*(*batchAccounts)[i].ID)

		// logger
		accountLogger := contextLogger.WithFields(log.Fields{
			"rg":      accountProperties.ResourceGroup,
			"account": *(*batchAccounts)[i].Name,
		})

		// Autodiscovery
		if !config.MustDiscoverBasedOnTags((*batchAccounts)[i].Tags) {
			accountLogger.Debugf("Account skipped by autodiscovery")
			continue
		}

		// Metrics
		nextBatchPoolQuota.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *(*batchAccounts)[i].Name).Set(float64(*(*batchAccounts)[i].PoolQuota))
		nextBatchDedicatedCoreQuota.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *(*batchAccounts)[i].Name).Set(float64(*(*batchAccounts)[i].DedicatedCoreQuota))

		// -- POOLS ------------------------------------------------------------

		pools, err := azure.ListBatchAccountPools(ctx, azureClients, sub, &(*batchAccounts)[i])

		if err != nil {
			accountLogger.Errorf("Unable to list account `%s` pools: %s", *(*batchAccounts)[i].Name, err)
		} else {
			for _, pool := range pools {
				wg.Add(1)

				go func(account *azurebatch.Account, pool azurebatch.Pool) {
					// Pool allocation state
					for _, state := range batch.PossibleAllocationStateValues() {
						nextBatchPoolsAllocationState.DeleteLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *pool.Name, string(state))
					}

					nextBatchPoolsAllocationState.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *pool.Name, string(pool.AllocationState)).Set(1)

					// Nodes state
					for _, state := range batch.PossibleComputeNodeStateValues() {
						nextBatchPoolsNodesState.DeleteLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *pool.Name, string(state))
					}

					nextBatchPoolsDedicatedNodes.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *pool.Name).Set(float64(*pool.PoolProperties.CurrentDedicatedNodes))

					// Metadata
					if pool.Metadata != nil {
						for _, metadata := range *pool.Metadata {
							nextBatchPoolsMetadata.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *pool.Name, *metadata.Name, *metadata.Value).Set(1)
						}
					}

					nodes, err := azure.ListBatchComputeNodes(ctx, azureClients, sub, account, &pool)

					if err != nil {
						accountLogger.WithFields(log.Fields{}).Error(err.Error())
					} else {
						for _, node := range *nodes {
							nextBatchPoolsNodesState.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *pool.Name, string(node.State)).Inc()
						}
					}

					accountLogger.WithFields(log.Fields{
						"metric":          "pool",
						"pool":            *pool.Name,
						"dedicated_nodes": *pool.PoolProperties.CurrentDedicatedNodes,
					}).Debug("")

					wg.Done()
				}(&(*batchAccounts)[i], pool)
			}
		}

		// -- JOBS -------------------------------------------------------------

		jobs, err := azure.ListBatchAccountJobs(ctx, azureClients, sub, &(*batchAccounts)[i])

		if err != nil {
			accountLogger.Errorf("Unable to list account jobs: %s", err)
		} else {
			for _, job := range jobs {
				wg.Add(1)

				go func(account *azurebatch.Account, job batch.CloudJob) {
					jobLogger := accountLogger.WithFields(log.Fields{
						"job_id": *job.ID,
					})

					// job.DisplayName can be nil but we don't want that
					displayName := *job.ID
					if job.DisplayName != nil {
						displayName = *job.DisplayName
					} else {
						jobLogger.Debugf("Job has no display name, defaulting to job.ID")
					}

					// <!-- metrics
					// We init JobStateActive state to 0 to be sure to have a value for each jobs so we can have alerts on the state value.
					nextBatchJobsStates.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID, string(batch.JobStateActive)).Set(0)
					nextBatchJobsStates.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID, string(job.State)).Set(1)
					// metrics -->

					// job metadata
					if job.Metadata != nil {
						for _, metadata := range *job.Metadata {
							// <!-- metrics
							nextBatchJobsMetadata.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID, *metadata.Name, *metadata.Value).Set(1)
							// metrics -->
						}
					}

					// job task count
					taskCounts, err := azure.GetBatchJobTaskCounts(ctx, azureClients, sub, account, &job)

					if err != nil {
						jobLogger.Errorf("Unable to get jobs task count: %s", err)
					} else {
						// <!-- metrics
						nextBatchJobsTasksActive.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID).Set(float64(*taskCounts.Active))
						nextBatchJobsTasksRunning.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID).Set(float64(*taskCounts.Running))
						nextBatchJobsTasksCompleted.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID).Set(float64(*taskCounts.Completed))
						nextBatchJobsTasksSucceeded.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID).Set(float64(*taskCounts.Succeeded))
						nextBatchJobsTasksFailed.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID).Set(float64(*taskCounts.Failed))
						nextBatchJobsInfo.WithLabelValues(*sub.DisplayName, accountProperties.ResourceGroup, *account.Name, *job.ID, displayName, *job.PoolInfo.PoolID).Set(1)
						// metrics -->

						jobLogger.WithFields(log.Fields{
							"metric":    "job",
							"job":       displayName,
							"pool":      *job.PoolInfo.PoolID,
							"active":    *taskCounts.Active,
							"running":   *taskCounts.Running,
							"completed": *taskCounts.Completed,
							"succeeded": *taskCounts.Succeeded,
							"failed":    *taskCounts.Failed,
						}).Debug("")
					}

					wg.Done()
				}(&(*batchAccounts)[i], job)
			}
		}
		// ----------------------------------------------------------- JOBS --!>
	}

	wg.Wait()

	// swapping current registered metrics with updated copies
	mu.Lock()
	*batchPoolQuota = *nextBatchPoolQuota
	*batchDedicatedCoreQuota = *nextBatchDedicatedCoreQuota
	*batchPoolsDedicatedNodes = *nextBatchPoolsDedicatedNodes
	*batchPoolsNodesState = *nextBatchPoolsNodesState
	*batchPoolsAllocationState = *nextBatchPoolsAllocationState
	*batchPoolsMetadata = *nextBatchPoolsMetadata
	*batchJobsTasksActive = *nextBatchJobsTasksActive
	*batchJobsTasksRunning = *nextBatchJobsTasksRunning
	*batchJobsTasksCompleted = *nextBatchJobsTasksCompleted
	*batchJobsTasksSucceeded = *nextBatchJobsTasksSucceeded
	*batchJobsTasksFailed = *nextBatchJobsTasksFailed
	*batchJobsInfo = *nextBatchJobsInfo
	*batchJobsStates = *nextBatchJobsStates
	*batchJobsMetadata = *nextBatchJobsMetadata
	mu.Unlock()

	return err
}
