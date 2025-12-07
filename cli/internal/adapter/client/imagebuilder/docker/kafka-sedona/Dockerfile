FROM apache/kafka:4.0.0

COPY entrypoint.sh entrypoint.sh
USER root
RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
