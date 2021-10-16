pipeline {
    agent any

    stages {
        stage('prepare') {
            steps {
                echo 'preparing the application'
                dir('src/github.com/rsmaxwell/players-tt-api') {
                    checkout([
                        $class: 'GitSCM', 
                        branches: [[name: '*/main']], 
                        extensions: [], 
                        userRemoteConfigs: [[url: 'https://github.com/rsmaxwell/players-tt-api']]
                    ])
                }
            }
        }

        stage('build') {
            steps {
                echo 'building the application'
            }
        }

        stage('test') {
            steps {
                echo 'testing the application'
            }
        }

        stage('deploy') {
            steps {
                echo 'deploying the application'
            }
        }
    }
}
            
