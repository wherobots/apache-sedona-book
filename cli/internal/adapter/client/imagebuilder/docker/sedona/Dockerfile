FROM apache/sedona:1.8.0

RUN wget https://repo1.maven.org/maven2/com/amazonaws/aws-java-sdk-bundle/1.12.262/aws-java-sdk-bundle-1.12.262.jar && \
    wget https://repo1.maven.org/maven2/org/apache/sedona/sedona-spark-shaded-3.5_2.12/1.7.2/sedona-spark-shaded-3.5_2.12-1.7.2.jar && \
    wget https://repo1.maven.org/maven2/org/datasyslab/geotools-wrapper/1.7.2-28.5/geotools-wrapper-1.7.2-28.5.jar


RUN wget https://repo1.maven.org/maven2/org/apache/hadoop/hadoop-aws/3.3.4/hadoop-aws-3.3.4.jar
RUN wget https://repo1.maven.org/maven2/org/postgresql/postgresql/42.7.4/postgresql-42.7.4.jar
RUN wget https://repo1.maven.org/maven2/com/mysql/mysql-connector-j/9.4.0/mysql-connector-j-9.4.0.jar

RUN rm /opt/spark/jars/sedona-spark-shaded-3.5_2.12-1.8.0.jar
COPY sedona-spark-shaded-3.5_2.12-1.8.1-SNAPSHOT.jar /opt/spark/jars

RUN cp aws-java-sdk-bundle-1.12.262.jar /opt/spark/jars && \
    cp hadoop-aws-3.3.4.jar /opt/spark/jars && \
    cp postgresql-42.7.4.jar /opt/spark/jars && \
    cp mysql-connector-j-9.4.0.jar /opt/spark/jars

RUN apt-get update && \
    apt-get install -y curl unzip && \
    rm -rf /var/lib/apt/lists/*

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install

RUN pip install rtree pystac

RUN wget https://repo1.maven.org/maven2/org/apache/iceberg/iceberg-spark-runtime-3.5_2.12/1.9.2/iceberg-spark-runtime-3.5_2.12-1.9.2.jar
RUN cp iceberg-spark-runtime-3.5_2.12-1.9.2.jar /opt/spark/jars

RUN pip install pyspark==3.5.2
RUN pip install contextily
RUN pip install folium
RUN pip install mapclassify
RUN pip install xgboost
RUN pip install "apache-sedona[db]"

COPY s3-conf.conf ./
COPY apply.sh ./

CMD ["/bin/bash", "./apply.sh"]
