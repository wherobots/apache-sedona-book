FROM mysql:9.1.0

ENV MYSQL_ROOT_PASSWORD=root
ENV MYSQL_DATABASE=sedona
ENV MYSQL_USER=sedona
ENV MYSQL_PASSWORD=sedona

RUN mkdir /app

COPY mysql_ddl.sql /app
COPY migrate.sh /app
COPY entrypoint.sh /app
RUN chmod +x /app/entrypoint.sh
RUN chmod +x /app/migrate.sh

ENTRYPOINT ["/app/entrypoint.sh"]


