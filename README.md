# Yggdrasil map

  * Internet: https://yggdrasil-map.cwinfo.org
  * Hyperboria: https://yggdrasil-map.h.cwinfo.org
  * Yggdrasil Network: https://yggdrasil-map.y.cwinfo.org
  * I2P Network: http://yggnet.i2p

Forked from the code for http://www.fc00.org (http://h.fc00.org on Hyperboria).

## Sending your view of the network

This code reads a map of known nodes from `y.yakamo.org:3000/current` (reachable over yggdrasil). You may alternatively generate your own view of the network by running [a crawler script](scripts/crawl-dht.py), but this may take some time (figuring out how to run it and use the results is left as an exercise to the user).

## Web server
```bash
git clone https://github.com/Arceliar/yggdrasil-map.git
sudo apt-get install python-flask python-flup python-mysqldb python-pygraphviz python-networkx

cd yggdrasil-map/web
cp web_config.example.cfg web_config.cfg
python web.py
```

You would need to edit web.py to adjust the address/port the server listens on, and may want to edit the web_config.cfg file. Note that most of the options in web_config.cfg are unused after forking from the fc00.org code, so this is mostly just a workaround until we have time to clean up this code.

Run `web/updateGraph.py` periodically to rerender nodes graph. You may want to customize reverse-proxy IP retrieval logic in web.py.

## Web server with Docker
### 1. Copy files from contrib/Docker/ to your folder of choice.

### 2. Build image
Example folder is /docker/yggdrasil/map
```bash
sudo docker build -t Arceliar/yggdrasil-map /docker/yggdrasil/map/
```

### 3. Create container
Run with using Docker host network
```bash
sudo docker run --name yggdrasil-map -d --net host --restart always Arceliar/yggdrasil-map
```

You can also use other Docker networks
```bash
sudo docker run --name yggdrasil-map -d --net local --ip 10.254.1.3 --ip6 fd80:deaf:1::3 --mac-address 02:42:01:00:00:03 -e HOST=fd80:deaf:1::3 -e PORT=80 --restart always Arceliar/yggdrasil-map
```
