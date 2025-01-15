run:
	go fmt .
	go run .

perf-info:
	 go tool pprof -pdf  . cpuprofile > cpuprofile.pdf

android-install:
	fyne package -os android -appID dev.jkulzer.findinberlin
	adb install fib_client.apk
