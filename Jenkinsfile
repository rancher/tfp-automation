#!groovy
node {
  def testsDir = "./tests/${env.TEST_PACKAGE}"
  def branch = "${env.BRANCH}"
  if ("${env.BRANCH}" != "null" && "${env.BRANCH}" != "") {
        branch = "${env.BRANCH}"
  }
  def repo = scm.userRemoteConfigs
  if ("${env.REPO}" != "null" && "${env.REPO}" != "") {
    repo = [[url: "${env.REPO}"]]
  }
  def timeout = "${env.TIMEOUT}"
  if ("${env.TIMEOUT}" != "null" && "${env.TIMEOUT}" != "") {
        timeout = "${env.TIMEOUT}" 
  }
  def rancher2ProviderVersion = "${env.RANCHER2_PROVIDER_VERSION}"
  if ("${env.RANCHER2_PROVIDER_VERSION}" != "null" && "${env.RANCHER2_PROVIDER_VERSION}" != "") {
        rancher2ProviderVersion = "${env.RANCHER2_PROVIDER_VERSION}" 
  }
  stage('Checkout') {
          deleteDir()
          checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${branch}"]],
                    extensions: scm.extensions + [[$class: 'CleanCheckout']],
                    userRemoteConfigs: repo
                  ])
        }
    stage('Build Docker image') {
            writeFile file: 'config.yml', text: env.CONFIG
            env.CATTLE_TEST_CONFIG='/home/jenkins/workspace/rancher_qa/tfp-automation/config.yml'
            def test = "docker build --build-arg CONFIG_FILE=config.yml --build-arg RANCHER2_PROVIDER_VERSION=\"${rancher2ProviderVersion}\" -f Dockerfile -t tfp-automation . "
            sh test
    }
    
    stage('Run Module Test') {
            def dockerImage = docker.image("tfp-automation")
            dockerImage.inside() {
                sh "go test -v -timeout ${timeout} -run ${params.TEST_CASE} ${testsDir}"
            }
    }
}