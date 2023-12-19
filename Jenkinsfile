#!groovy
node {
//   def rootPath = "/root/go/src/github.com/josh-diamond/tfp-autm/"
//   def workPath = "/root/go/src/github.com/rancher/rancher/tests/v2/validation/"
  def workPath = "/home/jenkins/workspace/rancher_qa/tfp-automation"

  def job_name = "${JOB_NAME}"
  if (job_name.contains('/')) { 
    job_names = job_name.split('/')
    job_name = job_names[job_names.size() - 1] 
  }
//   def golangTestContainer = "${job_name}${env.BUILD_NUMBER}-golangtest"
  def buildTestContainer = "${job_name}${env.BUILD_NUMBER}-buildtest"
//   def cleanupTestContainer = "${job_name}${env.BUILD_NUMBER}-cleanuptest"
  def envFile = ".env"
  def golangImageName = "rancher-validation-${job_name}${env.BUILD_NUMBER}"

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
            try {
                sh "ls -la"
                sh "docker run --name ${buildTestContainer} -t --env-file ${envFile} " + 
              "${golangImageName} sh -c \"${workPath}pipeline/scripts/setup_environment.sh\""
            } catch(err) {
                sh "docker stop ${buildTestContainer}"
                sh "docker rm -v ${buildTestContainer}"
                // sh "docker volume rm -f ${validationVolume}"
                error "Build Environment had failures."
              }
    }
    
    
    stage('Run Module Test') {
            def dockerImage = docker.image('tfp-automation')
            dockerImage.inside() {
                sh "go test -v -timeout ${timeout} -run ${params.TEST_CASE} ${testsDir}"
            }
    }
}