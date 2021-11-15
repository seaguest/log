a very simple logger,supports log file rotation.


Usage:

```
	logger := log.New("/opt/log", log.INFO, 100, 10)
	log.SetLogger(logger)

```
