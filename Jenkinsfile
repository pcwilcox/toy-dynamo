def GIT_COMMIT_MESSAGE = 'NOPE'

pipeline {
    agent any
    environment {
        def TEST_SCRIPT = "test_HW1.py"
        def PORT_EXT = "5000" // Set this to the externally-visible port
        def PORT_INT = "8080" // This is specified by the program requirements
        def CONTAINER = "CS128-HW1"
        def TAG = "local:${CONTAINER}"
        def BUILD_FLAGS = "--force-rm --no-cache --tag ${TAG}"
        def RUN_FLAGS = "-p ${PORT_EXT}:${PORT_INT} -d --name ${CONTAINER} --rm ${TAG}"
    }
    stages {
        stage('Build') {
            steps {
                echo 'Building container...'
                sh "docker build ${BUILD_FLAGS} ."
                GIT_COMMIT_MESSAGE = sh(returnStdout: true, script: "git log --oneline --format=%B -n 1 ${GIT_COMMIT} | head -n 1").trim()
            }
        }
        stage('Run') {
            steps {
                echo 'Running container...'
                sh "docker run ${RUN_FLAGS}"
            }
        }
        stage('Test') {
            steps {
                echo 'Testing app...'
                withEnv(['PYTHONPATH=/usr/bin/python']) {
                    sh "python ${TEST_SCRIPT}"
                }
            }
        }
    }
    post {
        always {
            echo 'Cleaning up...'
            sh "docker stop ${CONTAINER}"
            echo 'Sending Discord notification'
            discordSend description: 'Jenkins Pipeline Build', footer: "${GIT_COMMIT_MESSAGE}", link: env.BUILD_URL, successful: currentBuild.resultIsBetterOrEqualTo('SUCCESS'), unstable: false, title: JOB_NAME, webhookURL: 'https://discordapp.com/api/webhooks/498390089228091412/4s3NOtQyGfdBq2BBr0d_keemA84Lt2zOKsSWcvQlpaTgyPZOmDRaTTQd-n4B2yfw3wZq'
        }
    }
}

