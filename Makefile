run:
	go fmt .
	go run .

perf-info:
	 go tool pprof -pdf  . cpuprofile > cpuprofile.pdf

package:
	fyne package -appID dev.jkulzer.findinberlin --target android/arm64

android-install: package
	adb install fib_client.apk

