# minimal_httpd
minimal http server in Go
launch it with ./minimal_httpd <PORT> <ROOT_PATH> <LOG_PATH>
for instance :
./minimal_httpd 8080 /tmp/ /tmp/logs_minimal_httpd.txt<br>
Features :
  - display .html files
  - download non .html files <br>
(only GET is available for the time being)
