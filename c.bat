@echo off
set FILES=main.go reader.go params.go stats.go defines.go requests.go formats.go

if [%1] == [build] (
	go build -o cdrsender.exe %FILES%
) else (
	start go run %FILES%
)

