wrappedNode(label: "docker") {
  deleteDir()
  checkout scm
  def cloudFormation = (ArrayList)(sh(script: 'find . -iname "*.json"', returnStdout: true).split("\r?\n"))
  parallel(
    'validate': { ->
      stage(name: "validate json files") {
        withTool("jq@1.5") {
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
    'sanitize': { ->
      stage(name: "sanitize json files") {
        for (i in cloudFormation) {
          try {
            sh("docker run --rm -v `pwd`:/data docker4x/sanity:latest '${i}'")
          } catch (Exception exc) {
            currentBuild.result = 'UNSTABLE'
            echo "sanity failed"
            return
          }
        }
      }
    }
  )
}
