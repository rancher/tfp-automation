#!groovy
node {
  def rootPath = "/root/go/src/github.com/josh-diamond/tfp-automation/"
//   def workPath = "/root/go/src/github.com/rancher/rancher/tests/v2/validation/"
//   def workPath = "/home/jenkins/workspace/rancher_qa/tfp-automation"

  def job_name = "${JOB_NAME}"
  if (job_name.contains('/')) { 
    job_names = job_name.split('/')
    job_name = job_names[job_names.size() - 1] 
  }
  def testContainer = "${job_name}${env.BUILD_NUMBER}_test"
//   def golangTestContainer = "${job_name}${env.BUILD_NUMBER}-golangtest"
  def buildTestContainer = "${job_name}${env.BUILD_NUMBER}-buildtest"
//   def cleanupTestContainer = "${job_name}${env.BUILD_NUMBER}-cleanuptest"
  def envFile = ".env"
//   def testResultsOut = "results.xml"
  def golangImageName = "rancher-validation-${job_name}${env.BUILD_NUMBER}"
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
  def testcase = "${env.TEST_CASE}"
  if ("${env.TEST_CASE}" != "null" && "${env.TEST_CASE}" != "") {
        testcase = "${env.TEST_CASE}" 
  }
  def rancher2ProviderVersion = "${env.RANCHER2_PROVIDER_VERSION}"
  if ("${env.RANCHER2_PROVIDER_VERSION}" != "null" && "${env.RANCHER2_PROVIDER_VERSION}" != "") {
        rancher2ProviderVersion = "${env.RANCHER2_PROVIDER_VERSION}" 
  }
//   def testsDir = "github.com/josh-diamond/tfp-automation/tests/${env.TEST_PACKAGE}"
  def testsDir = "./tests/${env.TEST_PACKAGE}"
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
            env.CATTLE_TEST_CONFIG=rootPath+"config.yml"
            try {
                sh "chmod +x ./build.sh"
                sh "chmod +x ./configure.sh"
                sh "./configure.sh"
                sh "./build.sh"
            } catch(err) {
                sh "docker stop ${buildTestContainer}"
                sh "docker rm -v ${buildTestContainer}"
                error "Build Environment had failures."
              }
    }
    
    stage('Run Module Test') {
        //     def dockerImage = docker.image(${golangImageName})
        //     dockerImage.inside() {
        //         sh "go test -v -timeout ${timeout} -run ${params.TEST_CASE} ${testsDir}"
        //     }

        //          sh "docker run --name ${testContainer} -t --env-file ${envFile} " +
        //   "${golangImageName} sh -c \"/root/go/bin/gotestsum --format standard-verbose --packages=${testsDir} -- ${env.TEST_CASE} -timeout=${timeout} -v\""
        try {
          sh "docker run --name ${testContainer} -t --env-file ${envFile} " +
          "${golangImageName} sh -c \"/root/go/bin test -v -timeout ${timeout} ${params.TEST_CASE} ${testsDir}\""
        } catch(err) {
          echo 'Test run had failures. Collecting results...'
        }
    }
}