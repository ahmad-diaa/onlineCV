FROM jenkins/jenkins:lts
ENV JAVA_OPTS="-Xmx8192m"
USER root
RUN mkdir /var/log/jenkins
RUN chown -R  jenkins:jenkins /var/log/jenkins
USER jenkins
ENV JENKINS_OPTS="--logfile=/var/log/jenkins/jenkins.log"

