# RBAC

## Table of Contents
1. [RBAC](#RBAC)
2. [Authentication Providers](#Authentication-Providers)
3. [Local Qase Reporting](#Local-Qase-Reporting)

### RBAC

In the RBAC tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Add a cluster owner/project member to the cluster
4. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

The config file is going to be the exact same as what is seen when provisioning clusters; there are no additional details that you need to do. For a detailed reference, please see the [provisioning README](../provisioning/README.md). To see a sample Linode K3s config, please see below:

```yaml
rancher:
    host: "rancher_server_address"
    adminToken: "rancher_admin_token"
    insecure: true
    cleanup: true
terraform:
    cloudCredentialName: ""
    defaultClusterRoleForProjectMembers: "true"
    enableNetworkPolicy: false
    hostnamePrefix: ""
    machineConfigName: ""
    module: "linode_k3s"
    cni: "canal"
    nodeTemplateName: ""                # Needed for RKE1 clusters
    linodeCredentials:
        linodeToken: ""
    linodeConfig:
        linodeImage: "linode/ubuntu22.04"
        region: "us-east"
        linodeRootPass: ""
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/rbac --junitfile results/results.xml --jsonfile results/results.json -- -timeout=60m -v -run "TestTfpRBACTestSuite/TestTfpRBAC$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

### Authentication Providers

In the Auth Providers tests, the following workflow is followed:

1. Enable an authentication provider
2. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

This test has static test cases, where multiple authentication providers are enabled and disabled. Additionally, there exists a dynamicInput function where you can specify a single authenticated provider. For the dynamic tests, an example is shown below:

```yaml
rancher:
    host: "rancher_server_address"
    adminToken: "rancher_admin_token"
    insecure: true
    cleanup: true
terraform:
    authProvider: "github"             # Supported providers are: ad | azureAD | github | okta | openLDAP
    githubConfig:
    clientId: "<client id>"
    clientSecret: "<client secret>"
    resourcePrefix: ""
terratest:
    tfLogging: true
```

If you are running the static tests that will run all of the supported authenticated providers, your config will look like this:

```yaml
rancher:
    host: "rancher_server_address"
    adminToken: "rancher_admin_token"
    insecure: true
    cleanup: true
terraform:
    adConfig:
        port: 389
        servers: [""]
        serviceAccountPassword: ""
        serviceAccountUsername: ""
        userSearchBase: ""
        testUsername: ""
        testPassword: ""
    azureADConfig:
        applicationID: ""
        applicationSecret: ""
        authEndpoint: ""
        graphEndpoint: ""
        tenantID: ""
        tokenEndpoint: ""
    githubConfig:
        clientId: ""
        clientSecret: ""
    oktaConfig:
        displayNameField: ""
        groupsField: ""
        idpMetadataContent: |
            <placeholder>
        spCert: |
            <placeholder>
        spKey:  |
            <placeholder>>
        uidField: ""
        userNameField: ""
    openLDAPConfig:
        port: 389
        servers: [""]
        serviceAccountDistinguishedName: ""
        serviceAccountPassword: ""
        userSearchBase: ""
        testUsername: ""
        testPassword: ""
    resourcePrefix: ""
terratest:
    tfLogging: true
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/rbac --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpAuthConfigTestSuite/TestTfpAuthConfig$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.

## Local Qase Reporting
If you are planning to report to Qase locally, then you will need to have the following done:
1. The `terratest` block in your config file must have `localQaseReporting: true`.
2. The working shell session must have the following two environmental variables set:
     - `QASE_AUTOMATION_TOKEN=""`
     - `QASE_TEST_RUN_ID=""`
3. Append `./reporter` to the end of the `gotestsum` command. See an example below::
     - `gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rancher2/rbac --junitfile results.xml --jsonfile results.json -- -timeout=60m -v -run "TestTfpRBACTestSuite/TestTfpRBAC$";/path/to/tfp-automation/reporter`