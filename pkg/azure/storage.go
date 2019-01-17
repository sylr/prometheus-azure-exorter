package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2018-07-01/storage"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/sylr/prometheus-azure-exporter/pkg/tools"
)

var (
	cacheKeySubscriptionStorageAccounts = `sub-%s-storageaccounts`
	cacheKeyStorageAccountContainers    = `sub-%s-rg-%s-storageaccount-%s-containers`
	cacheKeyStorageAccountKeys          = `sub-%s-rg-%s-storageaccount-%s-keys`
)

var (
	// AzureAPIStorageCallsTotal Total number of Azure Storage API calls
	AzureAPIStorageCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "azure_api",
			Subsystem: "storage",
			Name:      "calls_total",
			Help:      "Total number of calls to the Azure API",
		},
		[]string{},
	)

	// AzureAPIStorageCallsFailedTotal Total number of failed Azure Storage API calls
	AzureAPIStorageCallsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "azure_api",
			Subsystem: "storage",
			Name:      "calls_failed_total",
			Help:      "Total number of failed calls to the Azure API",
		},
		[]string{},
	)

	// AzureAPIStorageCallsDurationSecondsBuckets Histograms of Azure Storage API calls durations in seconds
	AzureAPIStorageCallsDurationSecondsBuckets = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "azure_api",
			Subsystem: "storage",
			Name:      "calls_duration_seconds",
			Help:      "Histograms of Azure Storage API calls durations in seconds",
			Buckets:   []float64{0.10, 0.15, 0.20, 0.50, 1.0, 2.0, 3.0, 5.0},
		},
		[]string{},
	)
)

func init() {
	prometheus.MustRegister(AzureAPIStorageCallsTotal)
	prometheus.MustRegister(AzureAPIStorageCallsFailedTotal)
	prometheus.MustRegister(AzureAPIStorageCallsDurationSecondsBuckets)
}

// ObserveAzureStorageAPICall ...
func ObserveAzureStorageAPICall(duration float64, labels ...string) {
	AzureAPIStorageCallsTotal.WithLabelValues(labels...).Inc()
	AzureAPIStorageCallsDurationSecondsBuckets.WithLabelValues(labels...).Observe(duration)
}

// ObserveAzureStorageAPICallFailed ...
func ObserveAzureStorageAPICallFailed(duration float64, labels ...string) {
	AzureAPIStorageCallsFailedTotal.WithLabelValues(labels...).Inc()
}

// ListSubscriptionStorageAccounts ...
func ListSubscriptionStorageAccounts(ctx context.Context, clients *AzureClients, subscriptionID string) (*[]storage.Account, error) {
	c := tools.GetCache(5 * time.Minute)
	cacheKey := fmt.Sprintf(cacheKeySubscriptionStorageAccounts, subscriptionID)

	contextLogger := log.WithFields(log.Fields{
		"_id":          ctx.Value("id").(string),
		"subscription": subscriptionID,
	})

	if caccounts, ok := c.Get(cacheKey); ok {
		if accounts, ok := caccounts.(*[]storage.Account); ok {
			//contextLogger.Debugf("Got []storage.Account from cache")
			return accounts, nil
		} else {
			contextLogger.Errorf("Failed to cast object from cache back to []storage.Account")
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	client, err := clients.GetStorageAccountsClient(subscriptionID)

	if err != nil {
		return nil, err
	}

	t0 := time.Now()
	accounts, err := client.List(ctx)
	t1 := time.Since(t0).Seconds()

	ObserveAzureAPICall(t1)

	if err != nil {
		ObserveAzureAPICallFailed(t1)
		return nil, err
	}

	vals := accounts.Value
	c.SetDefault(cacheKey, vals)

	return vals, nil
}

// ListStorageAccountContainers ...
func ListStorageAccountContainers(ctx context.Context, clients *AzureClients, account *storage.Account) (*[]storage.ListContainerItem, error) {
	c := tools.GetCache(5 * time.Minute)

	accountResourceDetails, err := ParseResourceID(*account.ID)

	cacheKey := fmt.Sprintf(
		cacheKeyStorageAccountContainers,
		accountResourceDetails.SubscriptionID,
		accountResourceDetails.ResourceGroup,
		*account.Name,
	)

	contextLogger := log.WithFields(log.Fields{
		"_id":             ctx.Value("id").(string),
		"storage_account": *account.Name,
	})

	if ccontainers, ok := c.Get(cacheKey); ok {
		if containers, ok := ccontainers.(*[]storage.ListContainerItem); ok {
			//contextLogger.Debugf("Got []storage.ListContainerItem from cache")
			return containers, nil
		} else {
			contextLogger.Errorf("Failed to cast object from cache back to []storage.ListContainerItem")
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	client, err := clients.GetBlobContainersClient(accountResourceDetails.SubscriptionID)

	if err != nil {
		return nil, err
	}

	t0 := time.Now()
	containers, err := client.List(ctx, accountResourceDetails.ResourceGroup, *account.Name)
	t1 := time.Since(t0).Seconds()

	ObserveAzureAPICall(t1)
	ObserveAzureStorageAPICall(t1)

	if err != nil {
		ObserveAzureAPICallFailed(t1)
		ObserveAzureStorageAPICallFailed(t1)
		return nil, err
	}

	vals := *containers.Value
	c.SetDefault(cacheKey, &vals)

	return &vals, nil
}

// ListStorageAccountKeys ...
func ListStorageAccountKeys(ctx context.Context, clients *AzureClients, account *storage.Account) (*[]storage.AccountKey, error) {
	c := tools.GetCache(5 * time.Minute)

	accountResourceDetails, err := ParseResourceID(*account.ID)

	cacheKey := fmt.Sprintf(
		cacheKeyStorageAccountKeys,
		accountResourceDetails.SubscriptionID,
		accountResourceDetails.ResourceGroup,
		*account.Name,
	)

	contextLogger := log.WithFields(log.Fields{
		"_id":             ctx.Value("id").(string),
		"storage_account": *account.Name,
	})

	if ckeys, ok := c.Get(cacheKey); ok {
		if keys, ok := ckeys.(*[]storage.AccountKey); ok {
			//contextLogger.Debugf("Got []storage.AccountKey from cache")
			return keys, nil
		} else {
			contextLogger.Errorf("Failed to cast object from cache back to []storage.AccountKey")
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	client, err := clients.GetStorageAccountsClient(accountResourceDetails.SubscriptionID)

	if err != nil {
		return nil, err
	}

	t0 := time.Now()
	keys, err := client.ListKeys(ctx, accountResourceDetails.ResourceGroup, *account.Name)
	t1 := time.Since(t0).Seconds()

	ObserveAzureAPICall(t1)
	ObserveAzureStorageAPICall(t1)

	if err != nil {
		ObserveAzureAPICallFailed(t1)
		ObserveAzureStorageAPICallFailed(t1)
		return nil, err
	}

	vals := *keys.Keys
	c.SetDefault(cacheKey, &vals)

	return &vals, nil
}
