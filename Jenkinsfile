pipeline {
    agent none

    triggers {
        upstream(
            upstreamProjects: 'open-simulation-platform/cse-core/feature/151-build-conan-linux-binaries, open-simulation-platform/cse-client/master',
            threshold: hudson.model.Result.SUCCESS)
    }

    options { checkoutToSubdirectory('src/cse-server-go') }

    stages {
        stage('Build server') {
            parallel {
                stage('Windows') {
                    agent { label 'windows' }

                    environment {
                        GOPATH = "${WORKSPACE}"
                        GOBIN = "${WORKSPACE}/bin"
                        PATH = "${env.MINGW_HOME}/bin;${GOBIN};${env.PATH}"
                        CGO_CFLAGS = "-I${WORKSPACE}/src/cse-server-go/include"
                        CGO_LDFLAGS = "-L${WORKSPACE}/src/cse-server-go/dist/bin -lcsecorec"
                        CONAN_USER_HOME = "${env.BASE}/conan-repositories/${env.EXECUTOR_NUMBER}"
                        CONAN_USER_HOME_SHORT = "${env.CONAN_USER_HOME}"
                        OSP_CONAN_CREDS = credentials('jenkins-osp-conan-creds')
                    }
                    
                    tools {
                        go 'go-1.11'
                        //'com.cloudbees.jenkins.plugins.customtools.CustomTool' 'mingw-w64' awaiting fix in customToolsPlugin
                    }

                    stages {
                        stage ('Get dependencies') {
                            steps {
                                dir ("${GOBIN}") {
                                    sh 'curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'
                                }
                                copyArtifacts(
                                    projectName: 'open-simulation-platform/cse-client/static-resources',
                                    filter: 'resources/public/**/*',
                                    target: 'src/cse-server-go')
                                
                                dir ('src/cse-server-go') {
                                    sh 'go get github.com/gobuffalo/packr/packr'
                                    sh 'conan remote add osp https://osp-conan.azurewebsites.net/artifactory/api/conan/conan-local --force'
                                    sh 'conan user -p $OSP_CONAN_CREDS_PSW -r osp $OSP_CONAN_CREDS_USR'
                                    sh 'conan install . -s build_type=Release'
                                    sh 'dep ensure'
                                }
                            }
                        }
                        stage ('Packr') {
                            steps {
                                dir ('src/cse-server-go') {
                                    sh 'go clean -cache'
                                    sh 'packr build -v'
                                }
                            }
                        }
                        stage ('Prepare dist') {
                            steps {
                                dir ('src/cse-server-go/dist/bin') {
                                    sh 'cp -rf ../../cse-server-go.exe .'
                                }
                            }
                        }
                        stage ('Zip dist') {
                            when {
                                not { buildingTag() }
                            }
                            steps {
                                dir ('src/cse-server-go/dist') {
                                    zip (
                                        zipFile: "cse-server-go-win64.zip",
                                        archive: true
                                    )
                                }
                            }
                        }
                        stage ('Zip release') {
                            when { buildingTag() }
                            steps {
                                dir ('src/cse-server-go/dist') {
                                    zip (
                                        zipFile: "cse-server-go-${env.TAG_NAME}-win64.zip",
                                        archive: true
                                    )
                                }
                            }
                        }
                    }
                    post {
                        cleanup {
                            dir('src/cse-server-go/dist') {
                                deleteDir();
                            }
                        }
                    }
                }
                stage ('Linux') {
                    agent {
                        dockerfile {
                            filename 'Dockerfile'
                            dir 'src/cse-server-go/.dockerfiles'
                            label 'linux && docker'
                            args '-v ${HOME}/jenkins_slave/conan-repositories/${EXECUTOR_NUMBER}:/conan_repo'
                        }
                    }

                    environment {
                        GOPATH = "${WORKSPACE}"
                        CGO_CFLAGS = "-I${WORKSPACE}/src/cse-server-go/include" 
                        CGO_LDFLAGS = "-L${WORKSPACE}/src/cse-server-go/dist/lib -lcsecorec -Wl,-rpath,\$ORIGIN/../lib"
                        CONAN_USER_HOME = '/conan_repo'
                        CONAN_USER_HOME_SHORT = 'None'
                        OSP_CONAN_CREDS = credentials('jenkins-osp-conan-creds')
                    }

                    stages {
                        stage ('Get dependencies') {
                            steps {
                                copyArtifacts(
                                    projectName: 'open-simulation-platform/cse-client/static-resources',
                                    filter: 'resources/public/**/*',
                                    target: 'src/cse-server-go')
                                
                                dir ('src/cse-server-go') {
                                    sh 'conan remote add osp https://osp-conan.azurewebsites.net/artifactory/api/conan/conan-local --force'
                                    sh 'conan user -p $OSP_CONAN_CREDS_PSW -r osp $OSP_CONAN_CREDS_USR'
                                    sh 'conan install . -s build_type=Release -s compiler.libcxx=libstdc++11'
                                    sh 'patchelf --set-rpath \'$ORIGIN/../lib\' dist/lib/*'
                                    sh 'dep ensure'
                                }
                            }
                        }
                        stage ('Packr') {
                            steps {
                                dir ('src/cse-server-go') {
                                    sh 'go clean -cache'
                                    sh 'packr build -v'
                                }
                            }
                        }
                        stage ('Prepare dist') {
                            steps {
                                dir ('src/cse-server-go/dist/bin') {
                                    sh 'cp -rf ../../cse-server-go .'
                                }
                            }
                        }
                        stage ('Zip dist') {
                            when {
                                not { buildingTag() }
                            }
                            steps {
                                dir ('src/cse-server-go/dist') {
                                    zip (
                                        zipFile: "cse-server-go-linux.tar.gz",
                                        archive: true
                                    )
                                }
                            }
                        }
                        stage ('Zip release') {
                            when { buildingTag() }
                            steps {
                                dir ('src/cse-server-go/dist') {
                                    zip (
                                        zipFile: "cse-server-go-${env.TAG_NAME}-linux.tar.gz",
                                        archive: true
                                    )
                                }
                            }
                        }
                    }
                    post {
                        cleanup {
                            dir('src/cse-server-go/dist') {
                                deleteDir();
                            }
                        }
                    }
                }
            }
        }
    }
}
