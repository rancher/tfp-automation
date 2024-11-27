#!groovy
node {
  def homePath = pwd() + "/"
  def rootPath = "/root/go/src/github.com/rancher/tfp-automation/"
  def testsDir = "github.com/rancher/tfp-automation/tests/${env.TEST_PACKAGE}"
  def standaloneTestsDir = "github.com/rancher/tfp-automation/standalone/tests/${env.TEST_PACKAGE}"
  def job_name = "${JOB_NAME}"
  if (job_name.contains('/')) { 
    job_names = job_name.split('/')
    job_name = job_names[job_names.size() - 1] 
  }
  def testContainer = "${job_name}${env.BUILD_NUMBER}_test"
  def imageName = "tfp-automation-validation-${job_name}${env.BUILD_NUMBER}"
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
  def localProviderVersion = "${env.LOCALS_PROVIDER_VERSION}"
  if ("${env.LOCALS_PROVIDER_VERSION}" != "null" && "${env.LOCALS_PROVIDER_VERSION}" != "") {
        localProviderVersion = "${env.LOCALS_PROVIDER_VERSION}" 
  }
  def awsProviderVersion = "${env.AWS_PROVIDER_VERSION}"
  if ("${env.AWS_PROVIDER_VERSION}" != "null" && "${env.AWS_PROVIDER_VERSION}" != "") {
        awsProviderVersion = "${env.AWS_PROVIDER_VERSION}" 
  }
  withCredentials([ string(credentialsId: 'AWS_ACCESS_KEY_ID', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'AWS_SECRET_ACCESS_KEY', variable: 'AWS_SECRET_ACCESS_KEY'),
                    string(credentialsId: 'AWS_SSH_PEM_KEY', variable: 'AWS_SSH_PEM_KEY'),
                    string(credentialsId: 'QASE_AUTOMATION_TOKEN', variable: 'QASE_AUTOMATION_TOKEN')]) {
  stage('Checkout') {
          deleteDir()
          checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${branch}"]],
                    extensions: scm.extensions + [[$class: 'CleanCheckout']],
                    userRemoteConfigs: repo
                  ])
        }
        try {
          stage('Configure and Build') {
            writeFile file: 'config.yml', text: env.CONFIG
            writeFile file: 'key.pem', text: params.PEM_FILE
            env.CATTLE_TEST_CONFIG=rootPath+'config.yml'
            sh "docker build --build-arg CONFIG_FILE=config.yml --build-arg PEM_FILE=key.pem --build-arg TERRAFORM_VERSION=${terraformVersion} --build-arg RANCHER2_PROVIDER_VERSION=${rancher2ProviderVersion} --build-arg LOCALS_PROVIDER_VERSION=${localProviderVersion} --build-arg AWS_PROVIDER_VERSION=${awsProviderVersion} -f Dockerfile -t ${imageName} . "
          }
          stage('Run Module Test') {
            def testPackage = env.TEST_PACKAGE?.trim()
            def testResultsDir = rootPath+"results"
            sh "mkdir -p ${testResultsDir}"
            try {
                if (testPackage?.toLowerCase().contains("sanity")) {
                  sh "docker run --name ${testContainer} -t -e CATTLE_TEST_CONFIG=${rootPath}config.yml -v ${homePath}key.pem:${rootPath}key.pem " +
                  "${imageName} sh -c \"/root/go/bin/gotestsum --format standard-verbose --packages=${standaloneTestsDir} --junitfile results/${testResultsOut} --jsonfile results/${testResultsJSON} -- -timeout=${timeout} -v ${params.TEST_CASE};" +
                  "${rootPath}pipeline/scripts/build_qase_reporter.sh;" +
                  "${rootPath}reporter\""
                } else {
                  sh "docker run --name ${testContainer} -t -e CATTLE_TEST_CONFIG=${rootPath}config.yml -v ${homePath}key.pem:${rootPath}key.pem " +
                  "${imageName} sh -c \"/root/go/bin/gotestsum --format standard-verbose --packages=${testsDir} --junitfile results/${testResultsOut} --jsonfile results/${testResultsJSON} -- -timeout=${timeout} -v ${params.TEST_CASE};" +
                  "${rootPath}pipeline/scripts/build_qase_reporter.sh;" +
                  "${rootPath}reporter\""
                }
            } catch(err) {
                echo 'Test run had failures. Collecting results...'
            }
          }
          stage('Test Report') {
            sh "docker cp ${testContainer}:${rootPath}results/${testResultsOut} ."
            step([$class: 'JUnitResultArchiver', testResults: "**/results/${testResultsOut}"])
            sh "docker stop ${testContainer}"
            sh "docker rm -v ${testContainer}"
            sh "docker rmi -f ${imageName}"
          }
        } catch(err) {        
            sh "docker stop ${testContainer}"
            sh "docker rm -v ${testContainer}"
            sh "docker rmi -f ${imageName}"
        }
  }
}