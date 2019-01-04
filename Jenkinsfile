pipeline {
    agent none

    triggers {
        upstream(upstreamProjects: 'open-simulation-platform/cse-core/bug/fix-archive-artifacts-after-conan-build', threshold: hudson.model.Result.SUCCESS)
    }

    options { checkoutToSubdirectory('src/cse-server-go') }

    stages {
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
                            projectName: 'open-simulation-platform/cse-core/bug/fix-archive-artifacts-after-conan-build',
                            filter: 'windows/debug/**/*',
                            target: 'src',
                            selector: "lastSuccessful"
                            )
                        
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
