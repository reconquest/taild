# taild & tailcli

taild is daemon that runs `tail` processes in a unix way and sends response
through websocket connection.

tailcli is websocket client for taild that connects to taild and receives
output of tail command from taild service.

tailcli supports following arguments:

```
  -n, --lines=[+]NUM
         output the last NUM lines, instead of the last 10; or use -n +NUM to output starting with line NUM

  -f, --follow
         output appended data as the file grows;
```
