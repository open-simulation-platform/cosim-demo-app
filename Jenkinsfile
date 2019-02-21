pipeline {

    agent { label 'windows' }

    environment {
        _JAVA_OPTIONS="-Duser.home=${env.BASE}\\lein-repositories\\${env.EXECUTOR_NUMBER}"
    }

    tools {
        jdk 'jdk8'
        //Leiningen no auto-install available, installing manually
    }

    stages {
        stage('Build client') {
            steps {
                sh 'curl -O https://raw.githubusercontent.com/technomancy/leiningen/stable/bin/lein'
                sh './lein clean'
                sh './lein cljsbuild once min'
            }
            post {
                success {
                    archiveArtifacts artifacts: 'resources/public/**/*',  fingerprint: true
                }
            }
        }
    }
}
