mc alias set sedona http://localhost:9000 sedona sedona_password
mc mb sedona/apache-sedona-book-local

# Create a "directory" (actually just a prefix)
mc cp /dev/null sedona/apache-sedona-book-local/data