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
                sh "python ${TEST_SCRIPT}"
            }
        }
        stage('Cleanup') {
            steps {
                echo 'Cleaning up...'
                sh "docker stop ${CONTAINER}"
            }
        }
    }
}
