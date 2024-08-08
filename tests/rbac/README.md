# RBAC

## Table of Contents
1. [RBAC](#RBAC)
2. [Authentication Providers](#Auth)

### RBAC

In the RBAC tests, the following workflow is followed:

1. Provision a downstream cluster
2. Perform post-cluster provisioning checks
3. Add a cluster owner/project member to the cluster
4. Cleanup resources (Terraform explicitly needs to call its cleanup method so that each test doesn't experience caching issues)

The config file is going to be the exact same as what is seen when provisioning clusters; there are no additional details that you need to do. For reference, please see the [provisioning README](../provisioning/README.md).

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rbac --junitfile results.xml -- -timeout=60m -v -run "TestTfpRBACTestSuite/TestTfpRBAC$"`

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
```

See the below examples on how to run the tests:

`gotestsum --format standard-verbose --packages=github.com/rancher/tfp-automation/tests/rbac --junitfile results.xml -- -timeout=60m -v -run "TestTfpAuthConfigTestSuite/TestTfpAuthConfig$"`

If the specified test passes immediately without warning, try adding the -count=1 flag to get around this issue. This will avoid previous results from interfering with the new test run.