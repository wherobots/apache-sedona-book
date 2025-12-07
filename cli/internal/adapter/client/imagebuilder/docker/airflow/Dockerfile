FROM apache/airflow:2.10.2

COPY requirements.txt requirements.txt
RUN pip install -r requirements.txt

USER root

RUN apt-get update && apt-get install wget \
    && mkdir -p /etc/apt/keyrings \
    && wget -O - https://packages.adoptium.net/artifactory/api/gpg/key/public | tee /etc/apt/keyrings/adoptium.asc \
    && echo "deb [signed-by=/etc/apt/keyrings/adoptium.asc] https://packages.adoptium.net/artifactory/deb $(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) main" | tee /etc/apt/sources.list.d/adoptium.list \
    && apt update \
    && apt install -y temurin-11-jdk

RUN wget https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.262/aws-java-sdk-bundle-1.12.262.jar && \
    wget https://repo1.maven.org/maven2/org/apache/sedona/sedona-spark-shaded-3.5_2.12/1.7.2/sedona-spark-shaded-3.5_2.12-1.7.2.jar && \
    wget https://repo1.maven.org/maven2/org/datasyslab/geotools-wrapper/1.7.2-28.5/geotools-wrapper-1.7.2-28.5.jar

RUN wget https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.4/hadoop-aws-3.3.4.jar

RUN wget https://repo1.maven.org/maven2/org/apache/iceberg/iceberg-spark-runtime-3.5_2.12/1.9.2/iceberg-spark-runtime-3.5_2.12-1.9.2.jar
RUN wget https://repo1.maven.org/maven2/org/postgresql/postgresql/42.7.4/postgresql-42.7.4.jar

RUN cp aws-java-sdk-bundle-1.12.262.jar /home/airflow/.local/lib/python3.12/site-packages/pyspark/jars && \
    cp hadoop-aws-3.3.4.jar /home/airflow/.local/lib/python3.12/site-packages/pyspark/jars && \
    cp sedona-spark-shaded-3.5_2.12-1.7.2.jar /home/airflow/.local/lib/python3.12/site-packages/pyspark/jars && \
    cp geotools-wrapper-1.7.2-28.5.jar /home/airflow/.local/lib/python3.12/site-packages/pyspark/jars \
    && cp iceberg-spark-runtime-3.5_2.12-1.9.2.jar /home/airflow/.local/lib/python3.12/site-packages/pyspark/jars \
    && cp postgresql-42.7.4.jar /home/airflow/.local/lib/python3.12/site-packages/pyspark/jars

RUN mkdir /home/airflow/.local/lib/python3.12/site-packages/pyspark/conf
COPY spark-defaults.conf /home/airflow/.local/lib/python3.12/site-packages/pyspark/conf

RUN chmod 777 -R /home/airflow/.local/lib/python3.12/site-packages/pyspark/conf

COPY entrypoint.sh entrypoint.sh
RUN chmod +x entrypoint.sh
ENV AIRFLOW__LOGGING__LOGGING_LEVEL=WARNING
ENV PYTHONWARNINGS=ignore
USER airflow

ENTRYPOINT ["./entrypoint.sh"]
