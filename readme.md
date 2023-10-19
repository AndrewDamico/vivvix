# VIVVIX Adspender Conversion Tool

## Introduction
The VIVVIX Adspender Conversion tool bulk processes VIVVIX reports downloaded from the web application. Specifically, it
* Captures the date and/or date range present in the report
* Renames the file according to the first date represented
* Creates a metadata file showing the first and last date in the report
* includes a tool which shows coverage of dates within a given period and identifies any files with overlapping dates

## Settings
When the application loads it will attempt to load previously saved directory information.

If directory information is not found:
1. navigate to the Configuration menu
2. Select option 2 'Set Import Directory'
3. Load your input directory

## Compiling 
To compile the application for windows:
1. Compile the resource file:
```
windres -o versioninfo.syso versioninfo.rc
```
2. Build the application
```
go build -o 'vivvix report converter.exe'
```