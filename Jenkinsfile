wrappedNode(label: "docker") {
  deleteDir()
  stage "validate"
  checkout scm
  withTool("jq@1.5") {
    def cloudFormation = (ArrayList)(sh(script: 'find . -iname "*.json"', returnStdout: true).split("\r?\n"))
    for (i in cloudFormation) {
      try {
        sh("jq . '${i}' >/dev/null")
        sh("docker run --rm -v `pwd`:/data docker4x/sanity:latest '${i}'")
      } catch (Exception exc) {
        currentBuild.result = 'UNSTABLE'
        echo "jq failed"
        return
      }
    }
  }
}