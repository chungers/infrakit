def cloudFormation = ["aws/cloudformation/docker_for_aws.json", "aws/cloudformation/docker_for_aws_ddc.json", "aws/cloudformation/docker_for_aws_cloud.json", "azure/editions.json", "azure/editions_ddc.json", "azure/editions_cloud.json"]

wrappedNode(label: "docker") {
  try {
    deleteDir()
    stage "validate"
    checkout scm
    withTool("jq@1.5") {
      for (i in cloudFormation) {
        try {
          sh "jq . '${i}' >/dev/null 2>&1"
        } catch (Exception exc) {
          currentBuild.result = 'UNSTABLE'
          echo "jq failed"
          return
        }
      }
    }
  }
}