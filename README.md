# CI UPDATE SERVER

This is the first version of CI Update server. You can customize it for your needs.
Also it is my first practice in Golang. 

Thanks. Best regards.

## Install manualy
1. Install go `https://go.dev/doc/install` than export `export GOPATH=~\go` (.profile or .bashrc)
2. Install app `go install`
* Copy files `scp ./* root@127.0.0.1:/root/my_projects/ci-update-server`
* Install `go install`
3. Set up service
* Copy service file to system `cp /root/my_projects/ci-update-server/scripts/ci-update-server.service /etc/systemd/system/`
* Reload deamon sudo `systemctl daemon-reload`
* Enable service `sudo systemctl enable ci-update-server.service`
* Start service `sudo systemctl start ci-update-server.service`
4. How to check
* See process `ps -e | grep 'github'`
* See Listener `netstat -plnt | grep ':27075'`
* Reboot and check `sudo reboot`
* See log `journalctl -n 10 -u ci-update-server.service` (show the last 10 lines)

## Useful commands
* Start service `sudo systemctl start ci-update-server.service`
* Stop service `sudo systemctl stop ci-update-server.service`
* Disable autorun service `sudo systemctl disable ci-update-server.service`

## Links
### GO Project layout examples
* https://github.com/golang-standards/project-layout
* https://github.com/prometheus/prometheus/blob/main/