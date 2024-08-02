# RBAC

In the RBAC tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Add a cluster owner/project member to the cluster
4. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

The config file is going to be the exact same as what is seen when provisioning clusters; there are no additional details that you need to do. For reference, please see the [provisioning README](../provisioning/README.md).

See the below examples on how to run the tests:

### RBAC

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rbac --junitfile results.xml -- -timeout=60m -v -run "TestTfpRBACTestSuite/TestTfpRBAC$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.