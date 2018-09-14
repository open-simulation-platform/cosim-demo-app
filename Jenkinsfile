pipeline {
    agent none

    triggers {
        upstream(upstreamProjects: 'cse-core/PR-32', threshold: hudson.model.Result.SUCCESS)
    }

    options { checkoutToSubdirectory('src/cse-server-go') }

    stages {
        stage('Build client') {
            agent { label 'windows' }

            environment {
                _JAVA_OPTIONS="-Duser.home=${env.BASE}\\lein-repositories\\${env.EXECUTOR_NUMBER}"
            }

            tools {
                        jdk 'jdk8' 
                        //Leiningen no auto-install available, installing manually
                    }

            steps {
                dir('src/cse-server-go/cse-web-client-clojurescript') {
                    sh 'curl -O https://raw.githubusercontent.com/technomancy/leiningen/stable/bin/lein'
                    sh './lein cljsbuild once min'
                }
            }
        }
        stage('Build server') {
            parallel {
                stage('Build on Windows') {
                    agent { label 'windows' }

                    environment {
                        GOPATH = "${WORKSPACE}"
                        GOBIN = "${WORKSPACE}/bin"
                        PATH = "${env.MINGW_HOME}/bin;${GOBIN};${env.PATH}"
                        CGO_CFLAGS = "-I${WORKSPACE}/src/install/debug/include"
                        CGO_LDFLAGS = "-L${WORKSPACE}/src/install/debug/bin -lcsecorec"
                    }

                    tools {
                        go 'go-1.11' 
                        //'com.cloudbees.jenkins.plugins.customtools.CustomTool' 'mingw-w64' awaiting fix in customToolsPlugin
                    }

                    steps {
                        sh 'echo Building on Windows'

                        copyArtifacts(
                            projectName: 'cse-core/PR-32',
                            filter: 'install/debug/**/*',
                            target: 'src')
                        
                        dir ("${GOBIN}") {
                            sh 'curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'
                        }
                        dir ('src/cse-server-go') {
                            sh 'dep ensure'
                            sh 'go build'
                        }
                    }
                }
                stage ('Build on Linux') {
                    agent { label 'linux' }
                    
                    tools {
                        go 'go-1.11'
                    }

                    steps {
                        sh 'echo building on Linux'
                        sh 'go version'
                    }
                }
            }
        }
    }
}
