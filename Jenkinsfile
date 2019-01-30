pipeline {
    agent none

    triggers {
        upstream(
            upstreamProjects: 'open-simulation-platform/cse-core/feature/151-build-conan-linux-binaries, open-simulation-platform/cse-server-go/master',
            threshold: hudson.model.Result.SUCCESS)
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
                        CGO_CFLAGS = "-I${WORKSPACE}/src/windows/debug/include"
                        CGO_LDFLAGS = "-L${WORKSPACE}/src/windows/debug/bin -lcsecorec"
                    }

                    tools {
                        go 'go-1.11'
                        //'com.cloudbees.jenkins.plugins.customtools.CustomTool' 'mingw-w64' awaiting fix in customToolsPlugin
                    }

                    steps {
                        sh 'echo Building on Windows'

                        sh 'go get github.com/gobuffalo/packr/packr'

                        copyArtifacts(
                            projectName: 'open-simulation-platform/cse-core/master',
                            filter: 'windows/debug/**/*',
                            target: 'src')
                        
                        copyArtifacts(
                            projectName: 'open-simulation-platform/cse-client/master',
                            filter: 'resources/public/**/*',
                            target: 'src/cse-server-go')
                        
                        dir ("${GOBIN}") {
                            sh 'curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'
                        }
                        dir ('src/cse-server-go') {
                            sh 'dep ensure'
                            sh 'packr build'
                        }
                    }
                    
                    post {
                        success {
                            archiveArtifacts artifacts: 'src/cse-server-go/cse-server-go.exe',  fingerprint: true
                        }
                    }
                }
                stage ('Build on Linux') {
                    agent {
                        dockerfile {
                            filename 'Dockerfile'
                            dir 'src/cse-server-go/.dockerfiles'
                            label 'linux && docker'
                            args '-v ${HOME}/jenkins_slave/conan-repositories/${EXECUTOR_NUMBER}:/conan_repo'
                        }
                    }

                    environment {
                        GOPATH = "${WORKSPACE}:${WORKSPACE}/src/cse-server-go/.deps/src"
                        CGO_CFLAGS = "-I${WORKSPACE}/src/linux/release/include"
                        CGO_LDFLAGS = "-L${WORKSPACE}/src/linux/release/lib -lcsecorec -lcsecorecpp -Wl,-rpath,$ORIGIN/../lib"
                        CONAN_USER_HOME = '/conan_repo'
                        CONAN_USER_HOME_SHORT = 'None'
                        OSP_CONAN_CREDS = credentials('jenkins-osp-conan-creds')
                    }
                    
                    steps {
                        sh 'echo building on Linux'

                        copyArtifacts(
                            projectName: 'open-simulation-platform/cse-core/master',
                            filter: 'linux/release/**/*',
                            target: 'src')
                        
                        copyArtifacts(
                            projectName: 'open-simulation-platform/cse-client/master',
                            filter: 'resources/public/**/*',
                            target: 'src/cse-server-go')
                        
                        dir ('src/cse-server-go') {
                            sh 'conan remote add osp https://osp-conan.azurewebsites.net/artifactory/api/conan/conan-local --force'
                            sh 'conan user -p $OSP_CONAN_CREDS_PSW -r osp $OSP_CONAN_CREDS_USR'
                            sh 'conan install . -s build_type=Release -s compiler.libcxx=libstdc++11'
                            sh 'go clean -cache'
                            sh 'dep ensure'
                            sh 'packr build'
                        }

                        dir ('src/cse-server-go/distribution/bin') {
                            sh 'cp -rf ../../cse-server-go .'
                        }
                        
                        dir ('src/cse-server-go/distribution/lib') {
                            sh "cp -rf ${WORKSPACE}/src/linux/release/lib/* ."
                        }

                        dir ('src/cse-server-go/distribution') {
                            zip (
                                zipFile: 'cse-server.zip',
                                archive: true
                            )
                        }
                    }

                    post {
                        success {
                            archiveArtifacts artifacts: 'src/cse-server-go/cse-server-go',  fingerprint: true
                        }
                    }
                }
            }
        }
    }
}
