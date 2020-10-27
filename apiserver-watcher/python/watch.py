import time
import logging
from kubernetes import config

logging.basicConfig(
    format='%(asctime)s %(levelname)-8s %(message)s',
    level = logging.DEBUG)

config.load_incluster_config()

from kubernetes import client
v1 = client.CoreV1Api()

logging.info("Delay 10s for startup")
time.sleep(10)

while True:
    logging.info('Listing config maps')
    v1.list_namespaced_config_map('default')
    logging.info('OK')
    logging.info('Sleeping 10 minutes')
    time.sleep(600)