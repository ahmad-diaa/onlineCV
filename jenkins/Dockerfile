FROM jenkins/jenkins:lts
ENV JAVA_OPTS="-Xmx8192m"
USER root
RUN apt-get update && apt-get install libltdl7
RUN groupadd -g 999 docker
RUN usermod -a -G docker jenkins
RUN mkdir /var/cache/jenkins
RUN chown -R jenkins:jenkins /var/cache/jenkins
RUN mkdir /var/log/jenkins
RUN chown -R  jenkins:jenkins /var/log/jenkins
USER jenkins
ENV JENKINS_OPTS="--logfile=/var/log/jenkins/jenkins.log --webroot=/var/cache/jenkins/war"

