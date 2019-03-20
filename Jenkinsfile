pipeline {
    agent { docker { image 'golang' } }
    stages {
        stage('build') {
            steps {
                sh 'go version'
                sh 'go build agent/kiteagent.go'
            }
        }
    }
}
