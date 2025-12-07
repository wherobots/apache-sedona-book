## Cloud Native Geospatial Analytics With Apache Sedona

![sedona_book_cover.png](images/sedona_book_cover.png)

This repository contains the code examples and resources for the book **Cloud Native Geospatial Analytics With Apache Sedona** by Prashant Sharma, published by O'Reilly.

# Structure of the Repository
The repository has two main directories:
- book: Contains the code examples and resources for each chapter of the book.
- cli: Contains the command-line interface (CLI) tools and scripts used in the book.

# Getting Started
Each chapter of the books has it's own examples prepared in a separate directory under the `book` directory. 
You can easily run the examples using:
- docker-compose
- cli prepared for this book

# Prerequisites
To run the examples, you need to have the following installed on your machine:
- Docker
- Docker Compose

# Each chapter contains a README file, where we explain how to run the examples for that chapter.

# Running examples using the docker compose
![](run-compose-chapter-6.mov)
[run-compose-chapter-6.mov](run-compose-chapter-6.mov)

Note: Data prepared for this book is stored in the s3 buckets, based on your location and internet connection 
the download speed may vary. If you want to speed up the process you can download the data manually using the script and use
the minio bucket running in the docker compose.

Examples for a book are prepared mainly in book directory, where you can find 
- jupyter notebooks
- python scripts
- dbt project
- airflow dags and more

Book contains the following chapters 
- Chapter 1. Introduction to Apache Sedona
- [Chapter 2](book/chapter2). Getting Started with Apache Sedona
- [Chapter 3](book/chapter3). Loading Geospatial Data into Apache Sedona
- [Chapter 4](book/chapter4). Points, Lines, and Polygons: Vector Data Analysis with Spatial SQL
- [Chapter 5](book/chapter5). Raster Data Analysis
- [Chapter 6](book/chapter6). Apache Sedona and the PyData Ecosystem
- [Chapter 7](book/chapter7). Geospatial Data Science and Machine Learning
- [Chapter 8](book/chapter8). Building a Geospatial Data Lakehouse with Apache Parquet and Apache Iceberg
- [Chapter 9](book/chapter9). Using Apache Sedona with Cloud Data Providers
- [Chapter 10](book/chapter10). Optimizing Apache Sedona Applications

There is also chapter 11 where all the additional materials go which were not included in the main book. We will put more 
examples there in the future. 

We know that Apache Sedona is the processing engine which evolves quickly so we will also  to keep the repository updated
as soon as new features are released or changed. We can't do that simply in the book itself, so this repository will be 
the source of truth for the examples provided in the book and also good practises to use Apache Sedona.

Most of the examples you can open using the jupyter notebook endpoint by navigating to http://localhost:8888 in your browser.
When you are running the cli examples the jupyter notebook will be automatically opened in your default browser.

There couple of other examples which involves additional actions to make:
- running airflow dags for chapter 6 and in chapter 8 when we build geospatial lakehouse
- running dbt transformations in chapter 6 and chapter 8

For the Airflow go to http://localhost:8080 and login with username admin and password is visible when you start the 
docker compose in the terminal output. If you can't find it you can show it by running

```bash
./get-airflow-password.sh
```

To run dbt transformations you need to connect to docker container named `sedona-dbt` and run dbt commands there. You can connect to the container by running:

```bash
docker exec -it sedonadbt  /bin/bash
```