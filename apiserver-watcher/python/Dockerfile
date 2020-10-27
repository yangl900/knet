FROM python:latest
RUN pip install kubernetes

WORKDIR /w
COPY . /w

CMD [ "python", "./watch.py"]