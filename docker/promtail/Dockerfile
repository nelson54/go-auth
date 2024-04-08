FROM grafana/promtail:latest

RUN apt-get update && apt-get -y install logrotate

COPY ./docker/promtail/logrotate.conf /etc/logrotate.d/logrotate.conf
RUN chown root root /etc/logrotate.d/logrotate.conf

CMD service cron start && /usr/bin/promtail -config.file=/etc/promtail/config.yml




