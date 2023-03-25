# GoGet

## Default use

Out of the box, goget requires only a valid, complete, URL to work.

![default usage](./resources/goget_default.png)

This should start a download, log progress every second and save the file Free_Test_Data_1OMB_MP3.mp3 in your current directory.

![running](./resources/goget_running.png)

## Customization

You can modify the tool's behaviour with the following flags:

| Flag | Type | Default | Meaning
| :---:|:--:|:--:|:--|
| p | string | /   | where the file will be downloaded, defaults to your current directory |
| i | int    | 1   | the interval, in seconds, between each update log|
    

