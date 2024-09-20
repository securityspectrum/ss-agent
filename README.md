Downloads and removes unused modules/packages/dependencies with:
```
go mod tidy
```

Configure the configuration file by renaming the `config-template.json` to
`config.json` and changing the values to your own.

Run the program from source code:
```
go run . 
```

To generate a binary:

For windows:
```
GOOS=windows GOARCH=amd64 go build -o ss-agent-win.exe .
```

For macOS:
```
GOOS=darwin GOARCH=amd64 go build -o ss-agent-macos-darwin . 
chmod +x ss-agent-darwin
./ss-agent-darwin -verbose
```

For Linux:
```
GOOS=linux GOARCH=amd64 go build -o ss-agent-linux . 
chmod +x ss-agent-linux
./ss-agent-linux -verbose
```