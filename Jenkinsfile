#!groovy
node {
  def rootPath = "/home/jenkins/workspace/rancher_qa/tfp-automation/"
  def testsDir = "./tests/${env.TEST_PACKAGE}"
  def job_name = "${JOB_NAME}"
  if (job_name.contains('/')) { 
    job_names = job_name.split('/')
    job_name = job_names[job_names.size() - 1] 
  }
  def testResultsOut = "results.xml"
  def testResultsJSON = "results.json"
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
  def terraformVersion = "${env.TERRAFORM_VERSION}"
  if ("${env.TERRAFORM_VERSION}" != "null" && "${env.TERRAFORM_VERSION}" != "") {
        terraformVersion = "${env.TERRAFORM_VERSION}" 
  }
  def rancher2ProviderVersion = "${env.RANCHER2_PROVIDER_VERSION}"
  if ("${env.RANCHER2_PROVIDER_VERSION}" != "null" && "${env.RANCHER2_PROVIDER_VERSION}" != "") {
        rancher2ProviderVersion = "${env.RANCHER2_PROVIDER_VERSION}" 
  }
  withCredentials([ string(credentialsId: 'QASE_AUTOMATION_TOKEN', variable: 'QASE_AUTOMATION_TOKEN')]) {
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
            env.CATTLE_TEST_CONFIG=rootPath+'config.yml'
            sh "docker build --build-arg CONFIG_FILE=config.yml --build-arg TERRAFORM_VERSION=${terraformVersion} --build-arg RANCHER2_PROVIDER_VERSION=${rancher2ProviderVersion} -f Dockerfile -t tfp-automation . "
    }
    
    stage('Run Module Test') {
            def testResultsDir = rootPath+"results"
            sh "mkdir -p ${testResultsDir}"
            def dockerImage = docker.image('tfp-automation')
            if (rancher2ProviderVersion.contains("rc")) {
              dockerImage.inside("-u root -v ${testResultsDir}:/results") {
                try { 
                    sh "gotestsum --format standard-verbose --packages=${testsDir} --junitfile results/${testResultsOut} --jsonfile results/${testResultsJSON} -- -timeout=${timeout} -v ${params.TEST_CASE}"
                 } catch(err) {
                   echo 'Test run had failures. Collecting results...'
                 }
                 sh "${rootPath}pipeline/scripts/build_qase_reporter.sh"
                 if (fileExists("${rootPath}reporter")) {
                   sh "${rootPath}reporter"
                 } 
                }
              }
            else {
              dockerImage.inside("-u jenkins -v ${testResultsDir}:/results") {
                try {
                    sh "gotestsum --format standard-verbose --packages=${testsDir} --junitfile results/${testResultsOut} --jsonfile results/${testResultsJSON} -- -timeout=${timeout} -v ${params.TEST_CASE}"
                } catch(err) {
                  echo 'Test run had failures. Collecting results...'
                }
                sh "${rootPath}pipeline/scripts/build_qase_reporter.sh"
                if (fileExists("${rootPath}reporter")) {
                  sh "${rootPath}reporter"
                }       
              }
            }
            step([$class: 'JUnitResultArchiver', testResults: "**/results/${testResultsOut}"])
      }
    }
  }