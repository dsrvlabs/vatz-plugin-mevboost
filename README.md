# vatz-plugin-mevboost
- Usage
```sh
root@lido-mainnet-node1new:~# mevboost-monitor -h
Usage of mevboost-monitor:
  -addr string
    	IP Address(e.g. 0.0.0.0, 127.0.0.1) (default "127.0.0.1")
  -mev-boost-version string
    	Please check your mev docker container name (default "mev-boost-v1.3.2")
  -port int
    	Port number, default 9095 (default 9095)
```
- mevboost_monitor_start.sh
```sh
root@lido-mainnet-node1new:~/bin/vatz# cat mevboost_monitor_start.sh
#!/bin/bash

PLUGIN=mevboost-monitor
LOG=/var/log/mevboost-monitor/mevboost-monitor.log
DOCKER_CONTAINER=mev-boost

$PLUGIN -mev-boost-version $DOCKER_CONTAINER >> $LOG 2>&1 &
echo "Good 💋"
```
- mevboots_monitor_stop.sh
```sh
root@lido-mainnet-node1new:~/bin/vatz# cat mevboost_monitor_stop.sh
#!/bin/bash

PLUGIN=mevboost-monitor

killall $PLUGIN &>/dev/null || true
echo "$PLUGIN is killed ☠️a"
```
- default.yaml
```yaml
root@lido-mainnet-node1new:~/bin/vatz# cat default.yaml
vatz_protocol_info:
  protocol_identifier: "lido-eth2"
  port: 19090
  health_checker_schedule:
    - "0 1 * * *"
  notification_info:
    host_name: "lido-mainnet-node1new"
    default_reminder_schedule:
      - "*/15 * * * *"
    dispatch_channels:
      - channel: "discord"
        secret: "https://discord.com/api/webhooks/864070380687982592/yE59bi44sRtscP7z4o9uMLttAbqWeqqeszWzkb3ubLcQrwr3C77vJnWrS25QmLFKzbA-"
      - channel: "telegram"
        secret: "Put Your Bot's Token"
        chat_id: "Put Your Chat's chat_id'"
        reminder_schedule:
          - "*/5 * * * *"
plugins_infos:
  default_verify_interval: 60
  default_execute_interval: 180
  default_plugin_name: "vatz-plugin"
  plugins:
    - plugin_name: "mev-monitor"
      plugin_address: "localhost"
      plugin_port: 9095
      executable_methods:
        - method_name: "mev-boost-liveness"
```
