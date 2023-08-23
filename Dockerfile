FROM nokal/waiting-time-analysis:1.3.6

RUN apt-get update && apt-get install -y \
    curl \
    vim

WORKDIR /srv/webapp
ADD build/linux-amd64/ .
ADD run_analysis.bash .
ADD run_analysis_columns.bash .

EXPOSE 8080
CMD ["/srv/webapp/waiting-time-backend", "-host", "localhost", "-port", "8080"]
