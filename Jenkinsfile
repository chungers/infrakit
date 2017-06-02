#!groovy
properties(
  [
    buildDiscarder(logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '20')),
  ]
)

// Un-comment once Jenkins changes are in place
wrappedNode(label: "docker-edge && ubuntu && aufs") {
  deleteDir()
  checkout scm
  stage(name: "build JSON templates") {
    sh("make templates")
  }
  stage(name: "validate json files") {
    withTool("jq@1.5") {
      def cloudFormation = (ArrayList)(sh(script: 'find . -iname "*.json"', returnStdout: true).split("\r?\n"))
      for (i in cloudFormation) {
        try {
          sh("jq . '${i}' >/dev/null")
        } catch (Exception exc) {
          currentBuild.result = 'UNSTABLE'
          echo "jq failed"
          return
        }
      }
    }
  }
  stage(name: "build docker images") {
    try {
      sh("make dockerimages")
    } catch (Exception exc) {
      currentBuild.result = 'UNSTABLE'
      echo "docker image build failed"
      return
    }
  }

  stage(name: "trigger editions-release pipeline") {
    if (env.BRANCH_NAME == 'master') {
        build job: '../editions-release/master', wait: false
    }
  }
}
