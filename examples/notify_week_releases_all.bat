@echo off

REM This example checks retrieves week releases for all theaters and sends notifications if necessary.

REM Build binary

pushd ..\
echo Building...
go build -o bin\main.exe
echo Building... Finished.

REM Run application
echo Running week_releases job...
cd bin
main.exe -job week_releases -fcmauthkey YOUR_FCM_AUTH_KEY --alltheaters --notify > log_week_releases.txt

echo Job finished.

popd