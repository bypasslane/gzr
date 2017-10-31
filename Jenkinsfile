node('default') {
  stage('Build CLI') {
    sh './ci/build.sh'
  }
  stage('Release CLI to Github') {
    withCredentials([[$class: 'StringBinding', credentialsId: 'c2c752a0-cbc7-44ed-9ba7-5467bc9cd2ec', variable: 'GITHUB_TOKEN']]) {
      sh './ci/release-github.sh'
    }
  }
  stage('Release Web to Dockerhub') {
    withCredentials([file(credentialsId: 'docker-hub-registry-config', variable: 'DOCKERCFG')]) {
      sh 'cp -u ${DOCKERCFG} /var/jenkins_home/.dockercfg'
      sh 'chmod 600 /var/jenkins_home/.dockercfg'
    }
    sh './ci/build-docker.sh'
  }
}
