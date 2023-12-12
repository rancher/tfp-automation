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
            sh 'docker build --build-arg CONFIG_FILE=config.yml -f Dockerfile -t tfp-automation . '
    }
    
    stage('Run Module Test') {
            def dockerImage = docker.image('tfp-automation')
            dockerImage.inside() {
                sh "go test -v -timeout ${timeout} ${params.TEST_CASE} ${testsDir}"
            }
    }
}