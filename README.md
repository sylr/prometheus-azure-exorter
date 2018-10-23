prometheus-azure-exporter
=========================

[![Docker Repository on Quay](https://quay.io/repository/sylr/prometheus-azure-exporter/status "Docker Repository on Quay")](https://quay.io/repository/sylr/prometheus-azure-exporter)


This is a daemon which calls Azure API to fetch resources metrics and expose them
with HTTP using the prometheus format.

History
-------

After several incidents in Production with Azure Batch we decided that we needed something better
in terms of monitoring than what Microsoft is currently proposing.

Disclaimer
----------

This is my 2ng Go project so It is far from being perfect in terms of design and implementation.

You are very welcome to open issues and pull requests if you want to improve it.

Azure resources
---------------

| Namespaces              | Metrics                                | Labels
|-------------------------|----------------------------------------|-----------------------------------------------
| Azure                   | azure_api_calls_total                  |
|                         | azure_api_calls_failed_total           |
|                         | azure_api_batch_calls_total            | subscription, resource_group, account
|                         | azure_api_batch_calls_failed_total     | subscription, resource_group, account
| Batch                   | azure_batch_pool_quota                 | subscription, resource_group, account
|                         | azure_batch_dedicated_core_quota       | subscription, resource_group, account
|                         | azure_batch_pool_dedicated_nodes       | subscription, resource_group, account, pool
|                         | azure_batch_job_tasks_active           | subscription, resource_group, account, job_id, job_name
|                         | azure_batch_job_tasks_running          | subscription, resource_group, account, job_id, job_name
|                         | azure_batch_job_tasks_completed_total  | subscription, resource_group, account, job_id, job_name
|                         | azure_batch_job_tasks_succeeded_total  | subscription, resource_group, account, job_id, job_name
|                         | azure_batch_job_tasks_failed_total     | subscription, resource_group, account, job_id, job_name