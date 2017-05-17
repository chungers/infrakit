wrappedNode(label: "docker") {
  deleteDir()
  checkout scm
  parallel(
    'templates': { ->
      stage(name: "build JSON templates") {
        sh("make templates")
      }
    },
    'validate': { ->
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
    },
    'build': { ->
      stage(name: "build docker images") {
        sh("make dockerimages")
      }
    },
    'release': { ->
      stage(name: "trigger editions-release pipeline") {
        if (env.BRANCH_NAME == 'master') {
            build job: '../editions-release/master', wait: false
        }
      }
    }
  )
}
