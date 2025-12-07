FROM postgis/postgis:17-3.4

RUN mkdir /app
COPY postgis.sql /app
COPY postgis_ddl.sql /app
COPY migrate.sh /app

ENV POSTGRES_USER=postgis
ENV POSTGRES_PASSWORD=postgis
ENV POSTGRES_DB=postgis
ENV PGPASSWORD=postgis

COPY entrypoint.sh /app
RUN chmod +x /app/entrypoint.sh
RUN chmod +x /app/migrate.sh

ENTRYPOINT ["/app/entrypoint.sh"]
